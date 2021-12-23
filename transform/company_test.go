package transform

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cuducos/go-cnpj"
)

func TestNewCompany(t *testing.T) {
	row := []string{
		"33683111",
		"0002",
		"80",
		"2",
		"REGIONAL BRASILIA-DF",
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
		"",
		"",
		"",
	}

	identificadorMatrizFilial := 2
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
	municipio := "BRASILIA"

	expected := company{
		CNPJ:                             "33683111000280",
		IdentificadorMatrizFilial:        &identificadorMatrizFilial,
		NomeFantasia:                     "REGIONAL BRASILIA-DF",
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
		Municipio:                        &municipio,
		Telefone1:                        "",
		Telefone2:                        "",
		Fax:                              "",
		SituacaoEspecial:                 "",
		DataSituacaoEspecial:             nil,
		CNAESecundarios: []cnae{
			{Codigo: 6201501, Descricao: "Desenvolvimento de programas de computador sob encomenda"},
			{Codigo: 6202300, Descricao: "Desenvolvimento e licenciamento de programas de computador customizáveis"},
			{Codigo: 6203100, Descricao: "Desenvolvimento e licenciamento de programas de computador não-customizáveis"},
			{Codigo: 6209100, Descricao: "Suporte técnico, manutenção e outros serviços em tecnologia da informação"},
			{Codigo: 6311900, Descricao: "Tratamento de dados, provedores de serviços de aplicação e serviços de hospedagem na internet"},
		},
	}

	lookups, err := newLookups(filepath.Join("..", "testdata"))
	if err != nil {
		t.Errorf("expected no errors creating look up tables, got %v", err)
	}

	got, err := newCompany(row, &lookups)
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

	if got.CEP != expected.CEP {
		t.Errorf("expected CEP to be %s, got %s", expected.CEP, got.CEP)
	}

	if *got.CodigoMunicipio != *expected.CodigoMunicipio {
		t.Errorf("expected CodigoMunicipio to be %d, got %d", *expected.CodigoMunicipio, *got.CodigoMunicipio)
	}

	if *got.Municipio != *expected.Municipio {
		t.Errorf("expected Municipio to be %s, got %s", *expected.Municipio, *got.Municipio)
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
}

func TestCompanyToJson(t *testing.T) {
	d := t.TempDir()

	dataInicioAtividadeAsTime, err := time.Parse(dateInputFormat, "19670630")
	if err != nil {
		t.Errorf("error creating DataInicioAtividade for expected company: %s", err)
	}
	dataInicioAtividade := date(dataInicioAtividadeAsTime)
	c := company{
		CNPJ:                 "33683111000280",
		DataInicioAtividade:  &dataInicioAtividade,
		DataSituacaoEspecial: nil,
	}

	p, err := c.toJSON(d)
	if err != nil {
		t.Errorf("expected no error in converting %s to JSON, got %s", cnpj.Mask(c.CNPJ), err)
	}

	b, err := ioutil.ReadFile(p)
	if err != nil {
		t.Errorf("expected no error reading %s, got %s", p, err)
	}

	got := string(b)
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

func TestCompanyFromJson(t *testing.T) {
	d := t.TempDir()
	c := company{CNPJ: "33683111000280", QuadroSocietario: []partner{{CNPJCPFDoSocio: "42"}}}
	p, err := c.toJSON(d)
	got, err := companyFromJSON(p)
	if err != nil {
		t.Errorf("expected no error creating a company from json, got %s", err)
	}
	if got.CNPJ != c.CNPJ {
		t.Errorf("expected cnpj to be %s, got %s", c.CNPJ, got.CNPJ)
	}
	if got.QuadroSocietario[0].CNPJCPFDoSocio != c.QuadroSocietario[0].CNPJCPFDoSocio {
		t.Errorf("expected cnpj/cpf to be %s, got %s",
			c.QuadroSocietario[0].CNPJCPFDoSocio,
			got.QuadroSocietario[0].CNPJCPFDoSocio,
		)
	}
}
