package cmd

import (
	"fmt"

	"github.com/cuducos/minha-receita/transformnext"
	"github.com/spf13/cobra"
)

var transformNextCmd = &cobra.Command{
	Use:   "transform-next",
	Short: "Experimental ETL, work in progress, NOT recommended",
	RunE: func(_ *cobra.Command, _ []string) error {
		db, err := loadDatabase()
		if err != nil {
			return fmt.Errorf("could not find database: %w", err)
		}
		defer db.Close()
		return transformnext.Transform(dir, db, batchSize, maxParallelDBQueries, !noPrivacy)
	},
}

var cleanupTempCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean-up temporary ETL files",
	RunE: func(_ *cobra.Command, _ []string) error {
		return transformnext.Cleanup()
	},
}

func transformNextCLI() *cobra.Command {
	transformNextCmd.Flags().IntVarP(
		&maxParallelDBQueries,
		"max-parallel-db-queries",
		"m",
		transformnext.MaxParallelDBQueries,
		"maximum parallel database queries",
	)
	transformNextCmd.Flags().IntVarP(&batchSize, "batch-size", "b", transformnext.BatchSize, "size of the batch to save to the database")
	transformNextCmd.Flags().BoolVarP(&noPrivacy, "no-privacy", "p", noPrivacy, "include email addresses, CPF and other PII in the JSON data")
	return transformNextCmd
}
