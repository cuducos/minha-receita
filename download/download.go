package download

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type database interface{ MetaSave(string, string) error }

type getFilesConfig struct {
	handler getURLsHandler
	url     string
}

// Download all the files (might take several minutes).
func Download(
	db database,
	dir string,
	timeout time.Duration,
	urlsOnly, skip, tsv, saveToDB bool,
	parallel, retries int,
) error {
	c := &http.Client{Timeout: timeout}
	silent := urlsOnly
	if !silent {
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
	if !urlsOnly || tsv {
		fs, err = getSizes(c, fs, silent)
		if err != nil {
			return fmt.Errorf("error getting file sizes: %w", err)
		}
	}
	if urlsOnly {
		return listURLs(db, fs, tsv, saveToDB)
	}
	d, err := newDownloader(c, fs, uint(parallel), uint(retries), silent)
	if err != nil {
		return fmt.Errorf("error creating a downloader: %w", err)
	}
	return d.downloadAll()
}
