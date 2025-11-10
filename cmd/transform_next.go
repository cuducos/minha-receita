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
		if err := assertDirExists(); err != nil {
			return err
		}
		db, err := loadDatabase()
		if err != nil {
			return fmt.Errorf("could not find database: %w", err)
		}
		defer db.Close()
		if cleanUp {
			err = db.Drop()
			if err != nil {
				return err
			}
			err = db.Create()
			if err != nil {
				return err
			}
		}
		return transformnext.Transform(db)
	},
}

func transformNextCLI() *cobra.Command {
	transformCmd = addDataDir(transformNextCmd)
	transformCmd = addDatabase(transformNextCmd)
	return transformCmd
}
