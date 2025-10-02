package transform

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/cuducos/minha-receita/download"
)

const (
	// MaxParallelDBQueries is the default for maximum number of parallels save
	// queries sent to the database
	MaxParallelDBQueries = 8

	// MaxParallelKVWrites is the default for maximum number of parallels
	// writes on the key-value storage (Badger)
	MaxParallelKVWrites = 1024

	// BatchSize determines the size of the batches used to create the initial JSON
	// data in the database.
	BatchSize = 8192
)

var extraIdexes = [...]string{
	"cnae_fiscal",
	"cnaes_secundarios.codigo",
	"codigo_municipio",
	"codigo_municipio_ibge",
	"codigo_natureza_juridica",
	"qsa.cnpj_cpf_do_socio",
	"uf",
}

type database interface {
	PreLoad() error
	CreateCompanies([][]string) error
	PostLoad() error
	CreateExtraIndexes([]string) error
	MetaSave(string, string) error
}

type kvStorage interface {
	load(string, *lookups, int) error
	enrichCompany(*Company) error
	close() error
}

func saveUpdatedAt(db database, dir string) error {
	slog.Info("Saving the updated at date to the database…")
	p := filepath.Join(dir, download.FederalRevenueUpdatedAt)
	v, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", p, err)

	}
	return db.MetaSave("updated-at", string(v))
}

func createKeyValueStorage(dir string, pth string, l lookups, maxKV int) (err error) { // using named return so we can set it in the defer call
	kv, err := newBadgerStorage(pth, false)
	if err != nil {
		return fmt.Errorf("could not create badger storage: %w", err)
	}
	defer func() {
		if e := kv.close(); e != nil && err == nil {
			err = fmt.Errorf("could not close key/value storage: %w", e)
		}
	}()
	if err := kv.load(dir, &l, maxKV); err != nil {
		return fmt.Errorf("error loading data to badger: %w", err)
	}
	return nil
}

func createJSONs(dir string, pth string, db database, l lookups, maxDB, batchSize int, privacy bool) error {
	kv, err := newBadgerStorage(pth, true)
	if err != nil {
		return fmt.Errorf("could not create badger storage: %w", err)
	}
	defer func() {
		if err := kv.close(); err != nil {
			slog.Warn("could not close key-value storage", "path", pth, "error", err)
		}
	}()
	j, err := createJSONRecordsTask(dir, db, &l, kv, batchSize, privacy)
	if err != nil {
		return fmt.Errorf("error creating new task for venues in %s: %w", dir, err)
	}
	if err := j.run(maxDB); err != nil {
		return fmt.Errorf("error writing venues to database: %w", err)
	}
	return saveUpdatedAt(db, dir)
}

func postLoad(db database) error {
	slog.Info("Consolidating the database…")
	if err := db.PostLoad(); err != nil {
		return err
	}
	slog.Info("Database consolidated!")
	slog.Info("Creating indexes…")
	if err := db.CreateExtraIndexes(extraIdexes[:]); err != nil {
		return err
	}
	slog.Info("Indexes created!")
	return nil
}

// Transform the downloaded files for company venues creating a database record
// per CNPJ
func Transform(dir string, db database, maxDB, maxKV, s int, p bool) error {
	pth, err := os.MkdirTemp("", fmt.Sprintf("minha-receita-%s-*", time.Now().Format("20060102150405")))
	if err != nil {
		return fmt.Errorf("error creating temporary key-value storage: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(pth); err != nil {
			slog.Error("could not remove temporary", "directory", pth, "error", err)
		}
	}()
	l, err := newLookups(dir)
	if err != nil {
		return fmt.Errorf("error creating look up tables from %s: %w", dir, err)
	}
	if err := createKeyValueStorage(dir, pth, l, 1024); err != nil {
		return err
	}
	if err := createJSONs(dir, pth, db, l, maxDB, s, p); err != nil {
		return err
	}
	return postLoad(db)
}
