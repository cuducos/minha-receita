package download

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

// this server says it accepts HTTP range but it responds with the full file,
// so let's download it in a isolated step
func downloadNationalTreasure(dir string, skip bool) error {
	fs, err := getFiles(nationalTreasureBaseURL, nationalTreasureGetURLs, dir, skip)
	if err != nil {
		return fmt.Errorf("error gathering resources for national treasure download: %w", err)
	}
	if len(fs) == 0 {
		return nil
	}
	for _, f := range fs {
		if err := simpleDownload(f); err != nil {
			return err
		}
	}
	return nil
}

// Download all the files (might take several minutes).
func Download(dir string, timeout time.Duration, skip, restart bool, parallel, retries, chunkSize int) error {
	r, err := newRecover(dir, chunkSize, restart)
	if err != nil {
		return fmt.Errorf("error creating a download recover struct: %w", err)
	}
	defer r.close()
	c := &http.Client{Timeout: timeout}
	log.Output(1, "Downloading files from the National Treasure…")
	if err := downloadNationalTreasure(dir, skip); err != nil {
		return err
	}
	log.Output(1, "Preparing to download from the Federal Revenue official website…")
	fs, err := getFiles(federalRevenueURL, federalRevenueGetURLs, dir, skip)
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
	return download(c, fs, r, parallel, retries, chunkSize)
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
