// Package cmd wraps the Cobra commands and sub-commands to build a CLI.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/cuducos/minha-receita/api"
	"github.com/cuducos/minha-receita/db"
	"github.com/cuducos/minha-receita/download"
	"github.com/cuducos/minha-receita/sample"
	"github.com/cuducos/minha-receita/transform"
)

const help = `Minha Receita.

Toolbox to manage Minha Receita, including tools to handle extract, transform
and load data, manage the PostgreSQL instance, and to spin up the web server.

See --help for more details.
`

const apiHelper = `
Starts the web API.

Using GODEBUG environment variable changes the HTTP server verbosity (for
example: http2debug=1 is verbose and http2debug=2 is more verbose, as in
https://golang.org/pkg/net/http/

The HTTP server is prepared to send logs to New Relic. If the
NEW_RELIC_LICENSE_KEY environment variable is set, the app automatically
reports to New Relic. Otherwise, the New Relic wrappers are just ignored.`

const downloadHelper = `
Downloads the required ZIP and Excel files.

The main files are downloaded from the official website of the Brazilian
Federal Revenue. An extra Excel file is downloaded from IBGE.`

const transformHelper = `
Convert ths CSV files from the Federal Revenue for venues (ESTABELE group of
files) into records in the database, 1 record per CNPJ, joining information
from all other source CSV files.`

const sampleHelper = `
Creates versions of the source files from the Federal Revenue with a limited
number of lines, allowing us to manually test the process quicker.`

const defaultPort = "8000"

const defaultDataDir = "data"

var (
	dir            string
	databaseURI    string
	postgresSchema string

	// transform
	maxParallelDBQueries int
	batchSize            int
	cleanUp              bool
	noPrivacy            bool

	// download
	urlsOnly          bool
	timeout           string
	downloadRetries   int
	parallelDownloads int
	skipExistingFiles bool

	// api
	port     string
	newRelic string

	// sample
	maxLines  int
	targetDir string
)

func assertDirExists() error {
	i, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory %s does not exist", dir)
	}
	if err != nil {
		return err
	}
	if !i.Mode().IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}
	return nil
}

func loadDatabaseURI() (string, error) {
	if databaseURI != "" {
		return databaseURI, nil
	}
	u := os.Getenv("POSTGRES_URI")
	if u == "" {
		return "", fmt.Errorf("could not find a database URI, pass it as a flag or set POSTGRES_URI environment variable with the credentials for a PostgreSQL database")
	}
	return u, nil
}

var rootCmd = &cobra.Command{
	Use:   "minha-receita <command>",
	Short: "Minha Receita toolbox",
	Long:  help,
}

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Spins up the web API",
	Long:  apiHelper,
	RunE: func(_ *cobra.Command, _ []string) error {
		u, err := loadDatabaseURI()
		if err != nil {
			return err
		}
		pg, err := db.NewPostgreSQL(u, postgresSchema)
		if err != nil {
			return err
		}
		if port == "" {
			port = os.Getenv("PORT")
		}
		if port == "" {
			port = defaultPort
		}
		if newRelic == "" {
			newRelic = os.Getenv("NEW_RELIC_LICENSE_KEY")
		}
		api.Serve(&pg, port, newRelic)
		return nil
	},
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
		return download.Download(dir, dur, urlsOnly, skipExistingFiles, parallelDownloads, downloadRetries)
	},
}

var transformCmd = &cobra.Command{
	Use:   "transform",
	Short: "Transforms the CSV files into database records",
	Long:  transformHelper,
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := assertDirExists(); err != nil {
			return err
		}
		if err := db.AssertPostgresCLIExists(); err != nil {
			return err
		}

		u, err := loadDatabaseURI()
		if err != nil {
			return err
		}

		pg, err := db.NewPostgreSQL(u, postgresSchema)
		if err != nil {
			return err
		}
		defer pg.Close()

		if cleanUp {
			if err := pg.DropTable(); err != nil {
				return err
			}
			if err := pg.CreateTable(); err != nil {
				return err
			}
		}
		return transform.Transform(dir, &pg, maxParallelDBQueries, batchSize, !noPrivacy)
	},
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates the required tables in PostgreSQL",
	RunE: func(_ *cobra.Command, _ []string) error {
		u, err := loadDatabaseURI()
		if err != nil {
			return err
		}
		pg, err := db.NewPostgreSQL(u, postgresSchema)
		if err != nil {
			return err
		}
		defer pg.Close()
		return pg.CreateTable()
	},
}

