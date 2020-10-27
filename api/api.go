package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/cuducos/go-cnpj"
)

// JSONMessage is a helper to serialize a message to JSON.
type JSONMessage struct {
	Message string `json:"message"`
}

// WriteJSONMessage takes a text message, wraps it into a JSON output and
// writes it to a `https.ResponseWriter`.
func WriteJSONMessage(w http.ResponseWriter, m string) {
	j := JSONMessage{m}
	b, err := json.Marshal(j)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not wrap message in JSON: %s", m)
		return

	}
	w.Write(b)
}

// API wraps the HTTP request handler and database interface.
type API struct {
	db Database
}

// Handler wraps the database interface in a HTTP request/response cycle.
func (app API) Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		WriteJSONMessage(w, "Method GET not allowed for URL /")
		return
	}

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		WriteJSONMessage(w, "Conteúdo inválido na requisição POST.")
		return
	}

	v := r.Form.Get("cnpj")
	if v == "" {
		w.WriteHeader(http.StatusNotFound)
		WriteJSONMessage(w, "CNPJ não enviado na requisição POST.")
		return
	}

	if !cnpj.IsValid(v) {
		w.WriteHeader(http.StatusNotFound)
		WriteJSONMessage(w, fmt.Sprintf("CNPJ %s inválido.", cnpj.Mask(v)))
		return
	}

	c := app.db.GetCompany(v)
	if c == "" {
		w.WriteHeader(http.StatusNotFound)
		WriteJSONMessage(w, fmt.Sprintf("CNPJ %s não encontrado.", cnpj.Mask(v)))
		return
	}

	io.WriteString(w, c)
}
