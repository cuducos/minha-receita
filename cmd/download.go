package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cuducos/minha-receita/check"
	"github.com/cuducos/minha-receita/download"
	"github.com/spf13/cobra"
)

const downloadHelper = `
Downloads the required ZIP and Excel files.

The main files are downloaded from the official website of the Brazilian
Federal Revenue. An extra Excel file is downloaded from IBGE.`

var (
	urlsOnly          bool
	timeout           string
	downloadRetries   int
	parallelDownloads int
	skipExistingFiles bool
	insist            bool
)

func checkAndDownloadLoop(dir string, timeout time.Duration, parallel, retries int) {
	for {
		if err := check.Check(dir, true); err != nil {
			if !strings.Contains(err.Error(), "no zip files found") {
				log.Output(2, fmt.Sprintf("Error while checking for already downloaded files: %s", err))
			}
		}
		if err := download.Download(dir, timeout, false, true, parallel, retries); err != nil {
			log.Output(2, fmt.Sprintf("Error downloading files: %s", err))
			continue
		}
		break
	}
}

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Downloads the required ZIP and Excel files",
	Long:  downloadHelper,
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := assertDirExists(); err != nil {
			return err
		}
		dur, err := time.ParseDuration(timeout)
		if err != nil {
			return err
		}

		if insist {
			if !skipExistingFiles {
				log.Output(2, "The option --insist does not work without --skip. Activating --skip option.")
			}
			if urlsOnly {
				log.Output(2, "The option --insist does not work with --urls-only. Ignoring --urls-only option.")
			}
			checkAndDownloadLoop(dir, dur, parallelDownloads, downloadRetries)
			return nil
		}

		return download.Download(dir, dur, urlsOnly, skipExistingFiles, parallelDownloads, downloadRetries)
	},
}

func downloadCLI() *cobra.Command {
	downloadCmd = addDataDir(downloadCmd)
	downloadCmd.Flags().BoolVarP(&urlsOnly, "urls-only", "u", false, "only list the URLs")
	downloadCmd.Flags().BoolVarP(&skipExistingFiles, "skip", "x", false, "skip the download of existing files")
	downloadCmd.Flags().StringVarP(&timeout, "timeout", "t", "15m0s", "timeout for each download")
	downloadCmd.Flags().IntVarP(&downloadRetries, "retries", "r", download.MaxRetries, "maximum retries per file")
	downloadCmd.Flags().IntVarP(&parallelDownloads, "parallel", "p", download.MaxParallel, "maximum parallel downloads")
	downloadCmd.Flags().BoolVarP(
		&insist,
		"insist",
		"i",
		false,
		"restart if connection is broken before completing the downloads (automatically uses --skip and ignores --urls-only)",
	)
	return downloadCmd
}
