package cmd

import (
	"github.com/spf13/cobra"

	"github.com/cuducos/minha-receita/check"
)

const checkHelper = `
Checks the integrity of the downloaded ZIP files.

The main files downloaded from the official website of the Brazilian
Federal Revenue are ZIP files. This command tries to unarchive them to check
their integrity.`

var deleteZipFiles bool

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Checks the integrity of downloaded ZIP files",
	Long:  checkHelper,
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := assertDirExists(); err != nil {
			return err
		}
		return check.Check(dir, deleteZipFiles)
	},
}

func checkCLI() *cobra.Command {
	checkCmd = addDataDir(checkCmd)
	checkCmd.Flags().BoolVarP(&deleteZipFiles, "delete", "x", deleteZipFiles, "deletes ZIP files that fails the check")
	return checkCmd
}
