package download

import (
	"bytes"
	"fmt"
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
	fs, err = getSizes(c, fs, false)
	if err != nil {
		return fmt.Errorf("error getting file sizes: %w", err)
	}
	d, err := newDownloader(c, fs, parallel, retries, false)
	if err != nil {
		return fmt.Errorf("error creating a downloader: %w", err)
	}
	return d.downloadAll()
}

// UrlList build a buffer with the contents matching the Google Cloud's Storage
// Transfer service URL list.
func UrlList() ([]byte, error) {
	c := &http.Client{}
	confs := []getFilesConfig{
		{federalRevenueGetURLs, federalRevenueURL},
		{nationalTreasureGetURLs, nationalTreasureBaseURL},
	}
	fs, err := getFiles(c, confs, os.TempDir(), false)
	if err != nil {
		return nil, fmt.Errorf("error gathering resources for download: %w", err)
	}
	fs, err = getSizes(c, fs, true)
	if err != nil {
		return nil, fmt.Errorf("error getting file sizes: %w", err)
	}
	var ls []string
	buf := bytes.NewBufferString("TsvHttpData-1.0\n")
	for _, f := range fs {
		ls = append(ls, fmt.Sprintf("%s\t%d", f.url, f.size))
	}
	sort.Strings(ls)
	buf.WriteString(strings.Join(ls, "\n"))
	return buf.Bytes(), nil
}
