package download

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type file struct {
	url  string
	path string
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
		f := file{u, filepath.Join(dir, u[strings.LastIndex(u, "/")+1:])}
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
