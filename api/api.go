// Package api provides the HTTP server with wrappers for JSON responses. It
// validates data before passing it to the `db.Database`, which handles the
// query and serialization.
package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cuducos/go-cnpj"
	"github.com/cuducos/minha-receita/db"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	cacheMaxAge = time.Hour * 24
	timeout     = time.Second * 90
)

var cacheControl = fmt.Sprintf("max-age=%d", int(cacheMaxAge.Seconds()))

type database interface {
	GetCompany(string) (string, error)
	Search(context.Context, *db.Query) (string, error)
	MetaRead(string) (string, error)
}

type api struct {
	db   database
	host string
}

// messageResponse takes a text message and a HTTP status, wraps the message into a
// JSON output and writes it together with the proper headers to a response.
func (app *api) messageResponse(w http.ResponseWriter, s int, m string) {
	w.WriteHeader(s)
	if m != "" {
		w.Header().Set("Content-type", "application/json")
		if _, err := io.WriteString(w, fmt.Sprintf(`{"message":"%s"}`, m)); err != nil {
			slog.Error("could not write response message for", "status code", s, "message", m, "error", err)
		}
	}
	if s == http.StatusInternalServerError {
		slog.Error("Internal server error", "message", m)
	}
}

func (app *api) singleCompany(pth string, w http.ResponseWriter, r *http.Request, i int64) {
	w.Header().Set("Content-type", "application/json")
	if !cnpj.IsValid(pth) {
		app.messageResponse(w, http.StatusBadRequest, fmt.Sprintf("CNPJ %s inválido.", cnpj.Mask(pth[1:])))
		registerMetric("singleCompany", r.Method, http.StatusBadRequest, i)
		return
	}
	s, err := getCompany(app.db, pth)
	if err != nil {
		app.messageResponse(w, http.StatusNotFound, fmt.Sprintf("CNPJ %s não encontrado.", cnpj.Mask(pth)))
		registerMetric("singleCompany", r.Method, http.StatusNotFound, i)
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err := io.WriteString(w, s); err != nil {
		slog.Error("error responding to successful single company request", "request", r, "error", err)
	}
	registerMetric("singleCompany", r.Method, http.StatusOK, i)
}

func (app *api) paginatedSearch(q *db.Query, w http.ResponseWriter, r *http.Request, i int64) {
	w.Header().Set("Content-type", "application/json")
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	s, err := app.db.Search(ctx, q)
	if errors.Is(err, context.DeadlineExceeded) {
		slog.Error("paginated search timed out", "query", q)
		var b bytes.Buffer
		b.WriteString("Tempo de requisição esgotou (Timeout)")
		if q.Limit/2 > 1 {
			b.WriteString(fmt.Sprintf(
				". Essa busca solicitou %d CNPJs, experimente um número menor utilizando o parâmetro limit=%d, por exemplo.",
				q.Limit,
				q.Limit/2,
			))
		}
		app.messageResponse(w, http.StatusRequestTimeout, b.String())
		registerMetric("paginatedSearch", r.Method, http.StatusRequestTimeout, i)
		return
	}
	if err != nil {
		slog.Error("paginated search error", "error", err, "query", q)
		app.messageResponse(w, http.StatusNotFound, "Erro inesperado na busca.")
		registerMetric("paginatedSearch", r.Method, http.StatusNotFound, i)
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err := io.WriteString(w, s); err != nil {
		slog.Error("error responding to successful paginated search request", "query", q, "request", r, "error", err)
	}
	registerMetric("paginatedSearch", r.Method, http.StatusOK, i)
}

func (app *api) companyHandler(w http.ResponseWriter, r *http.Request) {
	i := time.Now().UnixMilli()
	w.Header().Set("Cache-Control", cacheControl)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")

	switch r.Method {
	case http.MethodGet:
		break
	case http.MethodOptions:
		w.WriteHeader(http.StatusOK)
		registerMetric("earlyReturn", r.Method, http.StatusOK, i)
		return
	default:
		app.messageResponse(w, http.StatusMethodNotAllowed, "Essa URL aceita apenas o método GET.")
		registerMetric("earlyReturn", r.Method, http.StatusMethodNotAllowed, i)
		return
	}
	pth := r.URL.Path
	if pth == "/" {
		q := db.NewQuery(r.URL.Query())
		if q == nil {
			http.Redirect(w, r, "https://docs.minhareceita.org", http.StatusFound)
			registerMetric("redirectedToDocs", r.Method, http.StatusFound, i)
			return
		}
		app.paginatedSearch(q, w, r, i)
		return
	}
	app.singleCompany(pth, w, r, i)
}

func (app *api) updatedHandler(w http.ResponseWriter, r *http.Request) {
	i := time.Now().UnixMilli()
	if r.Method != http.MethodGet {
		app.messageResponse(w, http.StatusMethodNotAllowed, "Essa URL aceita apenas o método GET.")
		registerMetric("updated", r.Method, http.StatusMethodNotAllowed, i)
		return
	}
	s, err := app.db.MetaRead("updated-at")
	if err != nil {
		app.messageResponse(w, http.StatusInternalServerError, "Erro buscando data de atualização.")
		registerMetric("updated", r.Method, http.StatusInternalServerError, i)
		return
	}
	if s == "" {
		w.WriteHeader(http.StatusInternalServerError)
		registerMetric("updated", r.Method, http.StatusInternalServerError, i)
		return
	}
	w.Header().Set("Cache-Control", cacheControl)
	app.messageResponse(w, http.StatusOK, fmt.Sprintf("%s é a data de extração dos dados pela Receita Federal.", s))
	registerMetric("updated", r.Method, http.StatusOK, i)
}

func (app *api) healthHandler(w http.ResponseWriter, r *http.Request) {
	i := time.Now().UnixMilli()
	if r.Method != http.MethodHead && r.Method != http.MethodGet {
		app.messageResponse(w, http.StatusMethodNotAllowed, "Essa URL aceita apenas os métodos GET e HEAD.")
		registerMetric("health", r.Method, http.StatusMethodNotAllowed, i)
		return
	}
	w.WriteHeader(http.StatusOK)
	registerMetric("health", r.Method, http.StatusOK, i)
}

func (app *api) allowedHostWrapper(h func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	if app.host == "" {
		return h
	}
	w := func(w http.ResponseWriter, r *http.Request) {
		if v := r.Header.Get("Host"); v != app.host {
			slog.Error("Host not allowed", "host", v)
			w.WriteHeader(http.StatusTeapot)
			return
		}
		h(w, r)
	}
	return w
}

// Serve spins up the HTTP server.
func Serve(db database, p string) error {
	if !strings.HasPrefix(p, ":") {
		p = ":" + p
	}
	app := api{db, os.Getenv("ALLOWED_HOST")}
	for _, r := range []struct {
		path    string
		handler func(http.ResponseWriter, *http.Request)
	}{
		{"/", app.companyHandler},
		{"/updated", app.updatedHandler},
		{"/healthz", app.healthHandler},
		{"/metrics", promhttp.Handler().ServeHTTP},
	} {
		http.HandleFunc(r.path, app.allowedHostWrapper(r.handler))
	}
	s := &http.Server{Addr: p, ReadTimeout: timeout * 2, WriteTimeout: timeout * 2}
	slog.Info(fmt.Sprintf("Serving at http://0.0.0.0%s", p))
	return s.ListenAndServe()
}
