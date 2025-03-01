package transform

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/cuducos/minha-receita/download"
)

const (
	// MaxParallelDBQueries is the default for maximum number of parallels save
	// queries sent to the database
	MaxParallelDBQueries = 8

	// BatchSize determines the size of the batches used to create the initial JSON
	// data in the database.
	BatchSize = 8192
)

type database interface {
	PreLoad() error
	CreateCompanies([][]string) error
	PostLoad() error
	MetaSave(string, string) error
}

type kvStorage interface {
	load(string, *lookups) error
	enrichCompany(*Company) error
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

func createKeyValueStorage(dir string, pth string, l lookups) error {
	kv, err := newBadgerStorage(pth, false)
	if err != nil {
		return fmt.Errorf("could not create badger storage: %w", err)
	}
	defer kv.close()
	if err := kv.load(dir, &l); err != nil {
		return fmt.Errorf("error loading data to badger: %w", err)
	}
	return nil
}

func createJSONs(dir string, pth string, db database, l lookups, maxParallelDBQueries, batchSize int, privacy bool) error {
	kv, err := newBadgerStorage(pth, true)
	if err != nil {
		return fmt.Errorf("could not create badger storage: %w", err)
	}
	defer kv.close()
	j, err := createJSONRecordsTask(dir, db, &l, kv, batchSize, privacy)
	if err != nil {
		return fmt.Errorf("error creating new task for venues in %s: %w", dir, err)
	}
	if err := j.run(maxParallelDBQueries); err != nil {
		return fmt.Errorf("error writing venues to database: %w", err)
	}
	return saveUpdatedAt(db, dir)
}

// Transform the downloaded files for company venues creating a database record
// per CNPJ
func Transform(dir string, db database, max, s int, p bool) error {
	pth, err := os.MkdirTemp("", fmt.Sprintf("minha-receita-%s-*", time.Now().Format("20060102150405")))
	if err != nil {
		return fmt.Errorf("error creating temporary key-value storage: %w", err)
	}
	defer os.RemoveAll(pth)
	l, err := newLookups(dir)
	if err != nil {
		return fmt.Errorf("error creating look up tables from %s: %w", dir, err)
	}
	if err := createKeyValueStorage(dir, pth, l); err != nil {
		return err
	}
	return createJSONs(dir, pth, db, l, max, s, p)
}
