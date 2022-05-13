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
)

var sampleCmd = &cobra.Command{
	Use:   "sample",
	Short: "Creates sample data of the source files from the Federal Revenue",
	Long:  sampleHelper,
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := assertDirExists(); err != nil {
			return err
		}
		return sample.Sample(dir, targetDir, maxLines)
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
	return sampleCmd
}
