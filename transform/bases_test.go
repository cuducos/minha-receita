package transform

import "testing"

func newTestBaseCNPJ() baseData {
	codigoPorte := 5
	porte := "DEMAIS"
	codigoNaturezaJuridica := 2011
	naturezaJuridica := "Empresa Pública"
	qualificacaoDoResponsavel := 13
	capitalSocial := float32(4.2)
	return baseData{
		&codigoPorte,
		&porte,
		"Razão Social",
		&codigoNaturezaJuridica,
		&naturezaJuridica,
		&qualificacaoDoResponsavel,
		&capitalSocial,
		"Responsável",
	}
}

var (
	baseCSVRow = []string{
		"BASE DO CNPJ",
		"Razão Social",
		"2011",
		"13",
		"4.20",
		"5",
		"Responsável",
	}
)

func TestNewBase(t *testing.T) {
	l, err := newLookups(testdata)
	if err != nil {
		t.Fatalf("could not create lookups: %s", err)
	}
	b, err := newBaseData(&l, baseCSVRow)
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
	if *b.CodigoNaturezaJuridica != 2011 {
		t.Errorf("expected CodigoNaturezaJuridica to be 2011, got %d", *b.CodigoNaturezaJuridica)
	}
	if *b.NaturezaJuridica != "Empresa Pública" {
		t.Errorf("expected NaturezaJuridica to be %s, got %s", l.natures[21], *b.NaturezaJuridica)
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
	l, err := newLookups(testdata)
	if err != nil {
		t.Fatalf("could not create lookups: %s", err)
	}
	expected := `{"codigo_porte":5,"porte":"DEMAIS","razao_social":"Razão Social","codigo_natureza_juridica":2011,"natureza_juridica":"Empresa Pública","qualificacao_do_responsavel":13,"capital_social":4.2,"ente_federativo_responsavel":"Responsável"}`
	b, err := loadBaseRow(&l, baseCSVRow)
	if err != nil {
		t.Errorf("expected no error loading base data row, got %s", err)
	}
	if string(b) != expected {
		t.Errorf("expected row to be %s, got %s", expected, string(b))
	}
}
