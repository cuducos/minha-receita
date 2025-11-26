package transformnext

import (
	"bytes"
	"encoding/json/v2"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

func maskCPF(name string) string {
	if len(name) < 11 {
		return name
	}
	tail := name[len(name)-11:]
	for _, c := range tail {
		if c < '0' || c > '9' {
			return name
		}
	}
	if len(name) > 11 {
		prev := name[len(name)-12]
		if prev >= '0' && prev <= '9' {
			return name
		}
	}
	return name[:len(name)-11] + "***" + tail[3:8] + "***"
}

type CNAE struct {
	Codigo    int    `json:"codigo" bson:"codigo"`
	Descricao string `json:"descricao" bson:"descricao"`
}

type TaxRegime struct {
	Ano                       int     `json:"ano" bson:"ano"`
	CNPJDaSCP                 *string `json:"cnpj_da_scp" bson:"cnpj_da_scp"`
	FormaDeTributação         string  `json:"forma_de_tributacao" bson:"forma_de_tributacao"`
	QuantidadeDeEscrituracoes int     `json:"quantidade_de_escrituracoes" bson:"quantidade_de_escrituracoes"`
}

type Partner struct {
	IdentificadorDeSocio                 *int    `json:"identificador_de_socio" bson:"identificador_de_socio"`
	NomeSocio                            string  `json:"nome_socio" bson:"nome_socio"`
	CNPJCPFDoSocio                       string  `json:"cnpj_cpf_do_socio" bson:"cnpj_cpf_do_socio"`
	CodigoQualificacaoSocio              *int    `json:"codigo_qualificacao_socio" bson:"codigo_qualificacao_socio"`
	QualificaoSocio                      *string `json:"qualificacao_socio" bson:"qualificacao_socio"`
	DataEntradaSociedade                 *date   `json:"data_entrada_sociedade" bson:"data_entrada_sociedade"`
	CodigoPais                           *int    `json:"codigo_pais" bson:"codigo_pais"`
	Pais                                 *string `json:"pais" bson:"pais"`
	CPFRepresentanteLegal                string  `json:"cpf_representante_legal" bson:"cpf_representante_legal"`
	NomeRepresentanteLegal               string  `json:"nome_representante_legal" bson:"nome_representante_legal"`
	CodigoQualificacaoRepresentanteLegal *int    `json:"codigo_qualificacao_representante_legal" bson:"codigo_qualificacao_representante_legal"`
	QualificacaoRepresentanteLegal       *string `json:"qualificacao_representante_legal" bson:"qualificacao_representante_legal"`
	CodigoFaixaEtaria                    *int    `json:"codigo_faixa_etaria" bson:"codigo_faixa_etaria"`
	FaixaEtaria                          *string `json:"faixa_etaria" bson:"faixa_etaria"`
}

type Company struct {
	CNPJ                             string      `json:"cnpj" bson:"cnpj"`
	IdentificadorMatrizFilial        *int        `json:"identificador_matriz_filial" bson:"identificador_matriz_filial"`
	DescricaoMatrizFilial            *string     `json:"descricao_identificador_matriz_filial" bson:"descricao_identificador_matriz_filial"`
	NomeFantasia                     string      `json:"nome_fantasia" bson:"nome_fantasia"`
	SituacaoCadastral                *int        `json:"situacao_cadastral" bson:"situacao_cadastral"`
	DescricaoSituacaoCadastral       *string     `json:"descricao_situacao_cadastral" bson:"descricao_situacao_cadastral"`
	DataSituacaoCadastral            *date       `json:"data_situacao_cadastral" bson:"data_situacao_cadastral"`
	MotivoSituacaoCadastral          *int        `json:"motivo_situacao_cadastral" bson:"motivo_situacao_cadastral"`
	DescricaoMotivoSituacaoCadastral *string     `json:"descricao_motivo_situacao_cadastral" bson:"descricao_motivo_situacao_cadastral"`
	NomeCidadeNoExterior             string      `json:"nome_cidade_no_exterior" bson:"nome_cidade_no_exterior"`
	CodigoPais                       *int        `json:"codigo_pais" bson:"codigo_pais"`
	Pais                             *string     `json:"pais" bson:"pais"`
	DataInicioAtividade              *date       `json:"data_inicio_atividade" bson:"data_inicio_atividade"`
	CNAEFiscal                       *int        `json:"cnae_fiscal" bson:"cnae_fiscal"`
	CNAEFiscalDescricao              *string     `json:"cnae_fiscal_descricao" bson:"cnae_fiscal_descricao"`
	DescricaoTipoDeLogradouro        string      `json:"descricao_tipo_de_logradouro" bson:"descricao_tipo_de_logradouro"`
	Logradouro                       string      `json:"logradouro" bson:"logradouro"`
	Numero                           string      `json:"numero" bson:"numero"`
	Complemento                      string      `json:"complemento" bson:"complemento"`
	Bairro                           string      `json:"bairro" bson:"bairro"`
	CEP                              string      `json:"cep" bson:"cep"`
	UF                               string      `json:"uf" bson:"uf"`
	CodigoMunicipio                  *int        `json:"codigo_municipio" bson:"codigo_municipio"`
	CodigoMunicipioIBGE              *int        `json:"codigo_municipio_ibge" bson:"codigo_municipio_ibge"`
	Municipio                        *string     `json:"municipio" bson:"municipio"`
	Telefone1                        string      `json:"ddd_telefone_1" bson:"ddd_telefone_1"`
	Telefone2                        string      `json:"ddd_telefone_2" bson:"ddd_telefone_2"`
	Fax                              string      `json:"ddd_fax" bson:"ddd_fax"`
	Email                            *string     `json:"email" bson:"email"`
	SituacaoEspecial                 string      `json:"situacao_especial" bson:"situacao_especial"`
	DataSituacaoEspecial             *date       `json:"data_situacao_especial" bson:"data_situacao_especial"`
	OpcaoPeloSimples                 *bool       `json:"opcao_pelo_simples" bson:"opcao_pelo_simples"`
	DataOpcaoPeloSimples             *date       `json:"data_opcao_pelo_simples" bson:"data_opcao_pelo_simples"`
	DataExclusaoDoSimples            *date       `json:"data_exclusao_do_simples" bson:"data_exclusao_do_simples"`
	OpcaoPeloMEI                     *bool       `json:"opcao_pelo_mei" bson:"opcao_pelo_mei"`
	DataOpcaoPeloMEI                 *date       `json:"data_opcao_pelo_mei" bson:"data_opcao_pelo_mei"`
	DataExclusaoDoMEI                *date       `json:"data_exclusao_do_mei" bson:"data_exclusao_do_mei"`
	RazaoSocial                      string      `json:"razao_social" bson:"razao_social"`
	CodigoNaturezaJuridica           *int        `json:"codigo_natureza_juridica" bson:"codigo_natureza_juridica"`
	NaturezaJuridica                 *string     `json:"natureza_juridica" bson:"natureza_juridica"`
	QualificacaoDoResponsavel        *int        `json:"qualificacao_do_responsavel" bson:"qualificacao_do_responsavel"`
	CapitalSocial                    *float32    `json:"capital_social" bson:"capital_social"`
	CodigoPorte                      *int        `json:"codigo_porte" bson:"codigo_porte"`
	Porte                            *string     `json:"porte" bson:"porte"`
	EnteFederativoResponsavel        string      `json:"ente_federativo_responsavel" bson:"ente_federativo_responsavel"`
	QuadroSocietario                 []Partner   `json:"qsa" bson:"qsa"`
	CNAESecundarios                  []CNAE      `json:"cnaes_secundarios" bson:"cnaes_secundarios"`
	RegimeTributario                 []TaxRegime `json:"regime_tributario" bson:"regime_tributario"`
}

func (c *Company) withPrivacy() {
	c.NomeFantasia = strings.TrimSpace(maskCPF(c.NomeFantasia))
	c.Email = nil
	if c.CodigoNaturezaJuridica != nil && strings.Contains(strings.ToLower(*c.NaturezaJuridica), "individual") {
		c.DescricaoTipoDeLogradouro = ""
		c.Logradouro = ""
		c.Numero = ""
		c.Complemento = ""
		c.Telefone1 = ""
		c.Telefone2 = ""
		c.Fax = ""
	}
}

func (c *Company) JSON(p *sync.Pool) (string, error) {
	b := p.Get().(*bytes.Buffer)
	defer func() {
		b.Reset()
		p.Put(b)
	}()
	if err := json.MarshalWrite(b, c); err != nil {
		return "", fmt.Errorf("error while mashaling company JSON: %w", err)
	}
	return b.String(), nil
}

func newCompany(srcs map[string]*source, kv *kv, row []string) (*Company, error) {
	var c Company
	var err error
	var g errgroup.Group
	c.CNPJ = strings.Join(row[:3], "")
	c.IdentificadorMatrizFilial, err = toInt(row[3])
	if err != nil {
		return nil, fmt.Errorf("could not parse IdentificadorMatrizFilial for %s: %w", c.CNPJ, err)
	}
	if err := c.descricaoMatrizFilial(); err != nil {
		return nil, fmt.Errorf("could not parse IdentificadorMatrizFilial for %s: %w", c.CNPJ, err)
	}
	c.NomeFantasia = row[4]
	c.SituacaoCadastral, err = toInt(row[5])
	if err != nil {
		return nil, fmt.Errorf("could not parse SituacaoCadastral for %s: %w", c.CNPJ, err)
	}
	if err := c.descricaoSituacaoCadastral(); err != nil {
		return nil, fmt.Errorf("could not get DescricaoSituacaoCadastral for %s: %w", c.CNPJ, err)
	}
	c.DataSituacaoCadastral, err = toDate(row[6])
	if err != nil {
		return nil, fmt.Errorf("could not parse DataSituacaoCadastral for %s: %w", c.CNPJ, err)
	}
	c.MotivoSituacaoCadastral, err = toInt(row[7])
	if err != nil {
		return nil, fmt.Errorf("could not parse MotivoSituacaoCadastral for %s: %w", c.CNPJ, err)
	}
	g.Go(func() error {
		var err error
		c.DescricaoMotivoSituacaoCadastral, err = stringFromKV(srcs, kv, "mot", row[7], 0)
		if err != nil {
			return fmt.Errorf("could not parse DescricaoMotivoSituacaoCadastral for %s: %w", c.CNPJ, err)
		}
		return nil
	})
	c.NomeCidadeNoExterior = row[8]
	c.CodigoPais, err = toInt(row[9])
	if err != nil {
		return nil, fmt.Errorf("could not parse CodigoPais for %s: %w", c.CNPJ, err)
	}
	g.Go(func() error {
		var err error
		c.Pais, err = stringFromKV(srcs, kv, "pai", row[9], 0)
		if err != nil {
			return fmt.Errorf("could not parse Pais for %s: %w", c.CNPJ, err)
		}
		return nil
	})
	c.DataInicioAtividade, err = toDate(row[10])
	if err != nil {
		return nil, fmt.Errorf("could not parse DataInicioAtividade for %s: %w", c.CNPJ, err)
	}
	c.CNAEFiscal, err = toInt(row[11])
	if err != nil {
		return nil, fmt.Errorf("could not parse CNAEFiscal for %s: %w", c.CNPJ, err)
	}
	g.Go(func() error {
		var err error
		c.CNAEFiscalDescricao, err = stringFromKV(srcs, kv, "cna", row[11], 0)
		if err != nil {
			return fmt.Errorf("could not parse CNAEFiscalDescricao for %s: %w", c.CNPJ, err)
		}
		return nil
	})
	c.DescricaoTipoDeLogradouro = row[13]
	c.Logradouro = row[14]
	c.Numero = row[15]
	c.Complemento = row[16]
	c.Bairro = row[17]
	c.CEP = row[18]
	c.UF = row[19]
	c.CodigoMunicipio, err = toInt(row[20])
	if err != nil {
		return nil, fmt.Errorf("could not parse CodigoMunicipio for %s: %w", c.CNPJ, err)
	}
	g.Go(func() error {
		ibge, err := stringFromKV(srcs, kv, "tab", row[20], 3)
		if err != nil {
			return fmt.Errorf("could not parse CodigoMunicipioIBGE for %s: %w", c.CNPJ, err)
		}
		c.CodigoMunicipioIBGE, err = toInt(*ibge)
		if err != nil {
			return fmt.Errorf("could not parse CodigoMunicipioIBGE number for %s: %w", c.CNPJ, err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		c.Municipio, err = stringFromKV(srcs, kv, "mun", row[20], 3)
		if err != nil {
			return fmt.Errorf("could not parse Municipio for %s: %w", c.CNPJ, err)
		}
		return nil
	})
	c.Telefone1 = row[21] + row[22]
	c.Telefone2 = row[23] + row[24]
	c.Fax = row[25] + row[26]
	c.Email = &row[27]
	c.SituacaoEspecial = row[28]
	c.DataSituacaoEspecial, err = toDate(row[29])
	if err != nil {
		return nil, fmt.Errorf("could not parse DataSituacaoEspecial for %s: %w", c.CNPJ, err)
	}
	g.Go(func() error { return c.base(srcs, kv) })
	g.Go(func() error { return c.simples(srcs, kv) })
	g.Go(func() error { return c.cnaes(srcs, kv, row[12]) })
	g.Go(func() error { return c.partners(srcs, kv) })
	g.Go(func() error { return c.taxes(srcs, kv) })
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return &c, nil
}
