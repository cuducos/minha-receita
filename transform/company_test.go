package transform

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cuducos/go-cnpj"
	"github.com/cuducos/minha-receita/testutils"
)

func TestNewCompany(t *testing.T) {
	row := []string{
		"33683111",
		"0002",
		"80",
		"2",
		"REGIONAL BRASILIA-DF 11122233344",
		"02",
		"20040522",
		"00",
		"",
		"",
		"19670630",
		"6204000",
		"6201501,6202300,6203100,6209100,6311900",
		"AVENIDA",
		"L2 SGAN",
		"601",
		"MODULO G",
		"ASA NORTE",
		"70836900",
		"DF",
		"9701",
		"",
		"",
		"",
		"",
		"",
		"",
		"serpro@serpro.gov.br",
		"",
		"",
	}

	identificadorMatrizFilial := 2
	DescricaoMatrizFilial := "FILIAL"
	situacaoCadastral := 2
	descricaoSituacaoCadastral := "ATIVA"
	dataSituacaoCadastralAsTime, err := time.Parse(dateInputFormat, row[6])
	if err != nil {
		t.Errorf("error creating DataSituacaoCadastral for expected company: %s", err)
	}
	dataSituacaoCadastral := date(dataSituacaoCadastralAsTime)
	motivoSituacaoCadastral := 0
	descricaoMotivoSituacaoCadastral := "SEM MOTIVO"
	dataInicioAtividadeAsTime, err := time.Parse(dateInputFormat, row[10])
	if err != nil {
		t.Errorf("error creating DataInicioAtividade for expected company: %s", err)
	}
	dataInicioAtividade := date(dataInicioAtividadeAsTime)
	codigoCNAEFiscal := 6204000
	codigoCNAEFiscalDescricao := "Consultoria em tecnologia da informação"
	CodigoMunicipio := 9701
	CodigoMunicipioIBGE := 5300108
	municipio := "BRASILIA"

	expected := Company{
		CNPJ:                             "33683111000280",
		IdentificadorMatrizFilial:        &identificadorMatrizFilial,
		DescricaoMatrizFilial:            &DescricaoMatrizFilial,
		NomeFantasia:                     "REGIONAL BRASILIA-DF ***22233***",
		SituacaoCadastral:                &situacaoCadastral,
		DescricaoSituacaoCadastral:       &descricaoSituacaoCadastral,
		DataSituacaoCadastral:            &dataSituacaoCadastral,
		MotivoSituacaoCadastral:          &motivoSituacaoCadastral,
		DescricaoMotivoSituacaoCadastral: &descricaoMotivoSituacaoCadastral,
		NomeCidadeNoExterior:             "",
		CodigoPais:                       nil,
		Pais:                             nil,
		DataInicioAtividade:              &dataInicioAtividade,
		CNAEFiscal:                       &codigoCNAEFiscal,
		CNAEFiscalDescricao:              &codigoCNAEFiscalDescricao,
		DescricaoTipoDeLogradouro:        "AVENIDA",
		Logradouro:                       "L2 SGAN",
		Numero:                           "601",
		Complemento:                      "MODULO G",
		Bairro:                           "ASA NORTE",
		CEP:                              "70836900",
		UF:                               "DF",
		CodigoMunicipio:                  &CodigoMunicipio,
		CodigoMunicipioIBGE:              &CodigoMunicipioIBGE,
		Municipio:                        &municipio,
		Telefone1:                        "",
		Telefone2:                        "",
		Fax:                              "",
		Email:                            nil,
		SituacaoEspecial:                 "",
		DataSituacaoEspecial:             nil,
		RegimeTributario:                 nil,
		CNAESecundarios: []CNAE{
			{Codigo: 6201501, Descricao: "Desenvolvimento de programas de computador sob encomenda"},
			{Codigo: 6202300, Descricao: "Desenvolvimento e licenciamento de programas de computador customizáveis"},
			{Codigo: 6203100, Descricao: "Desenvolvimento e licenciamento de programas de computador não-customizáveis"},
			{Codigo: 6209100, Descricao: "Suporte técnico, manutenção e outros serviços em tecnologia da informação"},
			{Codigo: 6311900, Descricao: "Tratamento de dados, provedores de serviços de aplicação e serviços de hospedagem na internet"},
		},
	}

	t.Run("with privacy", func(t *testing.T) {
		tmp, err := os.MkdirTemp("", fmt.Sprintf("minha-receita-%s-*", time.Now().Format("20060102150405")))
		if err != nil {
			t.Fatal("error creating temporary key-value storage: %w", err)
		}
		defer os.RemoveAll(tmp)
		kv, err := newBadgerStorage(tmp, false)
		if err != nil {
			t.Errorf("expected no error creating badger, got %s", err)
		}
		defer kv.close()
		lookups, err := newLookups(testdata)
		if err != nil {
			t.Errorf("expected no errors creating look up tables, got %v", err)
		}
		if err := kv.load(testdata, &lookups, 1024); err != nil {
			t.Errorf("expected no error loading values to badger, got %s", err)
		}
		got, err := newCompany(row, &lookups, kv, true)
		if err != nil {
			t.Errorf("expected no errors, got %v", err)
		}
		if got.CNPJ != expected.CNPJ {
			t.Errorf("expected CNPJ to be %s, got %s", expected.CNPJ, got.CNPJ)
		}
		if *got.IdentificadorMatrizFilial != *expected.IdentificadorMatrizFilial {
			t.Errorf(
				"expected IdentificadorMatrizFilial to be %d, got %d",
				*expected.IdentificadorMatrizFilial,
				*got.IdentificadorMatrizFilial,
			)
		}

		if *got.DescricaoMatrizFilial != *expected.DescricaoMatrizFilial {
			t.Errorf(
				"expected DescricaoMatrizFilial to be %s, got %s",
				*expected.DescricaoMatrizFilial,
				*got.DescricaoMatrizFilial,
			)
		}

		if got.NomeFantasia != expected.NomeFantasia {
			t.Errorf("expected NomeFantasia to be %s, got %s", expected.NomeFantasia, got.NomeFantasia)
		}

		if *got.SituacaoCadastral != *expected.SituacaoCadastral {
			t.Errorf("expected SituacaoCadastral to be %d, got %d", *expected.SituacaoCadastral, *got.SituacaoCadastral)
		}

		if *got.DescricaoSituacaoCadastral != *expected.DescricaoSituacaoCadastral {
			t.Errorf(
				"expected DescricaoSituacaoCadastral to be %s, got %s",
				*expected.DescricaoSituacaoCadastral,
				*got.DescricaoSituacaoCadastral,
			)
		}

		if *got.DataSituacaoCadastral != *expected.DataSituacaoCadastral {
			t.Errorf(
				"expected DataSituacaoCadastral to be %s, got %s",
				time.Time(*expected.DataSituacaoCadastral),
				time.Time(*got.DataSituacaoCadastral),
			)
		}

		if *got.MotivoSituacaoCadastral != motivoSituacaoCadastral {
			t.Errorf("expected MotivoSituacaoCadastral to be %d, got %d", motivoSituacaoCadastral, *got.MotivoSituacaoCadastral)
		}

		if *got.DescricaoMotivoSituacaoCadastral != *expected.DescricaoMotivoSituacaoCadastral {
			t.Errorf("expected DescricaoMotivoSituacaoCadastral to be nil, got %s", *got.DescricaoMotivoSituacaoCadastral)
		}

		if *got.CNAEFiscal != codigoCNAEFiscal {
			t.Errorf("expected CNAEFiscal to be %d, got %d", codigoCNAEFiscal, *got.CNAEFiscal)
		}

		if *got.CNAEFiscalDescricao != codigoCNAEFiscalDescricao {
			t.Errorf("expected CNAEFiscalDescricao to be %s, got %s", codigoCNAEFiscalDescricao, *got.CNAEFiscalDescricao)
		}

		if got.NomeCidadeNoExterior != expected.NomeCidadeNoExterior {
			t.Errorf("expected NomeCidadeNoExterior to be %s, got %s", expected.NomeCidadeNoExterior, got.NomeCidadeNoExterior)
		}

		if got.CodigoPais != nil {
			t.Errorf("expected CodigoPais to be nil, got %d", *got.CodigoPais)
		}

		if got.Pais != nil {
			t.Errorf("expected Pais to be nil, got %s", *got.Pais)
		}

		if *got.DataInicioAtividade != *expected.DataInicioAtividade {
			t.Errorf(
				"expected DataInicioAtividade to be %s, got %s",
				time.Time(*expected.DataInicioAtividade),
				time.Time(*got.DataInicioAtividade),
			)
		}

		if got.DescricaoTipoDeLogradouro != expected.DescricaoTipoDeLogradouro {
			t.Errorf("expected DescricaoTipoDeLogradouro to be %s, got %s", expected.DescricaoTipoDeLogradouro, got.DescricaoTipoDeLogradouro)
		}

		if got.Logradouro != expected.Logradouro {
			t.Errorf("expected Logradouro to be %s, got %s", expected.Logradouro, got.Logradouro)
		}

		if got.Numero != expected.Numero {
			t.Errorf("expected Numero to be %s, got %s", expected.Numero, got.Numero)
		}

		if got.Complemento != expected.Complemento {
			t.Errorf("expected Complemento to be %s, got %s", expected.Complemento, got.Complemento)
		}

		if got.Bairro != expected.Bairro {
			t.Errorf("expected Bairro to be %s, got %s", expected.Bairro, got.Bairro)
		}

		if *got.CodigoMunicipio != *expected.CodigoMunicipio {
			t.Errorf("expected CodigoMunicipio to be %d, got %d", *expected.CodigoMunicipio, *got.CodigoMunicipio)
		}

		if *got.Municipio != *expected.Municipio {
			t.Errorf("expected Municipio to be %s, got %s", *expected.Municipio, *got.Municipio)
		}

		if *got.CodigoMunicipioIBGE != *expected.CodigoMunicipioIBGE {
			t.Errorf(
				"expected CodigoMunicipioIBGE to be %d, got %d",
				*expected.CodigoMunicipioIBGE,
				*got.CodigoMunicipioIBGE,
			)
		}

		if got.CEP != expected.CEP {
			t.Errorf("expected CEP to be %s, got %s", expected.CEP, got.CEP)
		}

		if got.UF != expected.UF {
			t.Errorf("expected UF to be %s, got %s", expected.UF, got.UF)
		}

		for i, v := range got.CNAESecundarios {
			if v.Codigo != expected.CNAESecundarios[i].Codigo {
				t.Errorf("expected CNAESecundarios[%d].Codigo to be %d, got %d", i, expected.CNAESecundarios[i].Codigo, v.Codigo)
			}

			if v.Descricao != expected.CNAESecundarios[i].Descricao {
				t.Errorf("expected CNAESecundarios[%d].Descricao to be %s, got %s", i, expected.CNAESecundarios[i].Descricao, v.Descricao)
			}

		}
	})
	t.Run("without privacy", func(t *testing.T) {
		tmp, err := os.MkdirTemp("", fmt.Sprintf("minha-receita-%s-*", time.Now().Format("20060102150405")))
		if err != nil {
			t.Fatal("error creating temporary key-value storage: %w", err)
		}
		defer os.RemoveAll(tmp)
		kv, err := newBadgerStorage(tmp, false)
		if err != nil {
			t.Errorf("expected no error creating badger, got %s", err)
		}
		defer kv.close()
		lookups, err := newLookups(testdata)
		if err != nil {
			t.Errorf("expected no errors creating look up tables, got %v", err)
		}
		if err := kv.load(testdata, &lookups, 1024); err != nil {
			t.Errorf("expected no error loading values to badger, got %s", err)
		}
		email := "serpro@serpro.gov.br"
		expected.Email = &email
		expected.NomeFantasia = "REGIONAL BRASILIA-DF 11122233344"
		got, err := newCompany(row, &lookups, kv, false)
		if err != nil {
			t.Errorf("expected no errors, got %v", err)
		}
		if *got.Email != email {
			t.Errorf("expected Email to be %s, got %s", email, *got.Email)
		}
	})
}

