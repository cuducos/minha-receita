package cmd

import (
	"time"

	"github.com/cuducos/minha-receita/download"
	"github.com/spf13/cobra"
)

const (
	downloadHelper = `
Downloads the required ZIP and Excel files.

The main files are downloaded from the official website of the Brazilian
Federal Revenue. An extra Excel file is downloaded from IBGE. Since the server
is extremelly slow, all files are downloaded using multiple HTTP requests with
small content ranges.`

	urlsHelper = `
Shows the URLs of the requires ZIP and Excel files.

The main files are downloaded from the official website of the Brazilian
Federal Revenue. An extra Excel file is downloaded from IBGE.`
)

var (
	timeout           string
	downloadRetries   int
	parallelDownloads int
	chunkSize         int
	skipExistingFiles bool
	restart           bool
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
		return download.Download(dir, dur, skipExistingFiles, restart, parallelDownloads, downloadRetries, chunkSize)
	},
}

var urlsCmd = &cobra.Command{
	Use:   "urls",
	Short: "Shows the URLs for the required ZIP and Excel files",
	Long:  urlsHelper,
	RunE: func(_ *cobra.Command, _ []string) error {
		if skipExistingFiles {
			if err := assertDirExists(); err != nil {
				return err
			}
		}
		return download.URLs(dir, skipExistingFiles)
	},
}

func downloadCLI() *cobra.Command {
	downloadCmd = addDataDir(downloadCmd)
	downloadCmd.Flags().BoolVarP(&skipExistingFiles, "skip", "x", false, "skip the download of existing files")
	downloadCmd.Flags().StringVarP(&timeout, "timeout", "t", download.DefaultTimeout.String(), "timeout for each download")
	downloadCmd.Flags().IntVarP(&downloadRetries, "retries", "r", download.DefaultMaxRetries, "maximum retries per download, use -1 for unlimited")
	downloadCmd.Flags().IntVarP(&parallelDownloads, "parallel", "p", download.DefaultMaxParallel, "maximum parallel downloads")
	downloadCmd.Flags().IntVarP(&chunkSize, "chunk-size", "c", download.DefaultChunkSize, "max length of the bytes range for each HTTP request")
	downloadCmd.Flags().BoolVarP(&restart, "restart", "e", false, "restart all downloads from the beginning")
	return downloadCmd
}

func urlsCLI() *cobra.Command {
	urlsCmd.Flags().StringVarP(&dir, "directory", "d", defaultDataDir, "directory of the downloaded files, used only with --skip")
	urlsCmd.Flags().BoolVarP(&skipExistingFiles, "skip", "x", false, "skip the download of existing files")
	return urlsCmd
}
