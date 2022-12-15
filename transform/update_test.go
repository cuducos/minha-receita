package transform

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"
)

func TestUpdateTaskRun(t *testing.T) {
	t.Run("base CNPJ", func(t *testing.T) {
		got := setupUpdateTaskTest(t)
		codigoPorte := 5
		porte := "DEMAIS"
		codigoNaturezaJuridica := 2011
		qualificacaoDoResponsavel := 16
		naturezaJuridica := "Empresa Pública"
		capitalSocial, _ := toFloat("1061004800.000000")
		expected := company{
			RazaoSocial:               "SERVICO FEDERAL DE PROCESSAMENTO DE DADOS (SERPRO)",
			CodigoNaturezaJuridica:    &codigoNaturezaJuridica,
			NaturezaJuridica:          &naturezaJuridica,
			QualificacaoDoResponsavel: &qualificacaoDoResponsavel,
			CapitalSocial:             capitalSocial,
			CodigoPorte:               &codigoPorte,
			Porte:                     &porte,
			EnteFederativoResponsavel: "",
		}
		if got.RazaoSocial != expected.RazaoSocial {
			t.Errorf("expected RazaoSocial to be %s, got %s", expected.RazaoSocial, got.RazaoSocial)
		}
		if *got.CodigoNaturezaJuridica != *expected.CodigoNaturezaJuridica {
			t.Errorf("expected CodigoNaturezaJuridica to be %d, got %d", *expected.CodigoNaturezaJuridica, *got.CodigoNaturezaJuridica)
		}
		if *got.NaturezaJuridica != *expected.NaturezaJuridica {
			t.Errorf("expected NaturezaJuridica to be %s, got %s", *expected.NaturezaJuridica, *got.NaturezaJuridica)
		}
		if *got.QualificacaoDoResponsavel != *expected.QualificacaoDoResponsavel {
			t.Errorf("expected QualificacaoDoResponsavel to be %d, got %d", *expected.QualificacaoDoResponsavel, *got.QualificacaoDoResponsavel)
		}
		if *got.CapitalSocial != *expected.CapitalSocial {
			t.Errorf("expected CapitalSocial to be %f, got %f", *expected.CapitalSocial, *got.CapitalSocial)
		}
		if *got.CodigoPorte != *expected.CodigoPorte {
			t.Errorf("expected CodigoPorte to be %d, got %d", *expected.CodigoPorte, *got.CodigoPorte)
		}
		if *got.Porte != *expected.Porte {
			t.Errorf("expected Porte to be %s, got %s", *expected.Porte, *got.Porte)
		}
		if got.EnteFederativoResponsavel != "" {
			t.Errorf("expected EnteFederativoResponsavel to be empty, got %s", got.EnteFederativoResponsavel)
		}
	})

	t.Run("partners", func(t *testing.T) {
		identificadorDeSocio := 2
		codigoQualificacaoSocio := 10
		qualificaoSocio := "Diretor"
		dataEntradaSociedade, err := toDate("20161208")
		if err != nil {
			t.Errorf("expected no error creating expedted date, got %s", err)
		}
		codigoQualificacaoRepresentanteLegal := 0
		codigoFaixaEtaria := 7
		faixaEtaria := "Entre 61 a 70 anos"
		expected := partner{
			IdentificadorDeSocio:                 &identificadorDeSocio,
			NomeSocio:                            "ANTONIO DE PADUA FERREIRA PASSOS",
			CNPJCPFDoSocio:                       "***595901**",
			CodigoQualificacaoSocio:              &codigoQualificacaoSocio,
			QualificaoSocio:                      &qualificaoSocio,
			DataEntradaSociedade:                 dataEntradaSociedade,
			CodigoPais:                           nil,
			Pais:                                 nil,
			CPFRepresentanteLegal:                "***000000**",
			NomeRepresentanteLegal:               "",
			CodigoQualificacaoRepresentanteLegal: &codigoQualificacaoRepresentanteLegal,
			QualificacaoRepresentanteLegal:       nil,
			CodigoFaixaEtaria:                    &codigoFaixaEtaria,
			FaixaEtaria:                          &faixaEtaria,
		}

		c := setupUpdateTaskTest(t)
		var got partner
		for _, s := range c.QuadroSocietario {
			if s.NomeSocio == expected.NomeSocio {
				got = s
				break
			}
		}
		if *got.IdentificadorDeSocio != *expected.IdentificadorDeSocio {
			t.Errorf(
				"expected IdentificadorDeSocio to be %d, got %d",
				*expected.IdentificadorDeSocio,
				*got.IdentificadorDeSocio,
			)
		}
		if got.NomeSocio != expected.NomeSocio {
			t.Errorf("expected NomeSocio to be %s, got %s", expected.NomeSocio, got.NomeSocio)
		}
		if got.CNPJCPFDoSocio != expected.CNPJCPFDoSocio {
			t.Errorf("expected CNPJCPFDoSocio to be %s, got %s", expected.CNPJCPFDoSocio, got.CNPJCPFDoSocio)
		}
		if *got.CodigoQualificacaoSocio != *expected.CodigoQualificacaoSocio {
			t.Errorf(
				"expected CodigoQualificacaoSocio to be %d, got %d",
				*expected.CodigoQualificacaoSocio,
				*got.CodigoQualificacaoSocio,
			)
		}
		if *got.QualificaoSocio != *expected.QualificaoSocio {
			t.Errorf("expected QualificaoSocio to be %s, got %s", *expected.QualificaoSocio, *got.QualificaoSocio)
		}
		if *got.DataEntradaSociedade != *expected.DataEntradaSociedade {
			t.Errorf(
				"expected DataEntradaSociedade to be %s, got %s",
				time.Time(*expected.DataEntradaSociedade),
				time.Time(*got.DataEntradaSociedade),
			)
		}
		if got.CodigoPais != nil {
			t.Errorf("expected CodigoPais to be nil, got %d", *got.CodigoPais)
		}
		if got.Pais != nil {
			t.Errorf("expected Pais to be nil, got %s", *got.Pais)
		}
		if got.CPFRepresentanteLegal != expected.CPFRepresentanteLegal {
			t.Errorf(
				"expected CPFRepresentanteLegal to be %s, got %s",
				expected.CPFRepresentanteLegal,
				got.CPFRepresentanteLegal,
			)
		}
		if got.NomeRepresentanteLegal != expected.NomeRepresentanteLegal {
			t.Errorf(
				"expected NomeRepresentanteLegal to be %s, got %s",
				expected.NomeRepresentanteLegal,
				got.NomeRepresentanteLegal,
			)
		}
		if *got.CodigoQualificacaoRepresentanteLegal != *expected.CodigoQualificacaoRepresentanteLegal {
			t.Errorf(
				"expected CodigoQualificacaoRepresentanteLegal to be %d, got %d",
				*expected.CodigoQualificacaoRepresentanteLegal,
				*got.CodigoQualificacaoRepresentanteLegal,
			)
		}
		if got.QualificacaoRepresentanteLegal != nil {
			t.Errorf(
				"expected QualificacaoRepresentanteLegal to be nil, got %s",
				*got.QualificacaoRepresentanteLegal,
			)
		}
		if *got.CodigoFaixaEtaria != *expected.CodigoFaixaEtaria {
			t.Errorf(
				"expected CodigoFaixaetaria to be %d, got %d",
				*expected.CodigoFaixaEtaria,
				*got.CodigoFaixaEtaria,
			)
		}
		if *got.FaixaEtaria != *expected.FaixaEtaria {
			t.Errorf("expected Faixaetaria to be %s, got %s", *expected.FaixaEtaria, *got.FaixaEtaria)
		}
	})

	t.Run("taxes", func(t *testing.T) {
		got := setupUpdateTaskTest(t)
		dataOpcaoPeloSimples, err := time.Parse("2006-01-02", "2014-01-01")
		if err != nil {
			t.Errorf("expected no errors creating date got %s", err)
		}

		if *got.OpcaoPeloSimples != true {
			t.Errorf("expected OpcaoPeloSimples to be true, got %v", got.OpcaoPeloSimples)
		}
		if *got.DataOpcaoPeloSimples != date(dataOpcaoPeloSimples) {
			t.Errorf("expected DataOpcaoPeloSimples to be %s, got %s",
				dataOpcaoPeloSimples,
				time.Time(*got.DataOpcaoPeloSimples),
			)
		}
		if got.DataExclusaoDoSimples != nil {
			t.Errorf("expected DataExclusaoDoSimples to be nil, got %s", time.Time(*got.DataExclusaoDoSimples))
		}
		if *got.OpcaoPeloMEI != false {
			t.Errorf("expected OpcaoPeloMEI to be false, got %v", got.OpcaoPeloMEI)
		}
		if got.DataOpcaoPeloMEI != nil {
			t.Errorf("expected DataOpcaoPeloMEI to be nil, got %s", time.Time(*got.DataOpcaoPeloMEI))
		}
		if got.DataExclusaoDoMEI != nil {
			t.Errorf("expected DataExclusaoDoMEI to be nil, got %s", time.Time(*got.DataExclusaoDoMEI))
		}
	})
}

func setupUpdateTaskTest(t *testing.T) company {
	db := newTestDB(t)
	c := company{CNPJ: "33683111000280"}
	j, err := c.JSON()
	if err != nil {
		t.Errorf("expected no error converting company struct to json, got %s", err)
	}
	n, err := strconv.Atoi(c.CNPJ)
	if err != nil {
		t.Errorf("expected no error converting cnpj to int, got %s", err)
	}
	if err := db.CreateCompanies([][]any{{n, j}}); err != nil {
		t.Errorf("expected no error saving a company, got %s", err)
	}
	l, err := newLookups(testdata)
	if err != nil {
		t.Errorf("expected no error creating look up tables, got %s", err)
	}
	u, err := newUpdateTask(testdata, db, 1, &l)
	if err != nil {
		t.Errorf("expected no errors creating update task, got %s", err)
	}
	if err = u.run(); err != nil {
		t.Errorf("expected no errors running update task, got %s", err)
	}
	j, err = db.GetCompany(c.CNPJ)
	if err != nil {
		t.Errorf("expected no errors loading company, got %s", err)
	}
	if err = json.Unmarshal([]byte(j), &c); err != nil {
		t.Errorf("expected no errors transforming company to struct, got %s", err)
	}
	return c
}
