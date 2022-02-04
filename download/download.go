package download

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func createLastUpdateJSONFile(dir string, dt []string) error {
	// Create last_update.txt
	f, err := os.Create(filepath.Join(dir, "last_update.txt"))
	if err != nil {
		return err
	}
	defer f.Close()

	jsonS := make(map[string]string)
	jsonF := []string{"companies", "taxes"}
	for i := range jsonF {
		jsonS[jsonF[i]] = ""
		if i < len(dt) {
			jsonS[jsonF[i]] = dt[i]
		}
	}
	fBody, err := json.MarshalIndent(&jsonS, "", "    ")
	if err != nil {
		return err
	}
	_, err = f.Write(fBody)
	if err != nil {
		return err
	}
	return nil
}

// Download all the files (might take several minutes).
func Download(dir string, timeout time.Duration, urlsOnly, skip bool, parallel, retries int) error {
	c := &http.Client{Timeout: timeout}
	if !urlsOnly {
		log.Output(2, "Preparing to download from the Federal Revenue official website…")
	}

	doc, err := getHTMLDocument(c, federalRevenue)
	if err != nil {
		return err
	}

	fs, err := getFiles(doc, dir, skip)
	if err != nil {
		return err
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

	lastUpdateDates := getLastUpdate(doc)
	err = createLastUpdateJSONFile(dir, lastUpdateDates)
	if err != nil {
		return err
	}

	d, err := newDownloader(c, fs, 2, 4)
	if err != nil {
		return err
	}
	return d.downloadAll()
}
