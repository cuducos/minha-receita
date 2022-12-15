package cmd

import (
	"fmt"
	"os"

	"github.com/cuducos/minha-receita/api"
	"github.com/cuducos/minha-receita/db"
	"github.com/spf13/cobra"
)

const (
	defaultPort = "8000"
	apiHelper   = `
Starts the web API.

Using GODEBUG environment variable changes the HTTP server verbosity (for
example: http2debug=1 is verbose and http2debug=2 is more verbose, as in
https://golang.org/pkg/net/http/

The HTTP server is prepared to send logs to New Relic. If the
NEW_RELIC_LICENSE_KEY environment variable is set, the app automatically
reports to New Relic. Otherwise, the New Relic wrappers are just ignored.

The HTTP server is prepared to do a host header validation agains the value of
ALLOWED_HOST environment variable. If this variable is not set, this validation
is skipped.`
)

var (
	port     string
	newRelic string
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Spins up the web API",
	Long:  apiHelper,
	RunE: func(_ *cobra.Command, _ []string) error {
		u, err := loadDatabaseURI()
		if err != nil {
			return err
		}
		pg, err := db.NewPostgreSQL(u, postgresSchema)
		if err != nil {
			return err
		}
		defer pg.Close()
		if port == "" {
			port = os.Getenv("PORT")
		}
		if port == "" {
			port = defaultPort
		}
		if newRelic == "" {
			newRelic = os.Getenv("NEW_RELIC_LICENSE_KEY")
		}
		api.Serve(&pg, port, newRelic)
		return nil
	},
}

func apiCLI() *cobra.Command {
	apiCmd.Flags().StringVarP(
		&port,
		"port",
		"p",
		"",
		fmt.Sprintf("web server port (default PORT environment variable or %s)", defaultPort),
	)
	apiCmd.Flags().StringVarP(
		&newRelic,
		"new-relic-key",
		"n",
		"",
		"New Relic license key (deafult NEW_RELIC_LICENSE_KEY environment variable)",
	)
	return apiCmd
}
