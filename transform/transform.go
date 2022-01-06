package transform

import (
	"fmt"
)

// MaxParallelDBQueries is the default for maximum number of parallels save
// queries sent to the database
const MaxParallelDBQueries = 32

// BatchSize determines the size of the batches used to create the initial JSON
// data in the database.
const BatchSize = 8192

type database interface {
	GetCompany(string) (string, error)
	CreateCompanies([][]string) error
	UpdateCompany(string, string) error
	ListCompanies(string) ([]string, error)
}

// Transform the downloaded files for company venues creating a database record
// per CNPJ
func Transform(dir string, db database, maxParallelDBQueries, batchSize int) error {
	t, err := createJSONRecordsTask(dir, db, batchSize)
	if err != nil {
		return fmt.Errorf("error creating new task for venues in %s: %w", dir, err)
	}
	if err := t.run(maxParallelDBQueries); err != nil {
		return err
	}
	for _, f := range []updateFunc{addBases, addPartners, addTaxes} {
		if err := f(dir, db, t.lookups); err != nil {
			return err
		}
	}
	return nil
}
