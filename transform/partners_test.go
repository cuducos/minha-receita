package transform

import (
	"testing"
	"time"
)

var (
	partnerCSVRow = []string{
		"BASE DO CNPJ",
		"1",        // IdentificadorDeSocio
		"Hannah",   // NomeSocio
		"123",      // CNPJCPFDoSocio
		"42",       // CodigoQualificacaoSocio
		"20070812", // DataEntradaSociedade
		"5",        // CodigoPais
		"789",      // CPFRepresentanteLegal
		"Arendt",   // NomeRepresentanteLegal
		"21",       // CodigoQualificacaoRepresentanteLegal
		"4",        // CodigoFaixaEtaria
	}
	partnerLookups = lookups{
		qualifications: make(map[int]string),
		countries:      make(map[int]string),
	}
)

func TestNewPartner(t *testing.T) {
	partnerLookups.qualifications[42] = "Resposta"
	partnerLookups.qualifications[21] = "Metade"
	partnerLookups.countries[5] = "Brasil"
	p, err := newPartnerData(&partnerLookups, partnerCSVRow)
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
	if *p.CodigoQualificacaoSocio != 42 {
		t.Errorf("expected CodigoQualificacaoSocio to be 42, got %d", *p.CodigoQualificacaoSocio)
	}
	if *p.QualificaoSocio != "Resposta" {
		t.Errorf("expected QualificaoSocio to be Resposte, got %s", *p.QualificaoSocio)
	}
	datePointerEqual(t, p.DataEntradaSociedade, time.Date(2007, 8, 12, 0, 0, 0, 0, time.UTC), "DataEntradaSociedade")
	if *p.CodigoPais != 5 {
		t.Errorf("expected CodigoPais to be 5, got %d", *p.CodigoPais)
	}
	if *p.Pais != "Brasil" {
		t.Errorf("expected Pais to be Brasil, got %s", *p.Pais)
	}
	if p.CPFRepresentanteLegal != "789" {
		t.Errorf("expected CPFRepresentanteLegal to be 789, got %s", p.CPFRepresentanteLegal)
	}
	if p.NomeRepresentanteLegal != "Arendt" {
		t.Errorf("expected NomeRepresentanteLegal to be Arendt, got %s", p.NomeRepresentanteLegal)
	}
	if *p.CodigoQualificacaoRepresentanteLegal != 21 {
		t.Errorf("expected CodigoQualificacaoRepresentanteLegal to be 21, got %d", *p.CodigoQualificacaoRepresentanteLegal)
	}
	if *p.QualificacaoRepresentanteLegal != "Metade" {
		t.Errorf("expected QualificacaoRepresentanteLegal to be Metade, got %s", *p.QualificacaoRepresentanteLegal)
	}
	if *p.CodigoFaixaEtaria != 4 {
		t.Errorf("expected CodigoFaixaEtaria to be 4, got %d", *p.CodigoFaixaEtaria)
	}
	if *p.FaixaEtaria != "Entre 31 a 40 anos" {
		t.Errorf("expected FaixaEtaria to be Entre 31 a 40 anos, got %s", *p.FaixaEtaria)
	}
}

func TestLoadPartnerRow(t *testing.T) {
	expected := `{"identificador_de_socio":1,"nome_socio":"Hannah","cnpj_cpf_do_socio":"123","codigo_qualificacao_socio":42,"qualificacao_socio":"Resposta","data_entrada_sociedade":"2007-08-12","codigo_pais":5,"pais":"Brasil","cpf_representante_legal":"789","nome_representante_legal":"Arendt","codigo_qualificacao_representante_legal":21,"qualificacao_representante_legal":"Metade","codigo_faixa_etaria":4,"faixa_etaria":"Entre 31 a 40 anos"}`
	partnerLookups.qualifications[42] = "Resposta"
	partnerLookups.qualifications[21] = "Metade"
	partnerLookups.countries[5] = "Brasil"
	b, err := loadPartnerRow(&partnerLookups, partnerCSVRow)
	if err != nil {
		t.Errorf("expected no error loading partner row, got %s", err)
	}
	if string(b) != expected {
		t.Errorf("expected row to be %s, got %s", expected, string(b))
	}
}
