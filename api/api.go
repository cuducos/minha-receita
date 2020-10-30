package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/cuducos/go-cnpj"
)

// errorMessage is a helper to serialize an error message to JSON.
type errorMessage struct {
	Message string `json:"message"`
}

// writeError takes a text message and a HTP status, wraps the message into a
// JSON output and writes it toghether with the proper headers to a response.
func writeError(w http.ResponseWriter, m string, s int) {
	b, err := json.Marshal(errorMessage{m})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not wrap message in JSON: %s", m)
		return
	}

	w.WriteHeader(s)
	w.Write(b)
}

// API wraps the HTTP request handler and database interface.
type API struct {
	db database
}

// PostHandler wraps the database interface in a HTTP request/response cycle.
func (app API) PostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")

	if r.Method != http.MethodPost {
		writeError(w, "Essa URL aceita apenas o método POST.", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		writeError(w, "Conteúdo inválido na requisição POST.", http.StatusBadRequest)
		return
	}

	v := r.Form.Get("cnpj")
	if v == "" {
		writeError(w, "CNPJ não enviado na requisição POST.", http.StatusBadRequest)
		return
	}

	if !cnpj.IsValid(v) {
		writeError(w, fmt.Sprintf("CNPJ %s inválido.", cnpj.Mask(v)), http.StatusBadRequest)
		return
	}

	c := app.db.GetCompany(v)
	if c == "" {
		writeError(w, fmt.Sprintf("CNPJ %s não encontrado.", cnpj.Mask(v)), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, c)
}
