package api

import (
	"net/http"
	"os"

	"github.com/newrelic/go-agent/v3/newrelic"
)

func newRelicApp() *newrelic.Application {
	k := os.Getenv("NEW_RELIC_LICENSE_KEY")
	if k == "" {
		return nil
	}

	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName("Minha Receita"),
		newrelic.ConfigLicense(k),
	)
	if err != nil {
		panic(err)
	}
	return app
}

func newRelicHandle(app *newrelic.Application, pth string, f func(http.ResponseWriter, *http.Request)) (string, func(http.ResponseWriter, *http.Request)) {
	if app == nil {
		return pth, f
	}
	return newrelic.WrapHandleFunc(app, pth, f)
}
