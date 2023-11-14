package cmd

import (
	"github.com/cuducos/minha-receita/db"
	"github.com/cuducos/minha-receita/transform"
	"github.com/spf13/cobra"
)

const transformHelper = `
Convert the CSV files from the Federal Revenue for venues (ESTABELE group of
files) into records in the database, 1 record per CNPJ, joining information
from all other source CSV files.`

var (
	maxParallelDBQueries int
	batchSize            int
	cleanUp              bool
	noPrivacy            bool
	highMemory           bool
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
		return transform.Transform(dir, &pg, maxParallelDBQueries, batchSize, !noPrivacy, highMemory)
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
	transformCmd.Flags().BoolVarP(&highMemory, "high-memory", "x", highMemory, "high memory availability mode, faster but requires a lot of free RAM")
	return transformCmd
}
