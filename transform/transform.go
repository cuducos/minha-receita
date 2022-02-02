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
	UpdateCompanies(string, string) error
	AddPartner(string, string) error
	ListCompanies(string) ([]string, error)
}

// Transform the downloaded files for company venues creating a database record
// per CNPJ
func Transform(dir string, db database, maxParallelDBQueries, batchSize int) error {
	j, err := createJSONRecordsTask(dir, db, batchSize)
	if err != nil {
		return fmt.Errorf("error creating new task for venues in %s: %w", dir, err)
	}
	if err := j.run(maxParallelDBQueries); err != nil {
		return err
	}
	u, err := newUpdateTask(dir, db, j.lookups)
	if err != nil {
		return fmt.Errorf("error creating update task: %w", err)
	}
	return u.run()
}
