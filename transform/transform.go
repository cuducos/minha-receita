package transform

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/cuducos/minha-receita/download"
)

// MaxParallelDBQueries is the default for maximum number of parallels save
// queries sent to the database
const MaxParallelDBQueries = 8

// BatchSize determines the size of the batches used to create the initial JSON
// data in the database.
const BatchSize = 8192

type database interface {
	PreLoad() error
	CreateCompanies([][]any) error
	PostLoad() error
	MetaSave(string, string) error
}

type kvStorage interface {
	load(string, *lookups) error
	enrichCompany(*company) error
	close() error
}

func saveUpdatedAt(db database, dir string) error {
	log.Output(1, "Saving the updated at date to the database…")
	p := filepath.Join(dir, download.FederalRevenueUpdatedAt)
	v, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", p, err)

	}
	return db.MetaSave("updated-at", string(v))
}

// Transform the downloaded files for company venues creating a database record
// per CNPJ
func Transform(dir string, db database, maxParallelDBQueries, batchSize int, privacy, mem bool) error {
	l, err := newLookups(dir)
	if err != nil {
		return fmt.Errorf("error creating look up tables from %s: %w", dir, err)
	}
	kv, err := newBadgerStorage(mem)
	if err != nil {
		return fmt.Errorf("could not create badger storage: %w", err)
	}
	defer kv.close()
	if err := kv.load(dir, &l); err != nil {
		return fmt.Errorf("error loading data to badger: %w", err)
	}
	j, err := createJSONRecordsTask(dir, db, &l, kv, batchSize, privacy)
	if err != nil {
		return fmt.Errorf("error creating new task for venues in %s: %w", dir, err)
	}
	if err := j.run(maxParallelDBQueries); err != nil {
		return fmt.Errorf("error writing venues to database: %w", err)
	}
	return saveUpdatedAt(db, dir)
}
