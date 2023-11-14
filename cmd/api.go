package cmd

import (
	"fmt"
	"os"

	"github.com/cuducos/minha-receita/api"
	"github.com/cuducos/minha-receita/db"
	"github.com/cuducos/minha-receita/monitor"
	"github.com/newrelic/go-agent/v3/newrelic"
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

The HTTP server is prepared to do a host header validation against the value of
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
		if newRelic == "" {
			newRelic = os.Getenv("NEW_RELIC_LICENSE_KEY")
		}
		var nr *newrelic.Application
		if newRelic != "" {
			nr, err = monitor.NewRelicApp(newRelic)
			if err != nil {
				return err
			}
		}
		pg, err := db.NewPostgreSQL(u, postgresSchema, nr)
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
		api.Serve(&pg, port, nr)
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
		"New Relic license key (default NEW_RELIC_LICENSE_KEY environment variable)",
	)
	return apiCmd
}
