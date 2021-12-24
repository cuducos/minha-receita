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
files) into a a group of JSON files, 1 per CNPJ.`

const defaultPort = "8000"

var (
	outDir         string
	srcDir         string
	urlsOnly       bool
	timeout        string
	databaseURI    string
	postgresSchema string
	port           string
	newRelic       string
)

func assertDirExists(dir string) error {
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
	Short: "Minha Receita toolbox.",
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
		if err := assertDirExists(srcDir); err != nil {
			return err
		}
		dur, err := time.ParseDuration(timeout)
		if err != nil {
			return err
		}
		return download.Download(srcDir, dur, urlsOnly)
	},
}

var transformCmd = &cobra.Command{
	Use:   "transform",
	Short: "Transforms the CSV files into a group of JSON files, 1 per CNPJ",
	Long:  transformHelper,
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := assertDirExists(srcDir); err != nil {
			return err
		}
		if err := assertDirExists(outDir); err != nil {
			return err
		}
		return transform.Transform(srcDir, outDir)
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

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Imports the generated CSV into PostgreSQL",
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := assertDirExists(outDir); err != nil {
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
		return pg.ImportData(outDir)
	},
}

// CLI returns the root command from Cobra CLI tool.
func CLI() *cobra.Command {
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
	downloadCmd.Flags().BoolVarP(&urlsOnly, "urls-only", "u", false, "only list the URLs")
	downloadCmd.Flags().StringVarP(&timeout, "timeout", "t", "15m0s", "timeout for each download")
	for _, c := range []*cobra.Command{downloadCmd, transformCmd} {
		c.Flags().StringVarP(&srcDir, "source-directory", "s", "data", "directory of original CSV files")
	}
	for _, c := range []*cobra.Command{transformCmd, importCmd} {
		c.Flags().StringVarP(&srcDir, "output-directory", "o", "data", "directory of generated JSON & CSV files")
	}
	for _, c := range []*cobra.Command{createCmd, dropCmd, importCmd, apiCmd} {
		c.Flags().StringVarP(&databaseURI, "database-uri", "d", "", "PostgreSQL URI (default POSTGRES_URI environment variable)")
		c.Flags().StringVarP(&postgresSchema, "postgres-schema", "s", "public", "PostgreSQL schema")
	}
	for _, c := range []*cobra.Command{apiCmd, downloadCmd, transformCmd, createCmd, dropCmd, importCmd} {
		rootCmd.AddCommand(c)
	}
	return rootCmd
}