var dropCmd = &cobra.Command{
	Use:   "drop",
	Short: "Drops the tables in PostgreSQL",
	RunE: func(_ *cobra.Command, _ []string) error {
		u, err := loadDatabaseURI()
		if err != nil {
			return err
		}
		pg, err := db.NewPostgreSQL(u, postgresSchema)
		if err != nil {
			return err
		}
		defer pg.Close()
		return pg.DropTable()
	},
}

var sampleCmd = &cobra.Command{
	Use:   "sample",
	Short: "Creates sample data of the source files from the Federal Revenue",
	Long:  sampleHelper,
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := assertDirExists(); err != nil {
			return err
		}
		return sample.Sample(dir, targetDir, maxLines)
	},
}

// CLI returns the root command from Cobra CLI tool.
func CLI() *cobra.Command {
	downloadCmd.Flags().BoolVarP(&urlsOnly, "urls-only", "u", false, "only list the URLs")
	downloadCmd.Flags().BoolVarP(&skipExistingFiles, "skip", "x", false, "skip the download of existing files")
	downloadCmd.Flags().StringVarP(&timeout, "timeout", "t", "15m0s", "timeout for each download")
	downloadCmd.Flags().IntVarP(&downloadRetries, "retries", "r", download.MaxRetries, "maximum retries per file")
	downloadCmd.Flags().IntVarP(&parallelDownloads, "parallel", "p", download.MaxParallel, "maximum parallel downloads")
	transformCmd.Flags().IntVarP(
		&maxParallelDBQueries,
		"max-parallel-db-queries",
		"m",
		transform.MaxParallelDBQueries,
		"maximum parallel database queries",
	)
	transformCmd.Flags().IntVarP(&batchSize, "batch-size", "b", transform.BatchSize, "size of the batch to save to the database")
	transformCmd.Flags().BoolVarP(&cleanUp, "clean-up", "c", cleanUp, "drop & recreate the database table before starting")
	transformCmd.Flags().BoolVarP(&noPrivacy, "no-privacy", "p", noPrivacy, "include email addresses, CPF and other PII in the JSON data")
	for _, c := range []*cobra.Command{downloadCmd, transformCmd, sampleCmd} {
		c.Flags().StringVarP(&dir, "directory", "d", defaultDataDir, "directory of the downloaded CSV files")
	}
	for _, c := range []*cobra.Command{transformCmd, createCmd, dropCmd, apiCmd} {
		c.Flags().StringVarP(&databaseURI, "database-uri", "u", "", "PostgreSQL URI (default POSTGRES_URI environment variable)")
		c.Flags().StringVarP(&postgresSchema, "postgres-schema", "s", "public", "PostgreSQL schema")
	}
	sampleCmd.Flags().IntVarP(&maxLines, "max-lines", "m", sample.MaxLines, "maximum lines per file")
	sampleCmd.Flags().StringVarP(
		&targetDir,
		"target-directory",
		"t",
		filepath.Join(defaultDataDir, sample.TargetDir),
		"directory for the sample CSV files",
	)
	for _, c := range []*cobra.Command{apiCmd, downloadCmd, transformCmd, createCmd, dropCmd, sampleCmd} {
		rootCmd.AddCommand(c)
	}
	apiCmd.Flags().StringVarP(
		&port,
		"port",
		"p",
		"",
		fmt.Sprintf("web server port (default PORT environment variable or %s)", defaultPort),
	)
	apiCmd.Flags().StringVarP(
		&newRelic,
		"new-relic-key",
		"n",
		"",
		"New Relic license key (deafult NEW_RELIC_LICENSE_KEY environment variable)",
	)
	return rootCmd
}
