package transform

import (
	"path/filepath"
	"testing"
)

func TestAddMain(t *testing.T) {
	d := t.TempDir()
	c := company{CNPJ: "33683111000280"}
	p, err := c.toJSON(d)
	if err != nil {
		t.Errorf("expected to error creating a json, got %s", err)
	}
	for _, src := range []sourceType{partners, motives, cities, countries, cnaes, qualifications, base_cpnj, natures} {
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
	if err := addBaseCPNJ(d, &l); err != nil {
		t.Errorf("expected no errors adding main, got %s", err)
		return
	}
	got, err := companyFromJSON(p)
	codigoPorte := 5
	porte := "DEMAIS"
	codigoNaturezaJuridica := 2011
	qualificacaoDoResponsavel := 16
	capitalSocial, _ := toFloat("1061004800.000000")
	expected := company{
		RazaoSocial:               "SERVICO FEDERAL DE PROCESSAMENTO DE DADOS (SERPRO)",
		CodigoNaturezaJuridica:    &codigoNaturezaJuridica,
		QualificacaoDoResponsavel: &qualificacaoDoResponsavel,
		CapitalSocial:             capitalSocial,
		CodigoPorte:               &codigoPorte,
		Porte:                     &porte,
		EnteFederativoResponsavel: nil,
	}
	if got.RazaoSocial != expected.RazaoSocial {
		t.Errorf("expected RazaoSocial to be %s, got %s", c.RazaoSocial, got.RazaoSocial)
	}
	if *got.CodigoNaturezaJuridica != *expected.CodigoNaturezaJuridica {
		t.Errorf("expected CodigoNaturezaJuridica to be %d, got %d", *expected.CodigoNaturezaJuridica, *got.CodigoNaturezaJuridica)
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
