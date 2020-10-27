package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/cuducos/go-cnpj"
	"github.com/go-pg/pg/v10"
)

// Database interface to get a company (returns a JSON string).
type Database interface {
	GetCompany(string) string
}

// Cnae represents a row from the `cnae` database table.
type Cnae struct {
	tableName struct{} `pg:"cnae"`
	Codigo    string   `json:"codigo"`
	Descricao string   `json:"descricao"`
}

// Partner represents a row from the `socio` database table.
type Partner struct {
	tableName                            struct{}  `pg:"socio"`
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
	tableName                 struct{}   `pg:"empresa"`
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
	CnaeFiscal                int        `json:"cnae_fiscal"`
	CnaeFiscalDescricao       string     `pg:"-" json:"cnae_fiscal_descricao"`
	DescricaoTipoLogradouro   string     `json:"descricao_tipo_logradouro"`
	Logradouro                string     `json:"logradouro"`
	Numero                    string     `json:"numero"`
	Complemento               string     `json:"complemento"`
	Bairro                    string     `json:"bairro"`
	Cep                       int        `json:"cep"`
	Uf                        string     `json:"uf"`
	CodigoMunicipio           int        `json:"codigo_municipio"`
	Municipio                 string     `json:"municipio"`
	DddTelefone1              string     `pg:"ddd_telefone_1" json:"ddd_telefone_1"`
	DddTelefone2              string     `pg:"ddd_telefone_2" json:"ddd_telefone_2"`
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
	Qsa                       []*Partner `pg:"-" json:"qsa"`
	CnaesSecundarias          []*Cnae    `pg:"-" json:"cnaes_secundarias"`
}

func (c *Company) queryPartners(db *pg.DB, wg *sync.WaitGroup) {
	defer wg.Done()

	err := db.Model(&c.Qsa).Where("cnpj = ?", c.Cnpj).Select()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not get partners for %s: %v", cnpj.Mask(c.Cnpj), err)
	}
}

func (c *Company) queryActivities(db *pg.DB, wg *sync.WaitGroup) {
	defer wg.Done()

	_, err := db.Query(&c.CnaesSecundarias, `
		SELECT cnae_secundaria.cnae AS codigo, cnae.descricao
		FROM cnae_secundaria
		INNER JOIN cnae ON cnae_secundaria.cnae = cnae.codigo
		WHERE cnpj = ?
	`, c.Cnpj)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not get secondary CNAE for %s: %v", cnpj.Mask(c.Cnpj), err)
	}
}

func (c *Company) queryCnaeDescription(db *pg.DB, wg *sync.WaitGroup) {
	defer wg.Done()

	var cnae Cnae
	err := db.Model(&cnae).Where("codigo = ?", c.CnaeFiscal).Select()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not get secondary CNAE for %s: %v", cnpj.Mask(c.Cnpj), err)
	}
	c.CnaeFiscalDescricao = cnae.Descricao
}

// PostgreSQL database interface.
type PostgreSQL struct {
	db *pg.DB
}

// Close ends the conection with the database.
func (p *PostgreSQL) Close() {
	p.db.Close()
}

// GetCompany returns a string in JSON format, or an empty string if not found.
func (p *PostgreSQL) GetCompany(num string) string {
	var c Company
	err := p.db.Model(&c).Where("cnpj = ?", num).Select()
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(3)
	go c.queryCnaeDescription(p.db, &wg)
	go c.queryPartners(p.db, &wg)
	go c.queryActivities(p.db, &wg)
	wg.Wait()

	b, err := json.Marshal(c)
	if err != nil {
		log.Output(2, fmt.Sprintf("Could not serialize %s: %v", cnpj.Mask(c.Cnpj), err))
		return ""
	}
	return string(b)
}

// NewPostgreSQL creates a new PostgreSQL connection and ping it to make sure it works.
func NewPostgreSQL() PostgreSQL {
	var p PostgreSQL

	opt, err := pg.ParseURL(os.Getenv("POSTGRES_URI"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to parse POSTGRES_URI: %v\n", err)
		os.Exit(1)
	}

	p.db = pg.Connect(opt)
	if err := p.db.Ping(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to PostgreSQL: %v\n", err)
		os.Exit(1)
	}

	return p
}
