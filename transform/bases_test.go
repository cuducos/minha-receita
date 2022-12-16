package transform

import "testing"

var (
	r = []string{
		"BASE DO CNPJ",
		"Razão Social",
		"21",
		"13",
		"4.20",
		"5",
		"Responsável",
	}
	l = lookups{natures: make(map[int]string)}
)

func TestNewBase(t *testing.T) {
	l.natures[21] = "Natureza Jurídica"
	b, err := newBaseData(&l, r)
	if err != nil {
		t.Errorf("expected no error creating base data, got %s", err)
	}
	if *b.CodigoPorte != 5 {
		t.Errorf("expected CodigoPorte to be 5, got %d", b.CodigoPorte)
	}
	if *b.Porte != "DEMAIS" {
		t.Errorf("expected Porte to be DEMAIS, got %s", *b.Porte)
	}
	if b.RazaoSocial != r[1] {
		t.Errorf("expected RazaoSocial to be %s, got %s", r[1], b.RazaoSocial)
	}
	if *b.CodigoNaturezaJuridica != 21 {
		t.Errorf("expected CodigoNaturezaJuridica to be 21, got %d", *b.CodigoNaturezaJuridica)
	}
	if *b.NaturezaJuridica != l.natures[21] {
		t.Errorf("expected NaturezaJuridica to be %s, got %s", l.natures[21], *b.NaturezaJuridica)
	}
	if *b.QualificacaoDoResponsavel != 13 {
		t.Errorf("expected QualificacaoDoResponsavel to be 13, got %d", *b.QualificacaoDoResponsavel)
	}
	if *b.CapitalSocial != 4.2 {
		t.Errorf("expected CapitalSocial to be 4.2, got %f", *b.CapitalSocial)
	}
	if b.EnteFederativoResponsavel != r[6] {
		t.Errorf("expected EnteFederativoResponsavel to be %s, got %s", r[6], b.EnteFederativoResponsavel)
	}
}

func TestLoadBaseRow(t *testing.T) {
	expected := `{"codigo_porte":5,"porte":"DEMAIS","razao_social":"Razão Social","codigo_natureza_juridica":21,"natureza_juridica":"Natureza Jurídica","qualificacao_do_responsavel":13,"capital_social":4.2,"ente_federativo_responsavel":"Responsável"}`
	l.natures[21] = "Natureza Jurídica"
	b, err := loadBaseRow(&l, r)
	if err != nil {
		t.Errorf("expected no error loading base data row, got %s", err)
	}
	if string(b) != expected {
		t.Errorf("expected row to be %s, %s", expected, string(b))
	}
}
