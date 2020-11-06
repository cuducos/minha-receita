package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cuducos/minha-receita/api"
	"github.com/cuducos/minha-receita/db"
	"github.com/cuducos/minha-receita/download"
)

const help = `Minha Receita.

Toolbox to manage Minha Receita, including tools to handle data, database and
to spin up the web server.

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
Downloads the required compressed CSV and Excel files.

The main files are downloaded from the official website of the Brazilian
Federal Revenue, or from Brasil.IO mirror. An extra file is downloaded from
IBGE.`

const parseHelper = `
Parse the fixed-width files from the Federal Revenue.

Three compressed CSVs are created: empresa.csv.gz, socio.csv.gz and
cnae_secundarias.csv.gz.`

var dir string
var mirror bool

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
	Run: func(_ *cobra.Command, _ []string) {
		pg := db.NewPostgreSQL()
		api.Serve(&pg)
	},
}

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Downloads the required compressed CSV and Excel files",
	Long:  downloadHelper,
	Run: func(_ *cobra.Command, _ []string) {
		assertDirExists()
		download.Download(mirror, dir)
	},
}

var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "Parse the fixed-width files from the Federal Revenue",
	Long:  parseHelper,
	Run: func(_ *cobra.Command, _ []string) {
		assertDirExists()
		download.Parse(dir)
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
	Short: "Imports data into PostgreSQL",
	Long:  "Reads the compressed CSV and Excel files from a directory and copy their contents to the PostgreSQL tables, using the environment variable POSTGRES_URI to connect to the database.",
	Run: func(_ *cobra.Command, _ []string) {
		assertDirExists()
		pg := db.NewPostgreSQL()
		pg.ImportData(dir)
	},
}

func main() {
	downloadCmd.Flags().BoolVarP(&mirror, "mirror", "m", false, "use Brasil.IO mirror")
	downloadCmd.Flags().StringVarP(&dir, "directory", "d", "data", "data directory")
	parseCmd.Flags().StringVarP(&dir, "directory", "d", "data", "data directory")
	importCmd.Flags().StringVarP(&dir, "directory", "d", "data", "data directory")

	rootCmd.AddCommand(apiCmd)
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(parseCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(dropCmd)
	rootCmd.AddCommand(importCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
