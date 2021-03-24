package api

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cuducos/go-cnpj"

	"github.com/cuducos/minha-receita/db"
)

type mockDatabase struct{}

func (mockDatabase) CreateTables()       {}
func (mockDatabase) DropTables()         {}
func (mockDatabase) ImportData(_ string) {}

func (mockDatabase) GetCompany(n string) (db.Company, error) {
	var c db.Company
	n = cnpj.Unmask(n)
	if n == "19131243000197" {
		return db.Company{
			CNPJ:                "19131243000197",
			DataInicioAtividade: time.Date(2013, time.October, 3, 0, 0, 0, 0, time.UTC),
		}, nil
	}
	return c, errors.New("Company not found")
}

func TestCompanyHandler(t *testing.T) {
	f, err := filepath.Abs(filepath.Join("..", "testdata", "response.json"))
	if err != nil {
		t.Errorf("Could understand path %s", f)
	}
	b, err := ioutil.ReadFile(f)
	if err != nil {
		t.Errorf("Could not read from %s", f)
	}
	expected := strings.TrimSpace(string(b))

	cases := []struct {
		method  string
		path    string
		status  int
		content string
	}{
		{
			http.MethodHead,
			"/",
			http.StatusMethodNotAllowed,
			`{"message":"Essa URL aceita apenas o método GET."}`,
		},
		{
			http.MethodPost,
			"/",
			http.StatusMethodNotAllowed,
			`{"message":"Essa URL aceita apenas o método GET."}`,
		},
		{
			http.MethodGet,
			"/",
			http.StatusBadRequest,
			`{"message":"CNPJ não enviado na requisição GET."}`,
		},
		{
			http.MethodGet,
			"/foobar",
			http.StatusBadRequest,
			`{"message":"CNPJ foobar inválido."}`,
		},
		{
			http.MethodGet,
			"/00.000.000/0001-91",
			http.StatusNoContent,
			"",
		},
		{
			http.MethodGet,
			"/00000000000191",
			http.StatusNoContent,
			"",
		},
		{
			http.MethodGet,
			"/19.131.243/0001-97",
			http.StatusOK,
			expected,
		},
		{
			http.MethodGet,
			"/19131243000197",
			http.StatusOK,
			expected,
		},
	}

	for _, c := range cases {
		req, err := http.NewRequest(c.method, c.path, nil)
		if err != nil {
			t.Fatal("Expected an HTTP request, but got an error.")
		}
		if c.method == http.MethodPost {
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		}

		app := api{&mockDatabase{}}
		resp := httptest.NewRecorder()
		handler := http.HandlerFunc(app.getHandler)
		handler.ServeHTTP(resp, req)

		if resp.Code != c.status {
			t.Errorf("Expected %s to return %v, but got %v", c.method, c.status, resp.Code)
		}

		if c := resp.Header().Get("Content-type"); c != "application/json" {
			t.Errorf("\nExpected content-type to be application/json, but got %s", c)
		}

		if strings.TrimSpace(resp.Body.String()) != c.content {
			t.Errorf("\nExpected HTTP contents to be:\n\t%s\nGot:\n\t%s", c.content, resp.Body.String())
		}
	}
}

func TestHealthHandler(t *testing.T) {
	cases := []struct {
		method  string
		status  int
		content string
	}{
		{
			http.MethodGet,
			http.StatusOK,
			"",
		},
		{
			http.MethodPost,
			http.StatusMethodNotAllowed,
			`{"message":"Essa URL aceita apenas o método GET."}`,
		},
		{
			http.MethodHead,
			http.StatusMethodNotAllowed,
			`{"message":"Essa URL aceita apenas o método GET."}`,
		},
	}

	for _, c := range cases {
		req, err := http.NewRequest(c.method, "/healthz", nil)
		if err != nil {
			t.Fatal("Expected an HTTP request, but got an error.")
		}
		app := api{&mockDatabase{}}
		resp := httptest.NewRecorder()
		handler := http.HandlerFunc(app.healthHandler)
		handler.ServeHTTP(resp, req)

		if resp.Code != c.status {
			t.Errorf("Expected %s /healthz to return %v, but got %v", c.method, c.status, resp.Code)
		}
		if strings.TrimSpace(resp.Body.String()) != c.content {
			t.Errorf("\nExpected HTTP contents to be %s, got %s", c.content, resp.Body.String())
		}
	}
}
