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
func Download(dir string, timeout time.Duration, skip bool, parallel, retries, chunkSize int) error {
	c := &http.Client{Timeout: timeout}
	log.Output(2, "Preparing to download from the Federal Revenue official website…")
	confs := []getFilesConfig{
		{federalRevenueGetURLs, federalRevenueURL},
		{nationalTreasureGetURLs, nationalTreasureBaseURL},
	}
	fs, err := getFiles(c, confs, dir, skip)
	if err != nil {
		return fmt.Errorf("error gathering resources for download: %w", err)
	}
	if len(fs) == 0 {
		return nil
	}
	fs, err = getSizes(c, fs, false)
	if err != nil {
		return fmt.Errorf("error getting file sizes: %w", err)
	}
	return download(c, fs, parallel, retries, uint64(chunkSize))
}

// URLs shows the URLs to be downloaded.
func URLs(db database, dir string, skip, tsv, saveToDB bool) error {
	c := &http.Client{}
	confs := []getFilesConfig{
		{federalRevenueGetURLsNoUpdatedAt, federalRevenueURL},
		{nationalTreasureGetURLs, nationalTreasureBaseURL},
	}
	fs, err := getFiles(c, confs, dir, skip)
	if err != nil {
		return fmt.Errorf("error gathering resources for download: %w", err)
	}
	if tsv {
		fs, err = getSizes(c, fs, true)
		if err != nil {
			return fmt.Errorf("error getting file sizes: %w", err)
		}
	}
	return listURLs(db, fs, tsv, saveToDB)
}
