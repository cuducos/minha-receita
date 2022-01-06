package transform

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/cuducos/go-cnpj"
)

var companyNameClenupRegex = regexp.MustCompile(`(\D)(\d{3})(\d{5})(\d{3})$`) // masks CPF from in MEI names

func companyNameClenup(n string) string {
	return strings.TrimSpace(companyNameClenupRegex.ReplaceAllString(n, "$1***$3***"))
}

type company struct {
	CNPJ                             string    `json:"cnpj"`
	IdentificadorMatrizFilial        *int      `json:"identificador_matriz_filial"`
	NomeFantasia                     string    `json:"nome_fantasia"`
	SituacaoCadastral                *int      `json:"situacao_cadastral"`
	DescricaoSituacaoCadastral       *string   `json:"descricao_situacao_cadastral"`
	DataSituacaoCadastral            *date     `json:"data_situacao_cadastral"`
	MotivoSituacaoCadastral          *int      `json:"motivo_situacao_cadastral"`
	DescricaoMotivoSituacaoCadastral *string   `json:"descricao_motivo_situacao_cadastral"`
	NomeCidadeNoExterior             string    `json:"nome_cidade_no_exterior"`
	CodigoPais                       *int      `json:"codigo_pais"`
	Pais                             *string   `json:"pais"`
	DataInicioAtividade              *date     `json:"data_inicio_atividade"`
	CNAEFiscal                       *int      `json:"cnae_fiscal"`
	CNAEFiscalDescricao              *string   `json:"cnae_fiscal_descricao"`
	DescricaoTipoDeLogradouro        string    `json:"descricao_tipo_de_logradouro"`
	Logradouro                       string    `json:"logradouro"`
	Numero                           string    `json:"numero"`
	Complemento                      string    `json:"complemento"`
	Bairro                           string    `json:"bairro"`
	CEP                              string    `json:"cep"`
	UF                               string    `json:"uf"`
	CodigoMunicipio                  *int      `json:"codigio_municipio"`
	Municipio                        *string   `json:"municipio"`
	Telefone1                        string    `json:"ddd_telefone_1"`
	Telefone2                        string    `json:"ddd_telefone_2"`
	Fax                              string    `json:"ddd_fax"`
	SituacaoEspecial                 string    `json:"situacao_especial"`
	DataSituacaoEspecial             *date     `json:"data_situacao_especial"`
	OpcaoPeloSimples                 *bool     `json:"opcao_pelo_simples"`
	DataOpcaoPeloSimples             *date     `json:"data_opcao_pelo_simples"`
	DataExclusaoDoSimples            *date     `json:"data_exclusao_do_simples"`
	OpcaoPeloMEI                     *bool     `json:"opcao_pelo_mei"`
	DataOpcaoPeloMEI                 *date     `json:"data_opcao_pelo_mei"`
	DataExclusaoDoMEI                *date     `json:"data_exclusao_do_mei"`
	RazaoSocial                      string    `json:"razao_social"`
	CodigoNaturezaJuridica           *int      `json:"codigo_natureza_juridica"`
	NaturezaJuridica                 *string   `json:"natureza_juridica"`
	QualificacaoDoResponsavel        *int      `json:"qualificacao_do_responsavel"`
	CapitalSocial                    *float32  `json:"capital_social"`
	CodigoPorte                      *int      `json:"codigo_porte"`
	Porte                            *string   `json:"porte"`
	EnteFederativoResponsavel        *int      `json:"ente_federativo_responsavel"`
	DescricaoPorte                   string    `json:"descricao_porte"`
	QuadroSocietario                 []partner `json:"qsa"`
	CNAESecundarios                  []cnae    `json:"cnaes_secundarios"`
}

func (c *company) situacaoCadastral(v string) error {
	i, err := toInt(v)
	if err != nil {
		return fmt.Errorf("error trying to parse SituacaoCadastral %s: %w", v, err)
	}

	var s string
	switch *i {
	case 1:
		s = "NULA"
	case 2:
		s = "ATIVA"
	case 3:
		s = "SUSPENSA"
	case 4:
		s = "INAPTA"
	case 8:
		s = "BAIXADA"
	}

	c.SituacaoCadastral = i
	if s != "" {
		c.DescricaoSituacaoCadastral = &s
	}
	return nil
}

func newCompany(row []string, l *lookups) (company, error) {
	var c company
	c.CNPJ = row[0] + row[1] + row[2]
	c.NomeFantasia = companyNameClenup(row[4])
	c.NomeCidadeNoExterior = row[8]
	c.DescricaoTipoDeLogradouro = row[13]
	c.Logradouro = row[14]
	c.Numero = row[15]
	c.Complemento = row[16]
	c.Bairro = row[17]
	c.CEP = row[18]
	c.UF = row[19]
	c.Telefone1 = row[21] + row[22]
	c.Telefone2 = row[23] + row[24]
	c.Fax = row[25] + row[26]
	c.SituacaoEspecial = row[28]

	identificadorMatrizFilial, err := toInt(row[3])
	if err != nil {
		return c, fmt.Errorf("error trying to parse IdentificadorMatrizFilial %s: %w", row[3], err)
	}
	c.IdentificadorMatrizFilial = identificadorMatrizFilial

	if err := c.situacaoCadastral(row[5]); err != nil {
		return c, fmt.Errorf("error trying to parse SituacaoCadastral: %w", err)
	}

	dataSituacaoCadastral, err := toDate(row[6])
	if err != nil {
		return c, fmt.Errorf("error trying to parse DataSituacaoCadastral %s: %w", row[3], err)
	}
	c.DataSituacaoCadastral = dataSituacaoCadastral

	if err := c.motivoSituacaoCadastral(l, row[7]); err != nil {
		return c, fmt.Errorf("error trying to parse MotivoSituacaoCadastral: %w", err)
	}

	if err := c.pais(l, row[9]); err != nil {
		return c, fmt.Errorf("error trying to parse CodigoPais: %w", err)
	}

	dataInicioAtividade, err := toDate(row[10])
	if err != nil {
		return c, fmt.Errorf("error trying to parse DataInicioAtividade %s: %w", row[10], err)
	}
	c.DataInicioAtividade = dataInicioAtividade

	if err := c.cnaes(l, row[11], row[12]); err != nil {
		return c, fmt.Errorf("error trying to parse cnae: %w", err)
	}

	if err := c.municipio(l, row[20]); err != nil {
		return c, fmt.Errorf("error trying to parse CodigoMunicipio %s: %w", row[20], err)
	}

	dataSituacaoEspecial, err := toDate(row[29])
	if err != nil {
		return c, fmt.Errorf("error trying to parse DataSituacaoEspecial %s: %w", row[20], err)
	}
	c.DataSituacaoEspecial = dataSituacaoEspecial

	return c, nil
}

func (c *company) Save(db database) error {
	b, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("error while mashaling company JSON: %w", err)
	}
	return db.SaveCompany(c.CNPJ, string(b))
}

func companyFromString(j string) (company, error) {
	var c company
	if err := json.Unmarshal([]byte(j), &c); err != nil {
		return company{}, fmt.Errorf("error unmarshaling: %w", err)
	}
	return c, nil
}

func companyFromDB(db database, n string) (company, error) {
	j, err := db.GetCompany(cnpj.Unmask(n))
	if err != nil {
		return company{}, fmt.Errorf("error loading %s: %w", cnpj.Mask(n), err)
	}
	return companyFromString(j)
}
