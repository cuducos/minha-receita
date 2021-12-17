package api

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
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
		req, err := http.NewRequest(c.method, c.path, nil)
		if err != nil {
			t.Fatal("Expected an HTTP request, but got an error.")
		}
		if c.method == http.MethodPost {
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		}

		app := api{&mockDatabase{}}
		resp := httptest.NewRecorder()
		handler := http.HandlerFunc(app.companyHandler)
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

func TestCompanyHandlerBackwardCompatibility(t *testing.T) {
	cases := []struct {
		payload map[string]string
		status  int
		body    string
	}{
		{
			map[string]string{"cpf": "123"},
			http.StatusMethodNotAllowed,
			`{"message":"Essa URL aceita apenas o método GET."}`,
		},
		{
			map[string]string{"cnpj": "123"},
			http.StatusMethodNotAllowed,
			`{"message":"Essa URL aceita apenas o método GET."}`,
		},
		{
			map[string]string{"cnpj": "19131243000197"},
			http.StatusSeeOther,
			"",
		},
		{
			map[string]string{"cnpj": "19.131.243/0001-97"},
			http.StatusSeeOther,
			"",
		},
	}

	for _, c := range cases {
		d := url.Values{}
		for k, v := range c.payload {
			d.Set(k, v)
		}

		req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(d.Encode()))
		if err != nil {
			t.Fatal("Expected an HTTP request, but got an error.")
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		app := api{&mockDatabase{}}
		resp := httptest.NewRecorder()
		handler := http.HandlerFunc(app.companyHandler)
		handler.ServeHTTP(resp, req)

		if resp.Code != c.status {
			t.Errorf("Expected POST to return %v, but got %v", c.status, resp.Code)
		}

		if c := resp.Header().Get("Content-type"); c != "application/json" {
			t.Errorf("\nExpected content-type to be application/json, but got %s", c)
		}

		if strings.TrimSpace(resp.Body.String()) != c.body {
			t.Errorf("\nExpected HTTP contents to be:\n\t%s\nGot:\n\t%s", c.body, resp.Body.String())
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
