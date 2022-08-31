package download

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"
)

type getFilesConfig struct {
	handler getURLsHandler
	url     string
}

// Download all the files (might take several minutes).
func Download(dir string, timeout time.Duration, urlsOnly, skip bool, parallel, retries int) error {
	c := &http.Client{Timeout: timeout}
	if !urlsOnly {
		log.Output(2, "Preparing to download from the Federal Revenue official website…")
	}
	confs := []getFilesConfig{
		{federalRevenueGetURLs, federalRevenueURL},
		{nationalTreasureGetURLs, nationalTreasureBaseURL},
	}
	fs, err := getFiles(c, confs, dir, skip)
	if err != nil {
		return fmt.Errorf("error gathering resources for download: %w", err)
	}
	if urlsOnly {
		urls := make([]string, 0, len(fs))
		for _, f := range fs {
			urls = append(urls, f.url)
		}
		sort.Strings(urls)
		for _, u := range urls {
			fmt.Println(u)
		}
		return nil
	}
	d, err := newDownloader(c, fs, parallel, retries)
	if err != nil {
		return fmt.Errorf("error creating a downloader: %w", err)
	}
	return d.downloadAll()
}
