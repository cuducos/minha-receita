package transform

import (
	"fmt"
)

// MaxParallelDBQueries is the default for maximum number of parallels save
// queries sent to the database
const MaxParallelDBQueries = 8

// BatchSize determines the size of the batches used to create the initial JSON
// data in the database.
const BatchSize = 8192

type database interface {
	CreateCompanies([][]any) error
	CreateIndex() error
	UpdateCompanies([][]string) error
	AddPartners([][]string) error
	MetaSave(string, string) error
	PreLoad() error
	PostLoad() error
}

// Transform the downloaded files for company venues creating a database record
// per CNPJ
func Transform(dir string, db database, maxParallelDBQueries, batchSize int, privacy bool) error {
	if err := db.PreLoad(); err != nil {
		return fmt.Errorf("error running pre-load: %w", err)
	}
	if err := saveUpdatedAt(db, dir); err != nil {
		return fmt.Errorf("error saving the update at date: %w", err)
	}
	j, err := createJSONRecordsTask(dir, db, batchSize, privacy)
	if err != nil {
		return fmt.Errorf("error creating new task for venues in %s: %w", dir, err)
	}
	err = func() error {
		defer j.bar.Close()
		return j.run(maxParallelDBQueries)
	}()
	if err != nil {
		return err
	}
	u, err := newUpdateTask(dir, db, batchSize, j.lookups)
	if err != nil {
		return fmt.Errorf("error creating update task: %w", err)
	}
	defer u.bar.Close()
	if err := u.run(); err != nil {
		return fmt.Errorf("error running update task: %w", err)
	}
	if err := db.PostLoad(); err != nil {
		return fmt.Errorf("error running post-load: %w", err)
	}
	return nil
}
