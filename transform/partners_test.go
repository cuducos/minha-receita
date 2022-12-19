package transform

import (
	"testing"
	"time"
)

func newTestPartner() partnerData {
	identificacaoDoSocio := 1
	codigoQualificacaoSocio := 16
	qualificacaoSocio := "Presidente"
	codigoPais := 105
	pais := "BRASIL"
	codigoQualificacaoRepresentanteLegal := 10
	qualificacaoRepresentanteLegal := "Diretor"
	codigoFaixaEtaria := 4
	faixaEtarua := "Entre 31 a 40 anos"
	dataEntradaSociedade := date(time.Date(2007, 8, 12, 0, 0, 0, 0, time.UTC))
	return partnerData{
		&identificacaoDoSocio,
		"Hannah",
		"123",
		&codigoQualificacaoSocio,
		&qualificacaoSocio,
		&dataEntradaSociedade,
		&codigoPais,
		&pais,
		"789",
		"Arendt",
		&codigoQualificacaoRepresentanteLegal,
		&qualificacaoRepresentanteLegal,
		&codigoFaixaEtaria,
		&faixaEtarua,
	}
}

var (
	partnerCSVRow = []string{
		"BASE DO CNPJ",
		"1",        // IdentificadorDeSocio
		"Hannah",   // NomeSocio
		"123",      // CNPJCPFDoSocio
		"16",       // CodigoQualificacaoSocio
		"20070812", // DataEntradaSociedade
		"105",      // CodigoPais
		"789",      // CPFRepresentanteLegal
		"Arendt",   // NomeRepresentanteLegal
		"10",       // CodigoQualificacaoRepresentanteLegal
		"4",        // CodigoFaixaEtaria
	}
)

func TestNewPartner(t *testing.T) {
	l, err := newLookups(testdata)
	if err != nil {
		t.Fatalf("could not create lookups: %s", err)
	}
	p, err := newPartnerData(&l, partnerCSVRow)
	if err != nil {
		t.Errorf("expected no error creating partner data, got %s", err)
	}
	if *p.IdentificadorDeSocio != 1 {
		t.Errorf("expected IdentificadorDeSocio to be 1, got %d", *p.IdentificadorDeSocio)
	}
	if p.NomeSocio != "Hannah" {
		t.Errorf("expected NomeSocio to be Hannah, got %s", p.NomeSocio)
	}
	if p.CNPJCPFDoSocio != "123" {
		t.Errorf("expected CNPJCPFDoSocio to be 123, got %s", p.CNPJCPFDoSocio)
	}
	if *p.CodigoQualificacaoSocio != 16 {
		t.Errorf("expected CodigoQualificacaoSocio to be 16, got %d", *p.CodigoQualificacaoSocio)
	}
	if *p.QualificaoSocio != "Presidente" {
		t.Errorf("expected QualificaoSocio to be Presidente, got %s", *p.QualificaoSocio)
	}
	datePointerEqual(t, p.DataEntradaSociedade, time.Date(2007, 8, 12, 0, 0, 0, 0, time.UTC), "DataEntradaSociedade")
	if *p.CodigoPais != 105 {
		t.Errorf("expected CodigoPais to be 105, got %d", *p.CodigoPais)
	}
	if *p.Pais != "BRASIL" {
		t.Errorf("expected Pais to be Brasil, got %s", *p.Pais)
	}
	if p.CPFRepresentanteLegal != "789" {
		t.Errorf("expected CPFRepresentanteLegal to be 789, got %s", p.CPFRepresentanteLegal)
	}
	if p.NomeRepresentanteLegal != "Arendt" {
		t.Errorf("expected NomeRepresentanteLegal to be Arendt, got %s", p.NomeRepresentanteLegal)
	}
	if *p.CodigoQualificacaoRepresentanteLegal != 10 {
		t.Errorf("expected CodigoQualificacaoRepresentanteLegal to be 10, got %d", *p.CodigoQualificacaoRepresentanteLegal)
	}
	if *p.QualificacaoRepresentanteLegal != "Diretor" {
		t.Errorf("expected QualificacaoRepresentanteLegal to be Diretor, got %s", *p.QualificacaoRepresentanteLegal)
	}
	if *p.CodigoFaixaEtaria != 4 {
		t.Errorf("expected CodigoFaixaEtaria to be 4, got %d", *p.CodigoFaixaEtaria)
	}
	if *p.FaixaEtaria != "Entre 31 a 40 anos" {
		t.Errorf("expected FaixaEtaria to be Entre 31 a 40 anos, got %s", *p.FaixaEtaria)
	}
}

func TestLoadPartnerRow(t *testing.T) {
	expected := `{"identificador_de_socio":1,"nome_socio":"Hannah","cnpj_cpf_do_socio":"123","codigo_qualificacao_socio":16,"qualificacao_socio":"Presidente","data_entrada_sociedade":"2007-08-12","codigo_pais":105,"pais":"BRASIL","cpf_representante_legal":"789","nome_representante_legal":"Arendt","codigo_qualificacao_representante_legal":10,"qualificacao_representante_legal":"Diretor","codigo_faixa_etaria":4,"faixa_etaria":"Entre 31 a 40 anos"}`
	l, err := newLookups(testdata)
	if err != nil {
		t.Fatalf("could not create lookups: %s", err)
	}
	b, err := loadPartnerRow(&l, partnerCSVRow)
	if err != nil {
		t.Errorf("expected no error loading partner row, got %s", err)
	}
	if string(b) != expected {
		t.Errorf("expected row to be %s, got %s", expected, string(b))
	}
}
