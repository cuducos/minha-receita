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

	"github.com/cuducos/minha-receita/db"
)

// errorMessage is a helper to serialize an error message to JSON.
type errorMessage struct {
	Message string `json:"message"`
}

// messageResponse takes a text message and a HTP status, wraps the message into a
// JSON output and writes it toghether with the proper headers to a response.
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
	db db.Database
}

func (app api) getHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")

	if r.Method != http.MethodGet {
		messageResponse(w, http.StatusMethodNotAllowed, "Essa URL aceita apenas o método GET.")
		return
	}

	v := r.URL.Path
	if v == "/" {
		messageResponse(w, http.StatusBadRequest, "CNPJ não enviado na requisição GET.")
		return
	}

	if !cnpj.IsValid(v) {
		messageResponse(w, http.StatusBadRequest, fmt.Sprintf("CNPJ %s inválido.", cnpj.Mask(v[1:])))
		return
	}

	c, err := app.db.GetCompany(cnpj.Unmask(v))
	if err != nil {
		messageResponse(w, http.StatusNoContent, "")
		return
	}

	w.WriteHeader(http.StatusOK)
	s, err := c.JSON()
	if err != nil {
		messageResponse(w, http.StatusInternalServerError, fmt.Sprintf("Não foi possível retornar os dados de %s em JSON.", cnpj.Mask(v)))
		return
	}
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
func Serve(db db.Database) {
	port := os.Getenv("PORT")
	if port == "" {
		log.Output(2, "No PORT environment variable found, using 8000.")
		port = ":8000"
	}

	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	nr := newRelicApp()
	app := api{db: db}
	http.HandleFunc(newRelicHandle(nr, "/", app.getHandler))
	http.HandleFunc(newRelicHandle(nr, "/healthz", app.healthHandler))
	log.Fatal(http.ListenAndServe(port, nil))
}
