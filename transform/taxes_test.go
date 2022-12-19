package transform

import (
	"testing"
	"time"
)

func newTestTaxes() taxesData {
	simples := true
	dtSimples := date(time.Date(2022, time.December, 17, 0, 0, 0, 0, time.UTC))
	mei := false
	inMEI := date(time.Date(2022, time.November, 18, 0, 0, 0, 0, time.UTC))
	outMEI := date(time.Date(2022, time.December, 1, 0, 0, 0, 0, time.UTC))
	return taxesData{&simples, &dtSimples, nil, &mei, &inMEI, &outMEI}
}

var (
	taxesCSVRow = []string{
		"BASE DO CNPJ",
		"S",        // OpcaoPeloSimples
		"20221217", // DataOpcaoPeloSimples
		"",         // DataExclusaoDoSimples
		"N",        // OpcaoPeloMEI
		"20221118", // DataOpcaoPeloMEI
		"20221201", // DataExclusaoDoMEI
	}
)

func TestNewTaxes(t *testing.T) {
	d, err := newTaxesData(taxesCSVRow)
	if err != nil {
		t.Errorf("expected no error creating taxes data, got %s", err)
	}
	if *d.OpcaoPeloSimples == false {
		t.Errorf("expected OpcaoPeloSimples to be true, got %t", *d.OpcaoPeloSimples)
	}
	if *d.OpcaoPeloMEI != false {
		t.Errorf("expected OpcaoPeloMEI to be false, got %t", *d.OpcaoPeloMEI)
	}
	for _, tc := range []struct {
		name     string
		got      *date
		expected time.Time
	}{
		{"DataOpcaoPeloSimples", d.DataOpcaoPeloSimples, time.Date(2022, time.December, 17, 0, 0, 0, 0, time.UTC)},
		{name: "DataExclusaoDoSimples", got: d.DataExclusaoDoSimples},
		{"DataOpcaoPeloMEI", d.DataOpcaoPeloMEI, time.Date(2022, time.November, 18, 0, 0, 0, 0, time.UTC)},
		{"DataExclusaoDoMEI", d.DataExclusaoDoMEI, time.Date(2022, time.December, 1, 0, 0, 0, 0, time.UTC)},
	} {
		datePointerEqual(t, tc.got, tc.expected, tc.name)
	}
}

func TestLoadTaxesRow(t *testing.T) {
	expected := `{"opcao_pelo_simples":true,"data_opcao_pelo_simples":"2022-12-17","data_exclusao_do_simples":null,"opcao_pelo_mei":false,"data_opcao_pelo_mei":"2022-11-18","data_exclusao_do_mei":"2022-12-01"}`
	d, err := loadTaxesRow(&lookups{}, taxesCSVRow)
	if err != nil {
		t.Errorf("expected no error loading taxes data row, got %s", err)
	}
	if string(d) != expected {
		t.Errorf("expected row to be %s, got %s", expected, string(d))
	}
}

func datePointerEqual(t *testing.T, d *date, x time.Time, n string) {
	if d == nil && x.IsZero() {
		return
	}
	e := x.Format("2006-01-02")
	if d == nil && !x.IsZero() {
		t.Errorf("expected %s to be %s, got nil", n, e)
		return
	}
	f := time.Time(*d).Format("2006-01-02")
	if x.IsZero() && d != nil {
		t.Errorf("expected %s to be nil, got %s", n, f)
		return
	}
	if !x.Equal(time.Time(*d)) {
		t.Errorf("expected %s to be %s, got %s", n, e, f)
	}
}