func TestCompanyJSON(t *testing.T) {
	dataInicioAtividadeAsTime, err := time.Parse(dateInputFormat, "19670630")
	if err != nil {
		t.Errorf("error creating DataInicioAtividade for expected company: %s", err)
	}
	dataInicioAtividade := date(dataInicioAtividadeAsTime)
	c := Company{
		CNPJ:                 "33683111000280",
		DataInicioAtividade:  &dataInicioAtividade,
		DataSituacaoEspecial: nil,
	}

	got, err := c.JSON()
	if err != nil {
		t.Errorf("expected no error getting the company %s as json, got %s", cnpj.Mask(c.CNPJ), err)
	}
	if !strings.Contains(got, `"cnpj":"33683111000280"`) {
		t.Errorf("expected to find %s in a CNPJ field in %s", c.CNPJ, got)
	}
	if !strings.Contains(got, `"1967-06-30"`) {
		t.Errorf("expected to find 1967-06-30 in JSON %s", got)
	}
	if !strings.Contains(got, `"data_situacao_especial":null`) {
		t.Errorf("expected to find null for data_situacao_especial in JSON %s", got)
	}
}

func TestCompanyJSONFields(t *testing.T) {
	got := CompanyJSONFields()
	exp := []string{
		"bairro",
		"capital_social",
		"cep",
		"cnae_fiscal",
		"cnae_fiscal_descricao",
		"cnaes_secundarios.codigo",
		"cnaes_secundarios.descricao",
		"cnpj",
		"codigo_municipio",
		"codigo_municipio_ibge",
		"codigo_natureza_juridica",
		"codigo_pais",
		"codigo_porte",
		"complemento",
		"data_exclusao_do_mei",
		"data_exclusao_do_simples",
		"data_inicio_atividade",
		"data_opcao_pelo_mei",
		"data_opcao_pelo_simples",
		"data_situacao_cadastral",
		"data_situacao_especial",
		"ddd_fax",
		"ddd_telefone_1",
		"ddd_telefone_2",
		"descricao_identificador_matriz_filial",
		"descricao_motivo_situacao_cadastral",
		"descricao_situacao_cadastral",
		"descricao_tipo_de_logradouro",
		"email",
		"ente_federativo_responsavel",
		"identificador_matriz_filial",
		"logradouro",
		"motivo_situacao_cadastral",
		"municipio",
		"natureza_juridica",
		"nome_cidade_no_exterior",
		"nome_fantasia",
		"numero",
		"opcao_pelo_mei",
		"opcao_pelo_simples",
		"pais",
		"porte",
		"qsa.cnpj_cpf_do_socio",
		"qsa.codigo_faixa_etaria",
		"qsa.codigo_pais",
		"qsa.codigo_qualificacao_representante_legal",
		"qsa.codigo_qualificacao_socio",
		"qsa.cpf_representante_legal",
		"qsa.data_entrada_sociedade",
		"qsa.faixa_etaria",
		"qsa.identificador_de_socio",
		"qsa.nome_representante_legal",
		"qsa.nome_socio",
		"qsa.pais",
		"qsa.qualificacao_representante_legal",
		"qsa.qualificacao_socio",
		"qualificacao_do_responsavel",
		"razao_social",
		"regime_tributario.ano",
		"regime_tributario.cnpj_da_scp",
		"regime_tributario.forma_de_tributacao",
		"regime_tributario.quantidade_de_escrituracoes",
		"situacao_cadastral",
		"situacao_especial",
		"uf",
	}
	testutils.AssertArraysHaveSameItems(t, got, exp)
}
