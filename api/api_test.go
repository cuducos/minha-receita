package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cuducos/go-cnpj"
)

type mockDatabase struct{}

func (mockDatabase) GetCompany(n string) (string, error) {
	n = cnpj.Unmask(n)
	if n != "19131243000197" {
		return "", errors.New("Company not found")
	}

	b, err := os.ReadFile(filepath.Join("..", "testdata", "response.json"))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (mockDatabase) MetaRead(k string) (string, error) { return "42", nil }

func TestCompanyHandler(t *testing.T) {
	f, err := filepath.Abs(filepath.Join("..", "testdata", "response.json"))
	if err != nil {
		t.Errorf("Could understand path %s", f)
	}
	b, err := os.ReadFile(f)
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
			http.MethodOptions,
			"/",
			http.StatusOK,
			"",
		},
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
			http.StatusFound,
			"",
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
			http.StatusNotFound,
			`{"message":"CNPJ 00.000.000/0001-91 não encontrado."}`,
		},
		{
			http.MethodGet,
			"/00000000000191",
			http.StatusNotFound,
			`{"message":"CNPJ 00.000.000/0001-91 não encontrado."}`,
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
		t.Run(fmt.Sprintf("%s %s", c.method, c.path), func(t *testing.T) {
			req, err := http.NewRequest(c.method, c.path, nil)
			if err != nil {
				t.Fatal("Expected an HTTP request, but got an error.")
			}
			if c.method == http.MethodPost {
				req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			}

			app := api{db: &mockDatabase{}}
			resp := httptest.NewRecorder()
			handler := http.HandlerFunc(app.companyHandler)
			handler.ServeHTTP(resp, req)

			if resp.Code != c.status {
				t.Errorf("Expected %s to return %v, but got %v", c.method, c.status, resp.Code)
			}
			if c.content != "" {
				if body := strings.TrimSpace(resp.Body.String()); body != c.content {
					t.Errorf("\nExpected HTTP contents to be:\n\t%s\nGot:\n\t%s", c.content, resp.Body.String())
				}
				if c := resp.Header().Get("Content-type"); c != "application/json" {
					t.Errorf("\nExpected content-type to be application/json, but got %s", c)
				}
			}
		})
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
			http.StatusOK,
			"",
		},
	}

	for _, c := range cases {
		req, err := http.NewRequest(c.method, "/healthz", nil)
		if err != nil {
			t.Fatal("Expected an HTTP request, but got an error.")
		}
		app := api{db: &mockDatabase{}}
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

func TestUpdatedHandler(t *testing.T) {
	app := api{db: &mockDatabase{}}
	for _, c := range []struct {
		method  string
		status  int
		content string
	}{
		{http.MethodGet, http.StatusOK, `{"message":"42 é a data de extração dos dados pela Receita Federal."}`},
		{http.MethodPost, http.StatusMethodNotAllowed, `{"message":"Essa URL aceita apenas o método GET."}`},
		{http.MethodHead, http.StatusMethodNotAllowed, `{"message":"Essa URL aceita apenas o método GET."}`},
		{http.MethodOptions, http.StatusMethodNotAllowed, `{"message":"Essa URL aceita apenas o método GET."}`},
	} {
		req, err := http.NewRequest(c.method, "/updated", nil)
		if err != nil {
			t.Fatal("Expected an HTTP request, but got an error.")
		}
		resp := httptest.NewRecorder()
		handler := http.HandlerFunc(app.updatedHandler)
		handler.ServeHTTP(resp, req)

		if resp.Code != c.status {
			t.Errorf("Expected %s /urls to return %v, but got %v", c.method, c.status, resp.Code)
		}
		if strings.TrimSpace(resp.Body.String()) != c.content {
			t.Errorf("\nExpected HTTP contents to be %s, got %s", c.content, resp.Body.String())
		}
	}
}

func TestAllowedHostWrap(t *testing.T) {
	for _, c := range []struct {
		allowedHost string
		status      int
	}{
		{"", http.StatusOK},
		{"127.0.0.1", http.StatusOK},
		{"forty-two", http.StatusTeapot},
	} {
		t.Run(fmt.Sprintf("test returns %d when allowed host is %s", c.status, c.allowedHost), func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/19131243000197", nil)
			req.Header.Set("Host", "127.0.0.1")
			if err != nil {
				t.Fatal("Expected an HTTP request, but got an error.")
			}
			resp := httptest.NewRecorder()
			app := api{db: &mockDatabase{}, host: c.allowedHost}
			handler := http.HandlerFunc(app.allowedHostWrapper(app.companyHandler))
			handler.ServeHTTP(resp, req)
			if resp.Code != c.status {
				t.Errorf("Expected request with allowed host `%s` to return %d, but got %d (request header had `%s`) ", c.allowedHost, c.status, resp.Code, req.Header.Get("Host"))
			}
		})
	}

}
