package cmd

import (
	"github.com/spf13/cobra"

	"github.com/cuducos/minha-receita/check"
)

var (
	checksumSrcDir string
	deleteZipFiles bool
)

const checksumHelper = `
Checksum of the downloaded files.

Even though the official website of the Brazilian Federal Revenue does not offer
a checksum for their files, this command can be used to create or check the checksum
of downloaded files.`

var rootChecksumCmd = &cobra.Command{
	Use:   "checksum",
	Short: "Checksum of the downloaded files.",
	Long:  checksumHelper,
}

var createChecksumCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates checksum of downloaded files.",
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := assertDirExists(); err != nil {
			return err
		}
		return check.CreateChecksum(dir)
	},
}

var checkChecksumCmd = &cobra.Command{
	Use:   "check",
	Short: "Checks checksum of downloaded files.",
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := assertDirExists(); err != nil {
			return err
		}
		return check.CheckChecksum(checksumSrcDir, dir)
	},
}

const checkHelper = `
Checks the integrity of the downloaded ZIP files.

The main files downloaded from the official website of the Brazilian
Federal Revenue are ZIP files. This command tries to unarchive them to check
their integrity.`

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
	for _, cmd := range []*cobra.Command{createChecksumCmd, checkChecksumCmd} {
		cmd = addDataDir(cmd)

		rootChecksumCmd.AddCommand(cmd)
	}

	checkChecksumCmd.Flags().StringVarP(
		&checksumSrcDir,
		"src-directory",
		"s",
		defaultDataDir,
		"directory of the checksum file(s) to compare with",
	)

	checkCmd = addDataDir(checkCmd)
	checkCmd.Flags().BoolVarP(&deleteZipFiles, "delete", "x", deleteZipFiles, "deletes ZIP files that fails the check")
	checkCmd.AddCommand(rootChecksumCmd)

	return checkCmd
}
