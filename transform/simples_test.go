package transform

import (
	"path/filepath"
	"testing"
	"time"
)

func TestAddSimplesToCompanies(t *testing.T) {
	d := t.TempDir()
	c := company{CNPJ: "33683111000280"}
	p, err := c.toJSON(d)
	if err != nil {
		t.Errorf("expected to error creating a json, got %s", err)
	}
	ls, err := PathsForSource(simple, filepath.Join("..", "testdata"))
	if err != nil {
		t.Errorf("expected no error finding paths for %s, got %s", string(simple), err)
	}
	for _, f := range ls {
		copyFile(f, d)
	}
	if err := addSimplesToCompanies(d); err != nil {
		t.Errorf("expected no errors adding partners, got %s", err)
		return
	}
	c, err = companyFromJSON(p)
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
