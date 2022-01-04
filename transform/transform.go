package transform

import (
	"fmt"
)

// MaxParallelDBQueries is the default for maximum number of parallels save
// queries sent to the database
const MaxParallelDBQueries = 512

type database interface {
	GetCompany(string) (string, error)
	ListCompanies(string) ([]string, error)
	SaveCompany(string, string) error
}

// Transform the downloaded files for company venues creating a database record
// per CNPJ
func Transform(dir string, db database, maxParallelDBQueries int) error {
	t, err := createJSONRecords(dir, db)
	if err != nil {
		return fmt.Errorf("error creating new task for venues in %s: %w", dir, err)
	}
	if err := t.run(maxParallelDBQueries); err != nil {
		return err
	}
	if err := addBases(dir, db, t.lookups); err != nil {
		return err
	}
	if err := addPartners(dir, db, t.lookups); err != nil {
		return err
	}
	return addSimplesToCompanies(dir, db)
}
