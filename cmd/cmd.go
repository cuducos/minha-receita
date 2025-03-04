// Package cmd wraps the Cobra commands and sub-commands to build a CLI.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	defaultDataDir = "data"
	help           = `Minha Receita.

Toolbox to manage Minha Receita, including tools to handle extract, transform
and load data, manage the PostgreSQL instance, and to spin up the web server.

See --help for more details.
`
)

var dir string

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

var rootCmd = &cobra.Command{
	Use:   "minha-receita <command>",
	Short: "Minha Receita toolbox",
	Long:  help,
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates the required tables in the database",
	RunE: func(_ *cobra.Command, _ []string) error {
		db, err := loadDatabase()
		if err != nil {
			return fmt.Errorf("could not find database: %w", err)
		}
		defer db.Close()
		return db.Create()
	},
}

var dropCmd = &cobra.Command{
	Use:   "drop",
	Short: "Drops the tables in PostgreSQL",
	RunE: func(_ *cobra.Command, _ []string) error {
		db, err := loadDatabase()
		if err != nil {
			return fmt.Errorf("could not find database: %w", err)
		}
		defer db.Close()
		return db.Drop()
	},
}

var extraIndexesCmd = &cobra.Command{
	Use:   "extra-indexes",
	Short: "Creates extra indexes in the company field",
	RunE: func(_ *cobra.Command, args []string) error {
		db, err := loadDatabase()
		if err != nil {
			return fmt.Errorf("could not find database: %w", err)
		}
		defer db.Close()
		return db.ExtraIndexes(args)
	},
}

func addDataDir(c *cobra.Command) *cobra.Command {
	c.Flags().StringVarP(&dir, "directory", "d", defaultDataDir, "directory of the downloaded files")
	return c
}

func addDatabase(c *cobra.Command) *cobra.Command {
	c.Flags().StringVarP(&databaseURI, "database-uri", "u", "", "Database URI (default DATABASE_URL environment variable)")
	c.Flags().StringVarP(&postgresSchema, "postgres-schema", "s", "public", "PostgreSQL schema")
	return c
}

// CLI returns the root command from Cobra CLI tool.
func CLI() *cobra.Command {
	for _, c := range []*cobra.Command{createCmd, dropCmd, extraIndexesCmd} {
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
		extraIndexesCmd,
		transformCLI(),
		sampleCLI(),
		mirrorCLI(),
	} {
		rootCmd.AddCommand(c)
	}
	return rootCmd
}
