package download

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type getFilesConfig struct {
	handler getURLsHandler
	url     string
}

func simpleDownload(f file) error {
	h, err := os.Create(f.path)
	if err != nil {
		return fmt.Errorf("could not create %s: %w", f.path, err)
	}
	defer h.Close()
	resp, err := http.Get(f.url)
	if err != nil {
		return fmt.Errorf("error requesting %s: %w", f.url, err)

	}
	defer resp.Body.Close()
	_, err = io.Copy(h, resp.Body)
	if err != nil {
		return fmt.Errorf("error writing to %s: %w", f.path, err)
	}
	return nil
}

// this server says it accepts HTTP range but it responds with the full file,
// so let's download it in a isolated step
func downloadNationalTreasure(dir string, skip bool) error {
	fs, err := getFiles(getFilesConfig{nationalTreasureGetURLs, nationalTreasureBaseURL}, dir, skip)
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
	fs, err := getFiles(getFilesConfig{federalRevenueGetURLs, federalRevenueURL}, dir, skip)
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
	confs := []getFilesConfig{
		{federalRevenueGetURLsNoUpdatedAt, federalRevenueURL},
		{nationalTreasureGetURLs, nationalTreasureBaseURL},
	}
	var out []string
	for _, conf := range confs {
		u, err := getURLs(conf, dir)
		if err != nil {
			return fmt.Errorf("error gathering resources for download: %w", err)
		}
		out = append(out, u...)
	}
	sort.Strings(out)
	fmt.Println(strings.Join(out, "\n"))
	return nil
}
