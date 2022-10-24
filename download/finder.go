package download

import (
	"bufio"
	"bytes"
	"errors"
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
	size uint64
}

func newFile(url, dir string) file {
	return file{
		url:  url,
		path: filepath.Join(dir, url[strings.LastIndex(url, "/")+1:]),
	}
}

type getURLsHandler func(c *http.Client, url, dir string) ([]string, error)

func getURLs(client *http.Client, confs []getFilesConfig, dir string) ([]string, error) {
	var urls []string
	for _, c := range confs {
		u, err := c.handler(client, c.url, dir)
		if err != nil {
			return nil, fmt.Errorf("error getting urls: %w", err)
		}
		urls = append(urls, u...)
	}
	return urls, nil
}

func getFiles(client *http.Client, hs []getFilesConfig, dir string, skip bool) ([]file, error) {
	var fs []file
	urls, err := getURLs(client, hs, dir)
	if err != nil {
		return nil, fmt.Errorf("error getting files: %w", err)
	}
	for _, u := range urls {
		f := newFile(u, dir)
		h, err := os.Open(f.path)
		if !skip || errors.Is(err, os.ErrNotExist) {
			fs = append(fs, f)
			continue
		}
		if err == nil {
			h.Close()
		}
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
	type result struct {
		idx  int
		size uint64
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
				results <- result{idx, uint64(s)}
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
