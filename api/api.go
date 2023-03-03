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

// messageResponse takes a text message and a HTTP status, wraps the message into a
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
	w.Header().Set("Content-type", "application/json")
	w.Write(b)
}

type api struct {
	db   database
	host string
}

func (app *api) companyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", cacheControl)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")

	switch r.Method {
	case http.MethodGet:
		break
	case http.MethodOptions:
		w.WriteHeader(http.StatusOK)
		return
	default:
		messageResponse(w, http.StatusMethodNotAllowed, "Essa URL aceita apenas o método GET.")
		return
	}

	v := r.URL.Path
	command := "" // "" = returns all data.

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

	//check if the url contains "."
	if strings.Contains(v, ".") {
		command = strings.Split(v, "/")[2]
	} else {
		command = strings.Split(v, "/")[1]
	}

	//Não existe comando, retorna todos os dados
	if command != "" {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, s)
		return
	}

	//create a map to store the json
	var data map[string]interface{}
	json.Unmarshal([]byte(s), &data)

	//if command contains ".", it means that the user wants to get a specific data
	if strings.Contains(command, ".") {
		//split the command to get the path
		path := strings.Split(command, ".")

		//iterate over the path to get the data
		for i := 0; i < len(path); i++ {
			//if the data exists, return it
			if i == len(path)-1 {
				data = data[path[i]].(map[string]interface{})
			} else {
				messageResponse(w, http.StatusNotFound, fmt.Sprintf("dados %s do CNPJ %s não encontrados.", path[i], cnpj.Mask(v)))
			}
		}
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, fmt.Sprintf("%v", data))
}

func (app *api) updatedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		messageResponse(w, http.StatusMethodNotAllowed, "Essa URL aceita apenas o método GET.")
		return
	}
	s, err := app.db.MetaRead("updated-at")
	if err != nil {
		messageResponse(w, http.StatusInternalServerError, "Erro buscando data de atualização.")
		return
	}
	if s == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Cache-Control", cacheControl)
	messageResponse(w, http.StatusOK, fmt.Sprintf("%s é a data de extração dos dados pela Receita Federal.", s))
}

func (app *api) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		messageResponse(w, http.StatusMethodNotAllowed, "Essa URL aceita apenas o método GET.")
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
func Serve(db database, p, n string) {
	if !strings.HasPrefix(p, ":") {
		p = ":" + p
	}
	nr := newRelicApp(n)
	app := api{db: db, host: os.Getenv("ALLOWED_HOST")}
	for _, r := range []struct {
		path    string
		handler func(http.ResponseWriter, *http.Request)
	}{
		{"/", app.companyHandler},
		{"/updated", app.updatedHandler},
		{"/healthz", app.healthHandler},
	} {
		http.HandleFunc(newRelicHandle(nr, r.path, app.allowedHostWrapper(r.handler)))
	}
	log.Output(1, fmt.Sprintf("Serving at http://0.0.0.0%s", p))
	log.Fatal(http.ListenAndServe(p, nil))
}
