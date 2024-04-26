// Package cmd wraps the Cobra commands and sub-commands to build a CLI.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cuducos/minha-receita/db"
)

const (
	defaultDataDir = "data"
	help           = `Minha Receita.

Toolbox to manage Minha Receita, including tools to handle extract, transform
and load data, manage the PostgreSQL instance, and to spin up the web server.

See --help for more details.
`
)

var (
	dir            string
	databaseURI    string
	postgresSchema string
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
	u := os.Getenv("DATABASE_URL")
	if u == "" {
		return "", fmt.Errorf("could not find a database URI, pass it as a flag or set DATABASE_URL environment variable with the credentials for a PostgreSQL database")
	}
	return u, nil
}

var rootCmd = &cobra.Command{
	Use:   "minha-receita <command>",
	Short: "Minha Receita toolbox",
	Long:  help,
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates the required tables in PostgreSQL",
	RunE: func(_ *cobra.Command, _ []string) error {
		u, err := loadDatabaseURI()
		if err != nil {
			return err
		}
		pg, err := db.NewPostgreSQL(u, postgresSchema, nil)
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
		pg, err := db.NewPostgreSQL(u, postgresSchema, nil)
		if err != nil {
			return err
		}
		defer pg.Close()
		return pg.DropTable()
	},
}

func addDataDir(c *cobra.Command) *cobra.Command {
	c.Flags().StringVarP(&dir, "directory", "d", defaultDataDir, "directory of the downloaded files")
	return c
}

func addDatabase(c *cobra.Command) *cobra.Command {
	c.Flags().StringVarP(&databaseURI, "database-uri", "u", "", "PostgreSQL URI (default DATABASE_URL environment variable)")
	c.Flags().StringVarP(&postgresSchema, "postgres-schema", "s", "public", "PostgreSQL schema")
	return c
}

// CLI returns the root command from Cobra CLI tool.
func CLI() *cobra.Command {
	for _, c := range []*cobra.Command{createCmd, dropCmd} {
		addDatabase(c)
	}
	for _, c := range []*cobra.Command{
		apiCLI(),
		downloadCLI(),
		urlsCLI(),
		updatedAtCLI(),
		hasUpdateCLI(),
		checkCLI(),
		createCmd,
		dropCmd,
		transformCLI(),
		sampleCLI(),
		mirrorCLI(),
	} {
		rootCmd.AddCommand(c)
	}
	return rootCmd
}
