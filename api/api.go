// Package api provides the HTTP server with wrappers for JSON responses. It
// validates data before passing it to the `db.Database`, which handles the
// query and serialization.
package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cuducos/go-cnpj"
	"github.com/cuducos/minha-receita/db"
	"github.com/cuducos/minha-receita/monitor"
	"github.com/newrelic/go-agent/v3/newrelic"
)

const (
	cacheMaxAge = time.Hour * 24
	timeout     = time.Minute * 1
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
		io.WriteString(w, fmt.Sprintf(`{"message":"%s"}`, m))
	}
	if s == http.StatusInternalServerError {
		slog.Error("Internal server error", "message", m)
	}
}

func (app *api) singleCompany(pth string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	txn := newrelic.FromContext(r.Context())
	if txn != nil {
		txn.AddAttribute("handler", "singleCompany")
	}
	if !cnpj.IsValid(pth) {
		app.messageResponse(w, http.StatusBadRequest, fmt.Sprintf("CNPJ %s inválido.", cnpj.Mask(pth[1:])))
		return
	}
	s, err := getCompany(app.db, pth)
	if err != nil {
		app.messageResponse(w, http.StatusNotFound, fmt.Sprintf("CNPJ %s não encontrado.", cnpj.Mask(pth)))
		return
	}
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, s)
}

func (app *api) paginatedSearch(q *db.Query, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	txn := newrelic.FromContext(r.Context())
	if txn != nil {
		if q.Compact {
			txn.AddAttribute("handler", "compactPaginatedSearch")
		} else {
			txn.AddAttribute("handler", "paginatedSearch")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	s, err := app.db.Search(ctx, q)
	if err == context.DeadlineExceeded {
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
		if !q.Compact {
			b.WriteString(". Experimente adicionar compact=true na URL para listar apenas os números de CNPJ, diminuindo o tempo de resposta.")
		}
		app.messageResponse(w, http.StatusRequestTimeout, b.String())
		return
	}
	if err != nil {
		slog.Error("paginated search error", "error", err, "query", q)
		app.messageResponse(w, http.StatusNotFound, "Erro inesperado na busca.")
		return
	}
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, s)
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
	pth := r.URL.Path
	if pth == "/" {
		q := db.NewQuery(r.URL.Query())
		if q == nil {
			http.Redirect(w, r, "https://docs.minhareceita.org", http.StatusFound)
			return
		}
		app.paginatedSearch(q, w, r)
		return
	}
	app.singleCompany(pth, w, r)
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
		app.messageResponse(w, http.StatusMethodNotAllowed, "Essa URL aceita apenas os métodos GET e HEAD.")
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
			slog.Error("Host not allowed", "host", v)
			w.WriteHeader(http.StatusTeapot)
			return
		}
		h(w, r)
	}
	return w
}

// Serve spins up the HTTP server.
func Serve(db database, p string, nr *newrelic.Application) error {
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
	} {
		http.HandleFunc(monitor.NewRelicHandle(nr, r.path, app.allowedHostWrapper(r.handler)))
	}
	s := &http.Server{Addr: p}
	slog.Info(fmt.Sprintf("Serving at http://0.0.0.0%s", p))
	return s.ListenAndServe()
}
