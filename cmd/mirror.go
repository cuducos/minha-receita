package cmd

import (
	"fmt"
	"os"

	"github.com/cuducos/minha-receita/mirror"
	"github.com/spf13/cobra"
)

const mirrorHelper = `
Mirror of CNPJ files from the Federal Revenue.

Minha Receita maintains a mirror of data from the Federal Revenue CNPJ, in
addition to the executables. This is the wbe interface for the bucket of these
files.`

var mirrorCmd = &cobra.Command{
	Use:   "mirror",
	Long:  mirrorHelper,
	Short: "Starts the files mirror web interface.",
	RunE: func(_ *cobra.Command, _ []string) error {
		if port == "" {
			port = os.Getenv("PORT")
		}
		if port == "" {
			port = defaultPort
		}
		return mirror.Mirror(port)
	},
}

func mirrorCLI() *cobra.Command {
	mirrorCmd.Flags().StringVarP(
		&port,
		"port",
		"p",
		"",
		fmt.Sprintf("web server port (default PORT environment variable or %s)", defaultPort),
	)
	return mirrorCmd
}
