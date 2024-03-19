// Package api provides the HTTP server with wrappers for JSON responses. It
// validates data before passing it to the `db.Database`, which handles the
// query and serialization.
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cuducos/go-cnpj"
	"github.com/cuducos/minha-receita/monitor"
	"github.com/newrelic/go-agent/v3/integrations/logcontext-v2/logWriter"
	"github.com/newrelic/go-agent/v3/newrelic"
)

const cacheMaxAge = time.Hour * 24

var cacheControl = fmt.Sprintf("max-age=%d", int(cacheMaxAge.Seconds()))

type database interface {
	GetCompany(string) (string, error)
	MetaRead(string) (string, error)
}

// errorMessage is a helper to serialize an error message to JSON.
type errorMessage struct {
	Message string `json:"message"`
}

type api struct {
	db          database
	host        string
	errorLogger logWriter.LogWriter
}

// messageResponse takes a text message and a HTTP status, wraps the message into a
// JSON output and writes it together with the proper headers to a response.
func (app *api) messageResponse(w http.ResponseWriter, s int, m string) {
	if m == "" {
		w.WriteHeader(s)
		if s == http.StatusInternalServerError {
			app.errorLogger.Write([]byte("Internal server error without error message"))
		}
		return
	}

	b, err := json.Marshal(errorMessage{m})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.errorLogger.Write([]byte(fmt.Sprintf("Could not wrap message in JSON: %s", m)))
		return
	}

	w.WriteHeader(s)
	w.Header().Set("Content-type", "application/json")
	w.Write(b)

	if s == http.StatusInternalServerError {
		app.errorLogger.Write(b)
	}
}

func (app *api) companyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", cacheControl)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")

	switch r.Method {
	case http.MethodGet:
		break
	case http.MethodOptions:
		w.WriteHeader(http.StatusOK)
		return
	default:
		app.messageResponse(w, http.StatusMethodNotAllowed, "Essa URL aceita apenas o método GET.")
		return
	}

	v := r.URL.Path
	if v == "/" {
		http.Redirect(w, r, "https://docs.minhareceita.org", http.StatusFound)
		return
	}
	if !cnpj.IsValid(v) {
		app.messageResponse(w, http.StatusBadRequest, fmt.Sprintf("CNPJ %s inválido.", cnpj.Mask(v[1:])))
		return
	}

	s, err := app.db.GetCompany(cnpj.Unmask(v))
	if err != nil {
		app.messageResponse(w, http.StatusNotFound, fmt.Sprintf("CNPJ %s não encontrado.", cnpj.Mask(v)))
		return
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, s)
}

func (app *api) updatedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		app.messageResponse(w, http.StatusMethodNotAllowed, "Essa URL aceita apenas o método GET.")
		return
	}
	s, err := app.db.MetaRead("updated-at")
	if err != nil {
		app.messageResponse(w, http.StatusInternalServerError, "Erro buscando data de atualização.")
		return
	}
	if s == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Cache-Control", cacheControl)
	app.messageResponse(w, http.StatusOK, fmt.Sprintf("%s é a data de extração dos dados pela Receita Federal.", s))
}

func (app *api) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodHead && r.Method != http.MethodGet {
		app.messageResponse(w, http.StatusMethodNotAllowed, "Essa URL aceita apenas o método GET.")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (app *api) allowedHostWrapper(h func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	if app.host == "" {
		return h
	}
	w := func(w http.ResponseWriter, r *http.Request) {
		if v := r.Header.Get("Host"); v != app.host {
			log.Output(1, fmt.Sprintf("Host %s not allowed", v))
			w.WriteHeader(http.StatusTeapot)
			return
		}
		h(w, r)
	}
	return w
}

// Serve spins up the HTTP server.
func Serve(db database, p string, nr *newrelic.Application) {
	if !strings.HasPrefix(p, ":") {
		p = ":" + p
	}
	app := api{
		db:          db,
		host:        os.Getenv("ALLOWED_HOST"),
		errorLogger: logWriter.New(os.Stderr, nr),
	}
	for _, r := range []struct {
		path    string
		handler func(http.ResponseWriter, *http.Request)
	}{
		{"/", app.companyHandler},
		{"/updated", app.updatedHandler},
		{"/healthz", app.healthHandler},
	} {
		http.HandleFunc(monitor.NewRelicHandle(nr, r.path, app.allowedHostWrapper(r.handler)))
	}
	log.Output(1, fmt.Sprintf("Serving at http://0.0.0.0%s", p))
	log.Fatal(http.ListenAndServe(p, nil))
}
