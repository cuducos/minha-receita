// Package db implements the high level API for a database interface. The lines
// in this file should be agnostic in terms of the database provider.
//
// Files such as `postgres.go` and `postgres_sql.go` implements a
// specific database provider.
//
// `postgres.go` defines the high level methods described in the `db.Database`
// interface, as well as a `NewPostgreSQL` method to create this database.
//
// `postgres_sql.go` implements the SQL queries to run the database.
package db

import (
	"encoding/json"
	"os"
	"time"

	"github.com/cuducos/go-cnpj"
)

var pgURI = os.Getenv("POSTGRES_URI")

// Database interface to Minha Receita.
type Database interface {
	CreateTables()
	DropTables()
	ImportData(string)
	GetCompany(string) (Company, error)
}

// CNAE represents a row from the `cnae` database table.
type CNAE struct {
	Codigo    int    `json:"codigo"`
	Descricao string `json:"descricao"`
}

// Partner represents a row from the `socio` database table.
type Partner struct {
	Cnpj                                 string    `json:"cnpj"`
	IdentificadorDeSocio                 int       `json:"identificador_de_socio"`
	NomeSocio                            string    `json:"nome_socio"`
	CnpjCpfDoSocio                       string    `json:"cnpj_cpf_do_socio"`
	CodigoQualificacaoSocio              int       `json:"codigo_qualificacao_socio"`
	PercentualCapitalSocial              int       `json:"percentual_capital_social"`
	DataEntradaSociedade                 time.Time `json:"data_entrada_sociedade"`
	CpfRepresentanteLegal                string    `json:"cpf_representante_legal"`
	NomeRepresentanteLegal               string    `json:"nome_representante_legal"`
	CodigoQualificacaoRepresentanteLegal int       `json:"codigo_qualificacao_representante_legal"`
}

// Company represents a row from the `empresa` database table.
type Company struct {
	Cnpj                      string     `json:"cnpj"`
	IdentificadorMatrizFilial int        `json:"identificador_matriz_filial"`
	RazaoSocial               string     `json:"razao_social"`
	NomeFantasia              string     `json:"nome_fantasia"`
	SituacaoCadastral         int        `json:"situacao_cadastral"`
	DataSituacaoCadastral     time.Time  `json:"data_situacao_cadastral"`
	MotivoSituacaoCadastral   int        `json:"motivo_situacao_cadastral"`
	NomeCidadeExterior        string     `json:"nome_cidade_exterior"`
	CodigoNaturezaJuridica    int        `json:"codigo_natureza_juridica"`
	DataInicioAtividade       time.Time  `json:"data_inicio_atividade"`
	CNAEFiscal                int        `json:"cnae_fiscal"`
	CNAEFiscalDescricao       string     `json:"cnae_fiscal_descricao"`
	DescricaoTipoLogradouro   string     `json:"descricao_tipo_logradouro"`
	Logradouro                string     `json:"logradouro"`
	Numero                    string     `json:"numero"`
	Complemento               string     `json:"complemento"`
	Bairro                    string     `json:"bairro"`
	Cep                       int        `json:"cep"`
	Uf                        string     `json:"uf"`
	CodigoMunicipio           int        `json:"codigo_municipio"`
	Municipio                 string     `json:"municipio"`
	DddTelefone1              string     `json:"ddd_telefone_1"`
	DddTelefone2              string     `json:"ddd_telefone_2"`
	DddFax                    string     `json:"ddd_fax"`
	QualificacaoDoResponsavel int        `json:"qualificacao_do_responsavel"`
	CapitalSocial             float32    `json:"capital_social"`
	Porte                     int        `json:"porte"`
	OpcaoPeloSimples          bool       `json:"opcao_pelo_simples"`
	DataOpcaoPeloSimples      string     `json:"data_opcao_pelo_simples"`
	DataExclusaoDoSimples     string     `json:"data_exclusao_do_simples"`
	OpcaoPeloMei              bool       `json:"opcao_pelo_mei"`
	SituacaoEspecial          string     `json:"situacao_especial"`
	DataSituacaoEspecial      string     `json:"data_situacao_especial"`
	Qsa                       []*Partner `json:"qsa"`
	CNAEsSecundarias          []*CNAE    `json:"cnaes_secundarias"`
}

// JSON outputs a `Company` as a valid JSON string.
func (c *Company) JSON() (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *Company) String() string {
	return cnpj.Mask(c.Cnpj)
}
