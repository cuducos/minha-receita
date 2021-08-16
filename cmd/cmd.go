// Package cmd wraps the Cobra commands and sub-commands to build a CLI.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cuducos/minha-receita/adapter"
	"github.com/cuducos/minha-receita/api"
	"github.com/cuducos/minha-receita/db"
	"github.com/cuducos/minha-receita/download"
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
Unzips the downloaded files and merge them into CSV files.

The Federal Revenue splits data from the same datasets in multiple files. This
command creates unique files for each dataset merging the data into single CSV
files per dataset.

Optionally, compression can be used. No compression is quicker but generates
large files. Gzip (gz) is slower, but generates considerably smaller files.
ZMA (xz) is even slower, and generates sligthly smaller files than Gzip.`

var compression string
var dir string
var urlsOnly bool

func assertDirExists() {
	var err error
	i, err := os.Stat(dir)
	if os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("Directory %s does not exist.", dir))
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if !i.Mode().IsDir() {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("%s is not a directory.", dir))
		os.Exit(1)
	}
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
	RunE: func(_ *cobra.Command, _ []string) error {
		pg := db.NewPostgreSQL()
		return api.Serve(&pg)
	},
}

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Downloads the required ZIP and Excel files",
	Long:  downloadHelper,
	RunE: func(_ *cobra.Command, _ []string) error {
		assertDirExists()
		return download.Download(dir, urlsOnly)
	},
}

var transformCmd = &cobra.Command{
	Use:   "transform",
	Short: "Unzips the downloaded files and merge them into CSV files",
	Long:  transformHelper,
	RunE: func(_ *cobra.Command, _ []string) error {
		assertDirExists()
		return adapter.Transform(dir, compression, false)
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
	transformCmd.Flags().StringVarP(
		&compression,
		"compression",
		"c",
		"",
		fmt.Sprintf(
			"optional compression algorithm (options available: %s)",
			adapter.CompressionAlgorithms,
		),
	)
	for _, c := range []*cobra.Command{downloadCmd, transformCmd, importCmd} {
		c.Flags().StringVarP(&dir, "directory", "d", "data", "data directory")
	}

	for _, c := range []*cobra.Command{
		apiCmd,
		downloadCmd,
		transformCmd,
		createCmd,
		dropCmd,
		importCmd,
	} {
		rootCmd.AddCommand(c)
	}

	return rootCmd
}
