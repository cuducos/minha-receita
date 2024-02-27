package download

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type getURLsHandler func(url, dir string) ([]string, error)

func getURLs(url string, handler getURLsHandler, dir string, skip bool) ([]string, error) {
	urls, err := handler(url, dir)
	if err != nil {
		return nil, fmt.Errorf("error getting urls: %w", err)
	}
	if !skip {
		return urls, nil
	}

	var out []string
	for _, u := range urls {
		p := filepath.Join(dir, filepath.Base(u))
		f, err := os.Open(p)
		if !skip || errors.Is(err, os.ErrNotExist) {
			out = append(out, u)
			continue
		}
		if err == nil {
			f.Close()
		}
	}
	return out, nil
}

// this server says it accepts HTTP range but it responds with the full file,
// so let's download it in a isolated step
func downloadNationalTreasure(dir string, skip bool) error {
	urls, err := getURLs(nationalTreasureBaseURL, nationalTreasureGetURLs, dir, skip)
	if err != nil {
		return fmt.Errorf("error gathering resources for national treasure download: %w", err)
	}
	if len(urls) == 0 {
		return nil
	}
	for _, u := range urls {
		if err := simpleDownload(u, dir); err != nil {
			return err
		}
	}
	return nil
}

// Download all the files (might take days).
func Download(dir string, timeout time.Duration, skip, restart bool, parallel int, retries uint, chunkSize int64) error {
	log.Output(1, "Downloading file(s) from the National Treasure…")
	if err := downloadNationalTreasure(dir, skip); err != nil {
		return fmt.Errorf("error downloading files from the national treasure: %w", err)
	}
	log.Output(1, "Downloading files from the Federal Revenue…")
	urls, err := getURLs(federalRevenueURL, federalRevenueGetURLs, dir, skip)
	if err != nil {
		return fmt.Errorf("error gathering resources for download: %w", err)
	}
	if len(urls) == 0 {
		return nil
	}
	if err := download(dir, urls, parallel, retries, chunkSize, timeout, restart); err != nil {
		return fmt.Errorf("error downloading files from the federal revenue: %w", err)
	}
	return nil
}

// Download all the files from the project's mirror
func DownloadFromMirror(mirror string, dir string, timeout time.Duration, skip, restart bool, parallel int, retries uint, chunkSize int64) error {
	urls, err := getMirrorURLs(mirror)
	if err != nil {
		return fmt.Errorf("error getting mirror urls: %w", err)
	}
	if err := download(dir, urls, parallel, retries, chunkSize, timeout, restart); err != nil {
		return fmt.Errorf("error downloading files from the mirror: %w", err)
	}
	return nil
}

// URLs shows the URLs to be downloaded.
func URLs(dir string, skip bool) error {
	urls := []string{federalRevenueURL, nationalTreasureBaseURL}
	handlers := []getURLsHandler{federalRevenueGetURLsNoUpdatedAt, nationalTreasureGetURLs}
	var out []string
	for idx := range urls {
		u, err := getURLs(urls[idx], handlers[idx], dir, skip)
		if err != nil {
			return fmt.Errorf("error gathering resources for download: %w", err)
		}
		out = append(out, u...)
	}
	sort.Strings(out)
	fmt.Println(strings.Join(out, "\n"))
	return nil
}

// UpdatedAt shows the updated at of the files to be downloaded.
func UpdatedAt() error {
	u, err := fetchUpdatedAt(federalRevenueURL)
	if err != nil {
		return fmt.Errorf("error getting updated at: %w", err)
	}
	fmt.Println(u)
	return nil
}

// HasUpdate checks if there is an update available.
func HasUpdate(dir string) error {
	h, err := hasUpdate(federalRevenueURL, dir)
	if err != nil {
		return fmt.Errorf("error getting updated at: %w", err)
	}
	if !h {
		os.Exit(1)
	}
	return nil
}
