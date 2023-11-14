package monitor

import (
	"fmt"
	"net/http"

	"github.com/newrelic/go-agent/v3/newrelic"
)

func NewRelicApp(k string) (*newrelic.Application, error) {
	if k == "" {
		return nil, nil
	}
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName("Minha Receita"),
		newrelic.ConfigLicense(k),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating new relic app: %w", err)
	}
	return app, nil
}

func NewRelicHandle(app *newrelic.Application, pth string, f http.HandlerFunc) (string, http.HandlerFunc) {
	if app == nil {
		return pth, f
	}

	return newrelic.WrapHandleFunc(app, pth, f)
}
