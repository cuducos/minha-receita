package cmd

import (
	"github.com/cuducos/minha-receita/transformnext"
	"github.com/spf13/cobra"
)

var transformNextCmd = &cobra.Command{
	Use:   "transform-next",
	Short: "Experimental ETL, work in progress, NOT recommended",
	RunE: func(_ *cobra.Command, _ []string) error {
		return transformnext.Transform(dir)
	},
}

var cleanupTempCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean-up temporary ETL files",
	RunE: func(_ *cobra.Command, _ []string) error {
		return transformnext.Cleanup()
	},
}
