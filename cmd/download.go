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
Federal Revenue. An extra CSV file is downloaded from IBGE. Since the server
might be slow, all files are downloaded using multiple HTTP requests with
small content ranges.`

	urlsHelper = `
Shows the URLs of the required ZIP and CSV files.

The main files are downloaded from the official website of the Brazilian
Federal Revenue. An extra CSV file is downloaded from the National Treasure.`

	checkHelper = `
Checks the integrity of the downloaded ZIP files.

The main files downloaded from the official website of the Brazilian
Federal Revenue are ZIP files. This command tries to unarchive them to check
their integrity.`
)

var (
	timeout           string
	downloadRetries   uint
	parallelDownloads int
	chunkSize         int64
	skipExistingFiles bool
	restart           bool
	deleteZipFiles    bool
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

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Checks the integrity of downloaded ZIP files",
	Long:  checkHelper,
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := assertDirExists(); err != nil {
			return err
		}
		return download.Check(dir, deleteZipFiles)
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
	return downloadCmd
}

func urlsCLI() *cobra.Command {
	urlsCmd.Flags().StringVarP(&dir, "directory", "d", defaultDataDir, "directory of the downloaded files, used only with --skip")
	urlsCmd.Flags().BoolVarP(&skipExistingFiles, "skip", "x", false, "skip the download of existing files")
	return urlsCmd
}

func checkCLI() *cobra.Command {
	checkCmd = addDataDir(checkCmd)
	checkCmd.Flags().BoolVarP(&deleteZipFiles, "delete", "x", deleteZipFiles, "deletes ZIP files that fails the check")

	return checkCmd
}
