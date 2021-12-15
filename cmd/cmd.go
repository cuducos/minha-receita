// Package cmd wraps the Cobra commands and sub-commands to build a CLI.
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/cuducos/minha-receita/api"
	"github.com/cuducos/minha-receita/db"
	"github.com/cuducos/minha-receita/download"
	"github.com/cuducos/minha-receita/transform"
)

const help = `Minha Receita.

Toolbox to manage Minha Receita, including tools to handle extract, transform
and load data, manage the PostgreSQL instance, and to spin up the web server.

Requires a POSTGRES_URI environment variable with PostgreSQL credentials.

An optional POSTGRES_SCHEMA environment variable can be user to set a different
schema than “public” (which is the default).

See --help for more details.
`

const apiHelper = `
Starts the web API.

The port used is 8000, unless an environment variable PORT points to a
different number.

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
files) into a a group of JSON files, 1 per CNPJ.`

var dir string
var urlsOnly bool
var timeout string

func assertDirExists() error {
	var err error
	i, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return fmt.Errorf("Directory %s does not exist.", dir)
	}
	if err != nil {
		return err
	}

	if !i.Mode().IsDir() {
		return fmt.Errorf("%s is not a directory.", dir)
	}

	return nil
}

var rootCmd = &cobra.Command{
	Use:   "minha-receita <command>",
	Short: "Minha Receita toolbox.",
	Long:  help,
}

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Spins up the web API",
	Long:  apiHelper,
	Run: func(_ *cobra.Command, _ []string) {
		pg := db.NewPostgreSQL()
		api.Serve(&pg)
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
		return download.Download(dir, dur, urlsOnly)
	},
}

var transformCmd = &cobra.Command{
	Use:   "transform",
	Short: "Transforms the CSV files into a group of JSON files, 1 per CNPJ",
	Long:  transformHelper,
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := assertDirExists(); err != nil {
			return err
		}
		return transform.Transform(dir)
	},
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates the required tables in PostgreSQL",
	Long:  "Creates the required tables in PostgreSQL, using the environment variable POSTGRES_URI to connect to the database.",
	Run: func(_ *cobra.Command, _ []string) {
		pg := db.NewPostgreSQL()
		pg.CreateTables()
	},
}

var dropCmd = &cobra.Command{
	Use:   "drop",
	Short: "Drops the tables in PostgreSQL",
	Long:  "Drops the tables in PostgreSQL, using the environment variable POSTGRES_URI to connect to the database.",
	Run: func(_ *cobra.Command, _ []string) {
		pg := db.NewPostgreSQL()
		pg.DropTables()
	},
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Imports the generated CSV and the Excel files into PostgreSQL",
	Long:  "Reads the compressed CSV and Excel files from a directory and copy their contents to the PostgreSQL tables, using the environment variable POSTGRES_URI to connect to the database.",
	Run: func(_ *cobra.Command, _ []string) {
		assertDirExists()
		pg := db.NewPostgreSQL()
		pg.ImportData(dir)
	},
}

// CLI returns the root command from Cobra CLI tool.
func CLI() *cobra.Command {
	downloadCmd.Flags().BoolVarP(&urlsOnly, "urls-only", "u", false, "only list the URLs")
	downloadCmd.Flags().StringVarP(&timeout, "timeout", "t", "15m0s", "timeout for each download")
	for _, c := range []*cobra.Command{downloadCmd, transformCmd, importCmd} {
		c.Flags().StringVarP(&dir, "directory", "d", "data", "data directory")
	}
	for _, c := range []*cobra.Command{apiCmd, downloadCmd, transformCmd, createCmd, dropCmd, importCmd} {
		rootCmd.AddCommand(c)
	}
	return rootCmd
}
