package download

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/schollz/progressbar/v3"
)

type file struct {
	url  string
	path string
	size int64
}

func newFile(url, dir string) file {
	return file{
		url:  url,
		path: filepath.Join(dir, url[strings.LastIndex(url, "/")+1:]),
	}
}

type getURLsHandler func(url, dir string) ([]string, error)

func getURLs(url string, handler getURLsHandler, dir string, skip bool) ([]string, error) {
	var out []string
	urls, err := handler(url, dir)
	if err != nil {
		return nil, fmt.Errorf("error getting urls: %w", err)
	}
	if skip {
		for _, u := range urls {
			h, err := os.Open(filepath.Join(dir, filepath.Base(u)))
			if os.IsNotExist(err) {
				out = append(out, u)
				continue
			}
			if err == nil {
				h.Close()
			}
		}
	} else {
		out = append(out, urls...)
	}
	return out, nil
}

func getFiles(url string, handler getURLsHandler, dir string, skip bool) ([]file, error) {
	var fs []file
	urls, err := getURLs(url, handler, dir, skip)
	if err != nil {
		return nil, fmt.Errorf("error getting files: %w", err)
	}
	for _, u := range urls {
		fs = append(fs, newFile(u, dir))
	}
	return fs, nil
}

func downloadAndGetSize(c *http.Client, url string) (int64, error) {
	r, err := c.Get(url)
	if err != nil {
		log.Output(1, fmt.Sprintf("HTTP request to %s failed: %v", url, err))
		return 0, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("http request to %s got %s", url, r.Status)
	}

	var buf bytes.Buffer
	s, err := io.Copy(bufio.NewWriter(&buf), r.Body)
	if err != nil {
		return 0, fmt.Errorf("could not get size for %s: %w", url, err)
	}
	return s, nil
}

func getSize(c *http.Client, url string) (int64, error) {
	r, err := c.Head(url)
	if err != nil {
		return 0, fmt.Errorf("error sending a http head request to %s: %s", url, err)
	}
	defer r.Body.Close()

	if r.ContentLength <= 0 {
		return downloadAndGetSize(c, url)
	}
	return r.ContentLength, nil
}

func getSizes(c *http.Client, fs []file, s bool) ([]file, error) {
	if len(fs) == 0 {
		return nil, nil
	}
	type result struct {
		idx  int
		size int64
	}
	var isShuttingDown bool
	var m sync.Mutex
	results := make(chan result)
	errors := make(chan error)
	for i, f := range fs {
		go func(u string, idx int, isShuttingDown *bool) {
			s, err := getSize(c, u)
			m.Lock()
			if !*isShuttingDown {
				if err != nil {
					*isShuttingDown = true
					errors <- err
					return
				}
				results <- result{idx, s}
			}
			m.Unlock()
		}(f.url, i, &isShuttingDown)
	}
	defer func() {
		close(errors)
		close(results)
	}()
	newBar := progressbar.Default
	if s {
		newBar = progressbar.DefaultSilent
	}
	bar := newBar(int64(len(fs)), "Gathering file sizes")
	defer bar.Close()
	for {
		select {
		case err := <-errors:
			return []file{}, fmt.Errorf("error getting total size: %w", err)
		case r := <-results:
			fs[r.idx].size = r.size
			bar.Add(1)
		}
		if bar.IsFinished() {
			break
		}
	}
	return fs, nil
}
