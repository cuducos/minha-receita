package download

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
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
	cloudStorage string,
) error {
	if cloudStorage != "" {
		return transferJob(cloudStorage)
	}
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
	d, err := newDownloader(c, fs, parallel, retries, silent)
	if err != nil {
		return fmt.Errorf("error creating a downloader: %w", err)
	}
	return d.downloadAll()
}

func transferJob(b string) error {
	_, err := exec.LookPath("gcloud")
	if err != nil {
		return errors.New("gcloud client not installed or not in PATH")
	}
	cmd := exec.Command(
		"gcloud",
		"transfer",
		"jobs",
		"create",
		"https://minhareceita.org/urls",
		b,
	)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting gcloud: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("error executing gcloud: %w", err)
	}
	return nil
}
