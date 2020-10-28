package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/cuducos/go-cnpj"
)

// MockDatabase is a interface of Database to be used in tests.
type MockDatabase struct {
}

// GetCompany only returns values for the CNPJ 19.131.243/0001-97
func (MockDatabase) GetCompany(c string) string {
	if cnpj.Unmask(c) == "19131243000197" {
		return "Yay!"
	}
	return ""
}

func TestCompanyHandler(t *testing.T) {
	cases := []struct {
		method  string
		data    map[string]string
		status  int
		content string
	}{
		{
			http.MethodGet,
			nil,
			http.StatusMethodNotAllowed,
			`{"message":"Essa URL aceita apenas o método POST."}`,
		},
		{
			http.MethodPost,
			nil,
			http.StatusBadRequest,
			`{"message":"Conteúdo inválido na requisição POST."}`,
		},
		{
			http.MethodPost,
			map[string]string{"cpf": "foobar"},
			http.StatusBadRequest,
			`{"message":"CNPJ não enviado na requisição POST."}`,
		},
		{
			http.MethodPost,
			map[string]string{"cnpj": "foobar"},
			http.StatusBadRequest,
			`{"message":"CNPJ foobar inválido."}`,
		},
		{
			http.MethodPost,
			map[string]string{"cnpj": "00.000.000/0001-91"},
			http.StatusNotFound,
			`{"message":"CNPJ 00.000.000/0001-91 não encontrado."}`,
		},
		{
			http.MethodPost,
			map[string]string{"cnpj": "00000000000191"},
			http.StatusNotFound,
			`{"message":"CNPJ 00.000.000/0001-91 não encontrado."}`,
		},
		{
			http.MethodPost,
			map[string]string{"cnpj": "19.131.243/0001-97"},
			http.StatusOK,
			"Yay!", // TODO add proper content
		},
		{
			http.MethodPost,
			map[string]string{"cnpj": "19131243000197"},
			http.StatusOK,
			"Yay!", // TODO add proper content
		},
	}

	for _, c := range cases {
		var b io.Reader
		if c.data != nil {
			d := url.Values{}
			for k, v := range c.data {
				d.Set(k, v)
			}
			b = strings.NewReader(d.Encode())
		}
		req, err := http.NewRequest(c.method, "/", b)
		if c.method == http.MethodPost {
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		}

		if err != nil {
			t.Fatal("Expected an HTTP response, but got an error.")
		}

		app := API{&MockDatabase{}}
		resp := httptest.NewRecorder()
		handler := http.HandlerFunc(app.Handler)
		handler.ServeHTTP(resp, req)

		if resp.Code != c.status {
			t.Errorf("Expected %s to return %v, but got %v", c.method, c.status, resp.Code)
		}

		if c := resp.Header().Get("Content-type"); c != "application/json" {
			t.Errorf("\nExpected content-type to be application/json, but got %s", c)
		}

		if resp.Body.String() != c.content {
			t.Errorf("\nExpected HTTP contents to be:\n\t%s\nGot:\n\t%s", c.content, resp.Body.String())
		}
	}

}
