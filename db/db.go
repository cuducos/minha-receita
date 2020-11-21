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
	"fmt"
	"os"
	"time"

	"github.com/cuducos/go-cnpj"
)

var pgURI = os.Getenv("POSTGRES_URI")

// Date wraps time.Time as type that only outputs YYYY-MM-DD in JSON.
type Date time.Time

// MarshalJSON formats a `Date` as YYYY-MM-DD, or null for zero values.
func (d Date) MarshalJSON() ([]byte, error) {
	t := time.Time(d)
	if t.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, t.Format("2006-01-02"))), nil
}

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
	CNPJ                                 string `json:"cnpj"`
	IdentificadorDeSocio                 int    `json:"identificador_de_socio"`
	NomeSocio                            string `json:"nome_socio"`
	CNPJCPFDoSocio                       string `json:"cnpj_cpf_do_socio"`
	CodigoQualificacaoSocio              int    `json:"codigo_qualificacao_socio"`
	PercentualCapitalSocial              int    `json:"percentual_capital_social"`
	DataEntradaSociedade                 Date   `json:"data_entrada_sociedade"`
	CPFRepresentanteLegal                string `json:"cpf_representante_legal"`
	NomeRepresentanteLegal               string `json:"nome_representante_legal"`
	CodigoQualificacaoRepresentanteLegal int    `json:"codigo_qualificacao_representante_legal"`
}

// Company represents a row from the `empresa` database table.
type Company struct {
	CNPJ                      string     `json:"cnpj"`
	IdentificadorMatrizFilial int        `json:"identificador_matriz_filial"`
	RazaoSocial               string     `json:"razao_social"`
	NomeFantasia              string     `json:"nome_fantasia"`
	SituacaoCadastral         int        `json:"situacao_cadastral"`
	DataSituacaoCadastral     Date       `json:"data_situacao_cadastral"`
	MotivoSituacaoCadastral   int        `json:"motivo_situacao_cadastral"`
	NomeCidadeExterior        string     `json:"nome_cidade_exterior"`
	CodigoNaturezaJuridica    int        `json:"codigo_natureza_juridica"`
	DataInicioAtividade       Date       `json:"data_inicio_atividade"`
	CNAEFiscal                int        `json:"cnae_fiscal"`
	CNAEFiscalDescricao       string     `json:"cnae_fiscal_descricao"`
	DescricaoTipoLogradouro   string     `json:"descricao_tipo_logradouro"`
	Logradouro                string     `json:"logradouro"`
	Numero                    string     `json:"numero"`
	Complemento               string     `json:"complemento"`
	Bairro                    string     `json:"bairro"`
	CEP                       string     `json:"cep"`
	UF                        string     `json:"uf"`
	CodigoMunicipio           int        `json:"codigo_municipio"`
	Municipio                 string     `json:"municipio"`
	DDDTelefone1              string     `json:"ddd_telefone_1"`
	DDDTelefone2              string     `json:"ddd_telefone_2"`
	DDDFax                    string     `json:"ddd_fax"`
	QualificacaoDoResponsavel int        `json:"qualificacao_do_responsavel"`
	CapitalSocial             float32    `json:"capital_social"`
	Porte                     int        `json:"porte"`
	OpcaoPeloSimples          bool       `json:"opcao_pelo_simples"`
	DataOpcaoPeloSimples      Date       `json:"data_opcao_pelo_simples"`
	DataExclusaoDoSimples     Date       `json:"data_exclusao_do_simples"`
	OpcaoPeloMEI              bool       `json:"opcao_pelo_mei"`
	SituacaoEspecial          string     `json:"situacao_especial"`
	DataSituacaoEspecial      Date       `json:"data_situacao_especial"`
	QSA                       []*Partner `json:"qsa"`
	CNAESecundarias           []*CNAE    `json:"cnaes_secundarias"`
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
	return cnpj.Mask(c.CNPJ)
}
