package api

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/cuducos/go-cnpj"

	"github.com/cuducos/minha-receita/db"
)

const expected = `{"cnpj":"19131243000197","identificador_matriz_filial":0,"razao_social":"","nome_fantasia":"","situacao_cadastral":0,"data_situacao_cadastral":"0001-01-01T00:00:00Z","motivo_situacao_cadastral":0,"nome_cidade_exterior":"","codigo_natureza_juridica":0,"data_inicio_atividade":"0001-01-01T00:00:00Z","cnae_fiscal":0,"cnae_fiscal_descricao":"","descricao_tipo_logradouro":"","logradouro":"","numero":"","complemento":"","bairro":"","cep":0,"uf":"","codigo_municipio":0,"municipio":"","ddd_telefone_1":"","ddd_telefone_2":"","ddd_fax":"","qualificacao_do_responsavel":0,"capital_social":0,"porte":0,"opcao_pelo_simples":false,"data_opcao_pelo_simples":"","data_exclusao_do_simples":"","opcao_pelo_mei":false,"situacao_especial":"","data_situacao_especial":"","qsa":null,"cnaes_secundarias":null}`

type mockDatabase struct{}

func (mockDatabase) CreateTables()       {}
func (mockDatabase) DropTables()         {}
func (mockDatabase) ImportData(_ string) {}

func (mockDatabase) GetCompany(n string) (db.Company, error) {
	var c db.Company
	n = cnpj.Unmask(n)
	if n == "19131243000197" {
		return db.Company{Cnpj: n}, nil
	}
	return c, errors.New("Company not found")
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
			expected,
		},
		{
			http.MethodPost,
			map[string]string{"cnpj": "19131243000197"},
			http.StatusOK,
			expected,
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

		app := api{&mockDatabase{}}
		resp := httptest.NewRecorder()
		handler := http.HandlerFunc(app.postHandler)
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
