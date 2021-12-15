package transform

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type CNAE struct {
	Codigo    int    `json:"codigo"`
	Descricao string `json:"descricao"`
}

// TODO this will be used further, it is here just to document the expected output ATM
type Socio struct {
	IdentificadorDeSocio                 int     `json:"identificador_de_socio"`
	NomeSocio                            string  `json:"nome_socio"`
	CNPJCPFDoSocio                       string  `json:"cnpj_cpf_do_socio"`
	CodigoQualificacaoSocio              int     `json:"codigo_qualificacao_socio"`
	PercentualCapitalSocial              float32 `json:"percentual_capital_social"`
	DataEntradaSociedade                 *date   `json:"data_entrada_sociedade"`
	CPFRepresentanteLegal                string  `json:"cpf_representante_legal"`
	NomeRepresentanteLegal               string  `json:"nome_representante_legal"`
	CodigoQualificacaoRepresentanteLegal *int    `json:"codigo_qualificacao_representante_legal"`
}

type company struct {
	// Fields from the venues dataset
	CNPJ                             string  `json:"cnpj"`
	IdentificadorMatrizFilial        *int    `json:"identificador_matriz_filial"`
	NomeFantasia                     string  `json:"nome_fantasia"`
	SituacaoCadastral                *int    `json:"situacao_cadastral"`
	DescricaoSituacaoCadastral       *string `json:"descricao_situacao_cadastral"`
	DataSituacaoCadastral            *date   `json:"data_situacao_cadastral"`
	MotivoSituacaoCadastral          *int    `json:"motivo_situacao_cadastral"`
	DescricaoMotivoSituacaoCadastral *string `json:"descricao_motivo_situacao_cadastral"`
	NomeCidadeNoExterior             string  `json:"nome_cidade_no_exterior"`
	Pais                             string  `json:"pais"`
	DataInicioAtividade              *date   `json:"data_inicio_atividade"`
	CNAEFiscal                       *int    `json:"cnae_fiscal"`
	CNAEFiscalDescricao              string  `json:"cnae_fiscal_descricao"`
	DescricaoTipoDeLogradouro        string  `json:"descricao_tipo_de_logradouro"`
	Logradouro                       string  `json:"logradouro"`
	Numero                           string  `json:"numero"`
	Complemento                      string  `json:"complemento"`
	Bairro                           string  `json:"bairro"`
	CEP                              string  `json:"cep"`
	UF                               string  `json:"uf"`
	CodigoMunicipio                  *int    `json:"codigio_municipio"`
	Municipio                        string  `json:"municipio"`
	Telefone1                        string  `json:"ddd_telefone_1"`
	Telefone2                        string  `json:"ddd_telefone_2"`
	Fax                              string  `json:"ddd_fax"`
	SituacaoEspecial                 string  `json:"situacao_especial"`
	DataSituacaoEspecial             *date   `json:"data_situacao_especial"`
	CNAESecundarios                  []CNAE  `json:"cnaes_secundarios"`

	// TODO this will be used further, it is here just to document the expected output ATM
	// RazaoSocial               string  `json:"razao_social"`
	// CodigoNaturezaJuridica    *int     `json:"codigo_natureza_juridica"`
	// QualificacaoDoResponsavel *int     `json:"qualificacao_do_responsavel"`
	// CapitalSocial             *float32 `json:"capital_social"`
	// Porte                     *int     `json:"porte"`
	// EnteFederativoResponsavel *int     `json:"ente_federativo_responsavel"`

	// TODO backward compatibility
	// DescricaoPorte             string    `json:"descricao_porte"`
	// OpcaoPeloSimples           *bool      `json:"opcao_pelo_mei"`
	// DataOpcaoPeloSimples       *date `json:"data_opcao_pelo_simples"`
	// DataExclusaoDoSimples      *date `json:"data_exclusao_do_simples"`
	// OpcaoPeloMei               *bool      `json:"opcao_pelo_mei"`
	// QuadroSocietario           []Socio    `json:"qsa"`
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

func (c *company) cnaeSecundarios(v string) error {
	for _, n := range strings.Split(v, ",") {
		i, err := strconv.Atoi(n)
		if err != nil {
			return fmt.Errorf("error converting %s to int: %w", v, err)
		}
		c.CNAESecundarios = append(c.CNAESecundarios, CNAE{Codigo: i})
	}
	return nil
}

func newCompany(row []string, l lookups) (company, error) {
	var c company
	c.CNPJ = row[0] + row[1] + row[2]
	c.NomeFantasia = row[4]
	c.NomeCidadeNoExterior = row[8]
	c.Pais = row[9]
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

	dataInicioAtividade, err := toDate(row[10])
	if err != nil {
		return c, fmt.Errorf("error trying to parse DataInicioAtividade %s: %w", row[10], err)
	}
	c.DataInicioAtividade = dataInicioAtividade

	codigoCNAEFiscal, err := toInt(row[11])
	if err != nil {
		return c, fmt.Errorf("error trying to parse CNAEFiscal %s: %w", row[11], err)
	}
	c.CNAEFiscal = codigoCNAEFiscal

	codigoMunicipio, err := toInt(row[20])
	if err != nil {
		return c, fmt.Errorf("error trying to parse CodigoMunicipio %s: %w", row[20], err)
	}
	c.CodigoMunicipio = codigoMunicipio

	dataSituacaoEspecial, err := toDate(row[29])
	if err != nil {
		return c, fmt.Errorf("error trying to parse DataSituacaoEspecial %s: %w", row[20], err)
	}
	c.DataSituacaoEspecial = dataSituacaoEspecial

	if err = c.cnaeSecundarios(row[12]); err != nil {
		return c, fmt.Errorf("error parsing to parse CNAESecundarios %s: %w", row[12], err)
	}

	return c, nil
}

func (c *company) toJSON(outDir string) (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("error while mashaling company JSON: %w", err)
	}
	n, err := PathForCNPJ(c.CNPJ)
	if err != nil {
		return "", fmt.Errorf("error while getting the file path for %s: %w", c.CNPJ, err)
	}

	p := filepath.Join(outDir, n)
	err = os.MkdirAll(filepath.Dir(p), 0755)
	if err != nil {
		return "", fmt.Errorf("error creating %s: %w", filepath.Dir(p), err)
	}

	if err := ioutil.WriteFile(p, b, 0644); err != nil {
		return "", fmt.Errorf("error writing to %s: %w", p, err)
	}
	return p, nil
}
