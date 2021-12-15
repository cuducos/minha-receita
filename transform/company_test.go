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
	CodigoMunicipio := 9701

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
		Pais:                             "",
		DataInicioAtividade:              &dataInicioAtividade,
		CNAEFiscal:                       &codigoCNAEFiscal,
		DescricaoTipoDeLogradouro:        "AVENIDA",
		Logradouro:                       "L2 SGAN",
		Numero:                           "601",
		Complemento:                      "MODULO G",
		Bairro:                           "ASA NORTE",
		CEP:                              "70836900",
		UF:                               "DF",
		CodigoMunicipio:                  &CodigoMunicipio,
		Municipio:                        "",
		Telefone1:                        "",
		Telefone2:                        "",
		Fax:                              "",
		SituacaoEspecial:                 "",
		DataSituacaoEspecial:             nil,
		CNAESecundarios: []CNAE{
			{Codigo: 6201501},
			{Codigo: 6202300},
			{Codigo: 6203100},
			{Codigo: 6209100},
			{Codigo: 6311900},
		},
	}

	z, err := newArchivedCSV(filepath.Join("..", "testdata", "F.K03200$Z.D11009.MOTICSV.zip"), separator)
	if err != nil {
		t.Errorf("error creating archivedCSV for: %s", err)
	}
	defer z.close()

	var lookups lookups
	lookups.motives, err = z.toLookup()
	if err != nil {
		t.Errorf("error creating motives lookup table: %s", err)
	}

	got, err := newCompany(row, lookups)
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

	if got.NomeCidadeNoExterior != expected.NomeCidadeNoExterior {
		t.Errorf("expected NomeCidadeNoExterior to be %s, got %s", expected.NomeCidadeNoExterior, got.NomeCidadeNoExterior)
	}

	if got.Pais != expected.Pais {
		t.Errorf("expected Pais to be %s, got %s", expected.Pais, got.Pais)
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

	if got.UF != expected.UF {
		t.Errorf("expected UF to be %s, got %s", expected.UF, got.UF)
	}

	if got.CNAESecundarios[0].Codigo != 6201501 {
		t.Errorf("expected CNAESecundarios[0] to be 6201501, got %d", got.CNAESecundarios[0].Codigo)
	}
	if got.CNAESecundarios[1].Codigo != 6202300 {
		t.Errorf("expected CNAESecundarios[1] to be 6202300, got %d", got.CNAESecundarios[1].Codigo)
	}
	if got.CNAESecundarios[2].Codigo != 6203100 {
		t.Errorf("expected CNAESecundarios[2] to be 6203100, got %d", got.CNAESecundarios[2].Codigo)
	}
	if got.CNAESecundarios[3].Codigo != 6209100 {
		t.Errorf("expected CNAESecundarios[3] to be 6209100, got %d", got.CNAESecundarios[3].Codigo)
	}
	if got.CNAESecundarios[4].Codigo != 6311900 {
		t.Errorf("expected CNAESecundarios[4] to be 6311900, got %d", got.CNAESecundarios[4].Codigo)
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
