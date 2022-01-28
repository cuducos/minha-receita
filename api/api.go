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

	"github.com/cuducos/go-cnpj"
)

type database interface {
	GetCompany(string) (string, error)
}

// errorMessage is a helper to serialize an error message to JSON.
type errorMessage struct {
	Message string `json:"message"`
}

// messageResponse takes a text message and a HTP status, wraps the message into a
// JSON output and writes it together with the proper headers to a response.
func messageResponse(w http.ResponseWriter, s int, m string) {
	w.WriteHeader(s)
	if m == "" {
		return
	}

	b, err := json.Marshal(errorMessage{m})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not wrap message in JSON: %s", m)
		return
	}
	w.Write(b)
}

type api struct {
	db database
}

func (app api) backwardCompatibilityHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return fmt.Errorf("no backward compatibilityt with method %s", r.Method)
	}

	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("invalid payload")
	}

	v := r.Form.Get("cnpj")
	if v == "" {
		return fmt.Errorf("no CNPJ sent in the payload")
	}

	v = cnpj.Unmask(v)
	if !cnpj.IsValid(v) {
		return fmt.Errorf("invalid CNPJ")
	}

	http.Redirect(w, r, fmt.Sprintf("/%s", v), http.StatusSeeOther)
	return nil
}

func (app api) companyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		err := app.backwardCompatibilityHandler(w, r)
		if err != nil {
			messageResponse(w, http.StatusMethodNotAllowed, "Essa URL aceita apenas o método GET.")
		}
		return
	}

	v := r.URL.Path
	if v == "/" {
		http.Redirect(w, r, "https://docs.minhareceita.org", http.StatusFound)
		return
	}

	if !cnpj.IsValid(v) {
		messageResponse(w, http.StatusBadRequest, fmt.Sprintf("CNPJ %s inválido.", cnpj.Mask(v[1:])))
		return
	}

	s, err := app.db.GetCompany(cnpj.Unmask(v))
	if err != nil {
		messageResponse(w, http.StatusNotFound, fmt.Sprintf("CNPJ %s não encontrado.", cnpj.Mask(v)))
		return
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, s)
}

func (app api) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		messageResponse(w, http.StatusMethodNotAllowed, "Essa URL aceita apenas o método GET.")
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Serve spins up the HTTP server.
func Serve(db database, p, n string) {
	if !strings.HasPrefix(p, ":") {
		p = ":" + p
	}
	nr := newRelicApp(n)
	app := api{db: db}
	http.HandleFunc(newRelicHandle(nr, "/", app.companyHandler))
	http.HandleFunc(newRelicHandle(nr, "/healthz", app.healthHandler))
	log.Fatal(http.ListenAndServe(p, nil))
}
