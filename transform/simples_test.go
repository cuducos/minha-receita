package transform

import (
	"testing"
	"time"
)

func TestAddSimplesToCompanies(t *testing.T) {
	db := newMockDB()
	c := company{CNPJ: "33683111000280"}
	if err := c.Save(&db); err != nil {
		t.Errorf("expected no error saving a company, got %s", err)
	}
	if err := addSimplesToCompanies(testdata, &db); err != nil {
		t.Errorf("expected no errors adding partners, got %s", err)
		return
	}
	c, err := companyFromDB(&db, c.CNPJ)
	if err != nil {
		t.Errorf("expected no errors loading company, got %s", err)
	}
	dataOpcaoPeloSimples, err := time.Parse("2006-01-02", "2014-01-01")
	if err != nil {
		t.Errorf("expected no errors creating date got %s", err)
	}

	if *c.OpcaoPeloSimples != true {
		t.Errorf("expected OpcaoPeloSimples to be true, got %v", c.OpcaoPeloSimples)
	}
	if *c.DataOpcaoPeloSimples != date(dataOpcaoPeloSimples) {
		t.Errorf("expected DataOpcaoPeloSimples to be %s, got %s",
			dataOpcaoPeloSimples,
			time.Time(*c.DataOpcaoPeloSimples),
		)
	}
	if c.DataExclusaoDoSimples != nil {
		t.Errorf("expected DataExclusaoDoSimples to be nil, got %s", time.Time(*c.DataExclusaoDoSimples))
	}
	if *c.OpcaoPeloMEI != false {
		t.Errorf("expected OpcaoPeloMEI to be false, got %v", c.OpcaoPeloMEI)
	}
	if c.DataOpcaoPeloMEI != nil {
		t.Errorf("expected DataOpcaoPeloMEI to be nil, got %s", time.Time(*c.DataOpcaoPeloMEI))
	}
	if c.DataExclusaoDoMEI != nil {
		t.Errorf("expected DataExclusaoDoMEI to be nil, got %s", time.Time(*c.DataExclusaoDoMEI))
	}
}
