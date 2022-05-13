package cmd

import (
	"time"

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
)

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
	return downloadCmd
}
