package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cuducos/minha-receita/check"
	"github.com/cuducos/minha-receita/db"
	"github.com/cuducos/minha-receita/download"
	"github.com/spf13/cobra"
)

const (
	downloadHelper = `
Downloads the required ZIP and Excel files.

The main files are downloaded from the official website of the Brazilian
Federal Revenue. An extra Excel file is downloaded from IBGE.`

	urlsHelper = `
Shows the URLs of the requires ZIP and Excel files.

The main files are downloaded from the official website of the Brazilian
Federal Revenue. An extra Excel file is downloaded from IBGE.`
)

var (
	timeout           string
	downloadRetries   int
	parallelDownloads int
	skipExistingFiles bool
	insist            bool
	tsv               bool
	saveToDB          bool
)

func checkAndDownloadLoop(dir string, timeout time.Duration, parallel, retries int) {
	for {
		if err := check.Check(dir, true); err != nil {
			if !strings.Contains(err.Error(), "no zip files found") {
				log.Output(2, fmt.Sprintf("Error while checking for already downloaded files: %s", err))
			}
		}
		if err := download.Download(dir, timeout, true, parallel, retries); err != nil {
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
			checkAndDownloadLoop(dir, dur, parallelDownloads, downloadRetries)
			return nil
		}
		return download.Download(dir, dur, skipExistingFiles, parallelDownloads, downloadRetries)
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
		var pg db.PostgreSQL
		if saveToDB {
			u, err := loadDatabaseURI()
			if err != nil {
				return err
			}
			pg, err = db.NewPostgreSQL(u, postgresSchema)
			if err != nil {
				return err
			}
		}
		return download.URLs(&pg, dir, skipExistingFiles, tsv, saveToDB)
	},
}

func downloadCLI() *cobra.Command {
	downloadCmd = addDataDir(downloadCmd)
	downloadCmd.Flags().BoolVarP(&skipExistingFiles, "skip", "x", false, "skip the download of existing files")
	downloadCmd.Flags().StringVarP(&timeout, "timeout", "w", "15m0s", "timeout for each download")
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

func urlsCLI() *cobra.Command {
	urlsCmd.Flags().StringVarP(&dir, "directory", "d", defaultDataDir, "directory of the downloaded files, used only with --skip")
	urlsCmd.Flags().BoolVarP(&tsv, "tsv", "t", false, "use TSV when listing URLs")
	urlsCmd.Flags().BoolVarP(&saveToDB, "save-to-db", "s", false, "save URL list to DATABASE_URL when listing URLs")
	urlsCmd.Flags().BoolVarP(&skipExistingFiles, "skip", "x", false, "skip the download of existing files")
	return urlsCmd
}
