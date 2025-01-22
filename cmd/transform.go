package cmd

import (
	"os"

	"github.com/mbnunes/minha-receita/db"
	"github.com/mbnunes/minha-receita/transform"
	"github.com/spf13/cobra"
)

const transformHelper = `
Convert the CSV files from the Federal Revenue for venues (ESTABELE group of
files) into records in the database, 1 record per CNPJ, joining information
from all other source CSV files.

The transformation process is divided into two steps:
1. Load relational data to a key-value store
2. Load the full database using the key-value store

If no specific step is specified, both steps will be executed by default, and
the key-value store is automaically deleted at the end.

If used with --step-one, the path to the key-value is printed to the stdout,
and it is NOT deleted at the end. This is the path expected as an argument
to --step-two.
`

var (
	maxParallelDBQueries int
	batchSize            int
	cleanUp              bool
	noPrivacy            bool
	stepOne              bool
	stepTwo              string
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

		db_type := os.Getenv("DATABASE_TYPE")

		if db_type == "mongo" {
			mdb, err := db.NewMongoDB()
			if err != nil {
				return err
			}

			if cleanUp {

				err = mdb.DropCollection(os.Getenv("COLLECTION"))
				if err != nil {
					return err
				}

				err = mdb.CreateCollection(os.Getenv("COLLECTION"))
				if err != nil {
					return err
				}

				err = mdb.CreateIndexes(os.Getenv("COLLECTION"))
				if err != nil {
					return err
				}
			}
			return err
			// return transform.Transform(dir, &pg, maxParallelDBQueries, batchSize, !noPrivacy, stepOne, stepTwo)
		} else {

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
			return transform.Transform(dir, &pg, maxParallelDBQueries, batchSize, !noPrivacy, stepOne, stepTwo)
		}

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
	transformCmd.Flags().BoolVarP(&stepOne, "step-one", "1", stepOne, "load relational data to a key-value store")
	transformCmd.Flags().StringVarP(&stepTwo, "step-two", "2", stepTwo, "path to the key-value store from step 1 to load the full database")
	return transformCmd
}
