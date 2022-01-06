package transform

import (
	"fmt"
	"log"

	"github.com/cuducos/go-cnpj"
)

func addSimplesToCompany(_ *lookups, db database, r []string) error {
	strs, err := db.ListCompanies(r[0])
	if err != nil {
		return fmt.Errorf("error loading companies with base %s: %w", r[0], err)
	}
	if len(strs) == 0 {
		log.Output(2, fmt.Sprintf("No company found for CNPJ base %s", r[0]))
		return nil
	}
	for _, s := range strs {
		c, err := companyFromString(s)
		if err != nil {
			return fmt.Errorf("error loading company: %w", err)
		}
		c.OpcaoPeloSimples = toBool(r[1])
		c.DataOpcaoPeloSimples, err = toDate(r[2])
		if err != nil {
			return fmt.Errorf("error parsing DataOpcaoPeloSimples %s: %w", r[2], err)
		}
		c.DataExclusaoDoSimples, err = toDate(r[3])
		if err != nil {
			return fmt.Errorf("error parsing DataExclusaoDoSimples %s: %w", r[3], err)
		}
		c.OpcaoPeloMEI = toBool(r[4])
		c.DataOpcaoPeloMEI, err = toDate(r[5])
		if err != nil {
			return fmt.Errorf("error parsing DataOpcaoPeloMEI %s: %w", r[5], err)
		}
		c.DataExclusaoDoMEI, err = toDate(r[6])
		if err != nil {
			return fmt.Errorf("error parsing DataExclusaoDoMEI %s: %w", r[6], err)
		}
		if err = c.Update(db); err != nil {
			return fmt.Errorf("error saving %s: %w", cnpj.Mask(c.CNPJ), err)
		}
	}
	return nil
}

func addSimplesToCompanies(dir string, db database) error {
	t, err := newUpdateTask(simple, dir, db, nil)
	if err != nil {
		return fmt.Errorf("error creating update task for taxes: %w", err)
	}
	return t.run("Adding Simples and MEI", addSimplesToCompany)
}
