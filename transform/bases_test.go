package transform

import (
	"testing"

	"github.com/cuducos/go-cnpj"
)

func TestAddBases(t *testing.T) {
	db := newMockDB()
	c := company{CNPJ: "33683111000280"}
	if err := c.Save(&db); err != nil {
		t.Errorf("expected no error saving a company, got %s", err)
	}
	l, err := newLookups(testdata)
	if err != nil {
		t.Errorf("expected no error creating look up tables, got %s", err)
		return
	}
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
		EnteFederativoResponsavel: nil,
	}

	if err := addBases(testdata, &db, &l); err != nil {
		t.Errorf("expected no errors adding base cnpj, got %s", err)
		return
	}
	got, err := companyFromDB(&db, c.CNPJ)
	if err != nil {
		t.Errorf("expected no errors loading company from %s, got %s", cnpj.Mask(c.CNPJ), err)
		return
	}
	if got.RazaoSocial != expected.RazaoSocial {
		t.Errorf("expected RazaoSocial to be %s, got %s", c.RazaoSocial, got.RazaoSocial)
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
		t.Errorf("expected Porte to be %s, got %s", *c.Porte, *got.Porte)
	}
	if got.EnteFederativoResponsavel != nil {
		t.Errorf("expected EnteFederativoResponsavel to be nil, got %d", *got.EnteFederativoResponsavel)
	}
}
