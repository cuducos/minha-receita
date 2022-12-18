package transform

import "testing"

var (
	baseCSVRow = []string{
		"BASE DO CNPJ",
		"Razão Social",
		"21",
		"13",
		"4.20",
		"5",
		"Responsável",
	}
	baseLookups = lookups{natures: make(map[int]string)}
)

func TestNewBase(t *testing.T) {
	baseLookups.natures[21] = "Natureza Jurídica"
	b, err := newBaseData(&baseLookups, baseCSVRow)
	if err != nil {
		t.Errorf("expected no error creating base data, got %s", err)
	}
	if *b.CodigoPorte != 5 {
		t.Errorf("expected CodigoPorte to be 5, got %d", b.CodigoPorte)
	}
	if *b.Porte != "DEMAIS" {
		t.Errorf("expected Porte to be DEMAIS, got %s", *b.Porte)
	}
	if b.RazaoSocial != baseCSVRow[1] {
		t.Errorf("expected RazaoSocial to be %s, got %s", baseCSVRow[1], b.RazaoSocial)
	}
	if *b.CodigoNaturezaJuridica != 21 {
		t.Errorf("expected CodigoNaturezaJuridica to be 21, got %d", *b.CodigoNaturezaJuridica)
	}
	if *b.NaturezaJuridica != baseLookups.natures[21] {
		t.Errorf("expected NaturezaJuridica to be %s, got %s", baseLookups.natures[21], *b.NaturezaJuridica)
	}
	if *b.QualificacaoDoResponsavel != 13 {
		t.Errorf("expected QualificacaoDoResponsavel to be 13, got %d", *b.QualificacaoDoResponsavel)
	}
	if *b.CapitalSocial != 4.2 {
		t.Errorf("expected CapitalSocial to be 4.2, got %f", *b.CapitalSocial)
	}
	if b.EnteFederativoResponsavel != baseCSVRow[6] {
		t.Errorf("expected EnteFederativoResponsavel to be %s, got %s", baseCSVRow[6], b.EnteFederativoResponsavel)
	}
}

func TestLoadBaseRow(t *testing.T) {
	expected := `{"codigo_porte":5,"porte":"DEMAIS","razao_social":"Razão Social","codigo_natureza_juridica":21,"natureza_juridica":"Natureza Jurídica","qualificacao_do_responsavel":13,"capital_social":4.2,"ente_federativo_responsavel":"Responsável"}`
	baseLookups.natures[21] = "Natureza Jurídica"
	b, err := loadBaseRow(&baseLookups, baseCSVRow)
	if err != nil {
		t.Errorf("expected no error loading base data row, got %s", err)
	}
	if string(b) != expected {
		t.Errorf("expected row to be %s, got %s", expected, string(b))
	}
}
