package transform

import (
	"path/filepath"
	"testing"
	"time"
)

func TestAddPartners(t *testing.T) {
	d := t.TempDir()
	c := company{CNPJ: "33683111000280"}
	p, err := c.toJSON(d)
	if err != nil {
		t.Errorf("expected to error creating a json, got %s", err)
	}
	for _, src := range []sourceType{partners, motives, cities, countries, cnaes, qualifications, natures} {
		ls, err := PathsForSource(src, filepath.Join("..", "testdata"))
		if err != nil {
			t.Errorf("expected no error finding paths for %s, got %s", string(src), err)
		}
		for _, f := range ls {
			copyFile(f, d)
		}
	}
	l, err := newLookups(d)
	if err != nil {
		t.Errorf("expected no error creating look up tables, got %s", err)
		return
	}
	if err := addPartners(d, d, &l); err != nil {
		t.Errorf("expected no errors adding partners, got %s", err)
		return
	}
	c, err = companyFromJSON(p)
	if len(c.QuadroSocietario) != 6 {
		t.Errorf("expected 6 partners, got %d", len(c.QuadroSocietario))
	}

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
}
