package cmd

import (
	"github.com/cuducos/minha-receita/db"
	"github.com/cuducos/minha-receita/transform"
	"github.com/spf13/cobra"
)

const transformHelper = `
Convert the CSV files from the Federal Revenue for venues (ESTABELE group of
files) into records in the database, 1 record per CNPJ, joining information
from all other source CSV files.

The transformation process is divided into two steps:
1. Load relational data to a key-value store
2. Load the full database using the key-value store
`

var (
	maxParallelDBQueries int
	batchSize            int
	cleanUp              bool
	noPrivacy            bool
)

var transformCmd = &cobra.Command{
	Use:   "transform",
	Short: "Transforms the CSV files into database records",
	Long:  transformHelper,
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := assertDirExists(); err != nil {
			return err
		}
		u, err := loadDatabaseURI()
		if err != nil {
			return err
		}
		pg, err := db.NewPostgreSQL(u, postgresSchema, nil)
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

func transformCLI() *cobra.Command {
	transformCmd = addDataDir(transformCmd)
	transformCmd = addDatabase(transformCmd)
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
	return transformCmd
}
