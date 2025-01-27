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
is extremely slow, all files are downloaded using multiple HTTP requests with
small content ranges.`

	urlsHelper = `
Shows the URLs of the required ZIP and Excel files.

The main files are downloaded from the official website of the Brazilian
Federal Revenue. An extra Excel file is downloaded from IBGE.`

	hasUpdateHelper = `
Checks if there is an update available when it comes to the required ZIP from
the Federal Revenue.

Exists with exit code 0 if there is an update available and 1 otherwise.`
)

var (
	timeout           string
	downloadRetries   uint
	parallelDownloads int
	chunkSize         int64
	skipExistingFiles bool
	restart           bool
	useMirror         string
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
		if useMirror != "" {
			return download.DownloadFromMirror(useMirror, dir, dur, skipExistingFiles, restart, parallelDownloads, downloadRetries, chunkSize)
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

var updatedAtCmd = &cobra.Command{
	Use:   "updated-at",
	Short: "Shows the latest updated at date of the required ZIP from the Federal Revenue.",
	RunE: func(_ *cobra.Command, _ []string) error {
		return download.UpdatedAt()
	},
}

var hasUpdateCmd = &cobra.Command{
	Use:   "has-update",
	Short: "Checks if there is an update available when it comes to the required ZIP from the Federal Revenue.",
	Long:  hasUpdateHelper,
	RunE: func(_ *cobra.Command, _ []string) error {
		return download.HasUpdate(dir)
	},
}

func downloadCLI() *cobra.Command {
	downloadCmd = addDataDir(downloadCmd)
	downloadCmd.Flags().BoolVarP(&skipExistingFiles, "skip", "x", false, "skip the download of existing files")
	downloadCmd.Flags().StringVarP(&timeout, "timeout", "t", download.DefaultTimeout.String(), "timeout for each download")
	downloadCmd.Flags().UintVarP(&downloadRetries, "retries", "r", download.DefaultMaxRetries, "maximum retries per download")
	downloadCmd.Flags().IntVarP(&parallelDownloads, "parallel", "p", download.DefaultMaxParallel, "maximum parallel downloads")
	downloadCmd.Flags().Int64VarP(&chunkSize, "chunk-size", "c", download.DefaultChunkSize, "max length of the bytes range for each HTTP request")
	downloadCmd.Flags().BoolVarP(&restart, "restart", "e", false, "restart all downloads from the beginning")
	downloadCmd.Flags().StringVarP(&useMirror, "mirror", "m", "", "download from the mirror, not from the original source (YYYY-MM-DD)")
	return downloadCmd
}

func urlsCLI() *cobra.Command {
	urlsCmd.Flags().StringVarP(&dir, "directory", "d", defaultDataDir, "directory of the downloaded files, used only with --skip")
	urlsCmd.Flags().BoolVarP(&skipExistingFiles, "skip", "x", false, "skip the download of existing files")
	return urlsCmd
}

func updatedAtCLI() *cobra.Command {
	return updatedAtCmd
}

func hasUpdateCLI() *cobra.Command {
	hasUpdateCmd = addDataDir(hasUpdateCmd)
	return hasUpdateCmd
}
