package transform

import (
	"encoding/json"
	"fmt"
)

type taxesData struct {
	OpcaoPeloSimples      *bool `json:"opcao_pelo_simples"`
	DataOpcaoPeloSimples  *date `json:"data_opcao_pelo_simples"`
	DataExclusaoDoSimples *date `json:"data_exclusao_do_simples"`
	OpcaoPeloMEI          *bool `json:"opcao_pelo_mei"`
	DataOpcaoPeloMEI      *date `json:"data_opcao_pelo_mei"`
	DataExclusaoDoMEI     *date `json:"data_exclusao_do_mei"`
}

func addTax(_ *lookups, db database, r []string) error {
	var err error
	d := taxesData{
		OpcaoPeloSimples: toBool(r[1]),
		OpcaoPeloMEI:     toBool(r[4]),
	}
	d.DataOpcaoPeloSimples, err = toDate(r[2])
	if err != nil {
		return fmt.Errorf("error parsing DataOpcaoPeloSimples %s: %w", r[2], err)
	}
	d.DataExclusaoDoSimples, err = toDate(r[3])
	if err != nil {
		return fmt.Errorf("error parsing DataExclusaoDoSimples %s: %w", r[3], err)
	}
	d.DataOpcaoPeloMEI, err = toDate(r[5])
	if err != nil {
		return fmt.Errorf("error parsing DataOpcaoPeloMEI %s: %w", r[5], err)
	}
	d.DataExclusaoDoMEI, err = toDate(r[6])
	if err != nil {
		return fmt.Errorf("error parsing DataExclusaoDoMEI %s: %w", r[6], err)
	}
	b, err := json.Marshal(&d)
	if err != nil {
		return fmt.Errorf("error converting taxes data to json for %s: %w", r[0], err)
	}
	if err = db.UpdateCompanies(r[0], string(b)); err != nil {
		return fmt.Errorf("error updating taxes for base cnpj %s: %w", r[0], err)
	}
	return nil
}
