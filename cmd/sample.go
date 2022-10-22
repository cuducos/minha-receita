package cmd

import (
	"path/filepath"

	"github.com/cuducos/minha-receita/sample"
	"github.com/spf13/cobra"
)

const sampleHelper = `
Creates versions of the source files from the Federal Revenue with a limited
number of lines, allowing us to manually test the process quicker.`

var (
	maxLines  int
	targetDir string
	updatedAt string
)

var sampleCmd = &cobra.Command{
	Use:   "sample",
	Short: "Creates sample data of the source files from the Federal Revenue",
	Long:  sampleHelper,
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := assertDirExists(); err != nil {
			return err
		}
		return sample.Sample(dir, targetDir, maxLines, updatedAt)
	},
}

func sampleCLI() *cobra.Command {
	sampleCmd = addDataDir(sampleCmd)
	sampleCmd.Flags().IntVarP(&maxLines, "max-lines", "m", sample.MaxLines, "maximum lines per file")
	sampleCmd.Flags().StringVarP(
		&targetDir,
		"target-directory",
		"t",
		filepath.Join(defaultDataDir, sample.TargetDir),
		"directory for the sample CSV files",
	)
	sampleCmd.Flags().StringVarP(
		&updatedAt,
		"updated-at",
		"u",
		"",
		"updated at date to be used if the data directory does not have a updated_at.txt file, format YYYY-MM-DD",
	)

	return sampleCmd
}
