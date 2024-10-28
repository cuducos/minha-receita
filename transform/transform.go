package transform

import (
	"errors"
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

	badgerFilePrefix = "minha-receita-badger-"
)

type database interface {
	PreLoad() error
	CreateCompanies([][]any) error
	PostLoad() error
	MetaSave(string, string) error
}

type kvStorage interface {
	load(string, *lookups) error
	enrichCompany(*company) error
	close(bool) error
}

type mode int

const (
	// runs the two steps and cleans up temporary key-value directory
	// from step one
	both mode = iota

	// loads relational data to a key-value store and does not delete it from
	// the temporary directory
	stepOne

	// expects a path to the directory created in StepOne and use it to persist
	// data to PostgreSQL
	stepTwo
)

func transformMode(s1 bool, s2 string) (mode, error) {
	switch {
	case s1 && s2 != "":
		return both, errors.New("cannot use both --step-one and --step-two")
	case s1 && s2 == "":
		return stepOne, nil
	case !s1 && s2 != "":
		return stepTwo, nil
	}
	return both, nil
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

func runStepOne(dir string, l lookups, isolated bool) (string, error) {
	tmp, err := os.MkdirTemp("", fmt.Sprintf("%s-%s", badgerFilePrefix, time.Now().Format("20060102150405")))
	if err != nil {
		return "", fmt.Errorf("error creating temporary key-value storage: %w", err)
	}
	kv, err := newBadgerStorage(tmp)
	if err != nil {
		return "", fmt.Errorf("could not create badger storage: %w", err)
	}
	defer kv.close(isolated)
	if err := kv.load(dir, &l); err != nil {
		return "", fmt.Errorf("error loading data to badger: %w", err)
	}
	if isolated {
		fmt.Println(kv.path)
	}
	return kv.path, nil
}

func runStepTwo(dir string, tmp string, db database, l lookups, maxParallelDBQueries, batchSize int, privacy, isolated bool) error {
	kv, err := newBadgerStorage(tmp)
	if err != nil {
		return fmt.Errorf("could not create badger storage: %w", err)
	}
	defer kv.close(isolated)
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
func Transform(dir string, db database, maxParallelDBQueries, batchSize int, privacy, s1 bool, s2 string) error {
	m, err := transformMode(s1, s2)
	if err != nil {
		return fmt.Errorf("error determining transform mode: %w", err)
	}
	var tmp string
	l, err := newLookups(dir)
	if err != nil {
		return fmt.Errorf("error creating look up tables from %s: %w", dir, err)
	}
	if m != stepTwo {
		tmp, err = runStepOne(dir, l, m == stepOne)
		if err != nil {
			return fmt.Errorf("error creating key-value storage: %w", err)
		}
	}
	if m != stepOne {
		if s2 != "" {
			tmp = s2
		}
		return runStepTwo(dir, tmp, db, l, maxParallelDBQueries, batchSize, privacy, m == stepTwo)
	}
	return nil
}
