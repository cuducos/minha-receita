package transform

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/dgraph-io/badger/v3"
)

const badgerFilePrefix = "minha-receita-badger-"

func keyForPartners(n string) string { return fmt.Sprintf("partners%s", n) }
func keyForBase(n string) string     { return fmt.Sprintf("base%s", n) }
func keyForTaxes(n string) string    { return fmt.Sprintf("taxes%s", n) }

func partnersOf(b *badgerStorage, n string) ([]partner, error) {
	p := []partner{}
	err := b.db.View(func(txn *badger.Txn) error {
		i, err := txn.Get([]byte(keyForPartners(n)))
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("could not get key %s: %w", keyForPartners(n), err)
		}
		v, err := i.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("could not read value for key %s: %w", keyForPartners(n), err)
		}
		if err := json.Unmarshal(v, &p); err != nil {
			return fmt.Errorf("could not parse partners: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error getting partners for %s: %w", n, err)
	}
	return p, nil
}

func baseOf(b *badgerStorage, n string) (baseData, error) {
	var d baseData
	err := b.db.View(func(txn *badger.Txn) error {
		i, err := txn.Get([]byte(keyForBase(n)))
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("could not get key %s: %w", keyForBase(n), err)
		}
		v, err := i.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("could not read value for key %s: %w", keyForBase(n), err)
		}
		if err := json.Unmarshal(v, &d); err != nil {
			return fmt.Errorf("could not parse base: %w", err)
		}
		return nil
	})
	if err != nil {
		return baseData{}, fmt.Errorf("error getting base for %s: %w", n, err)
	}
	return d, nil
}

func taxesOf(b *badgerStorage, n string) (taxesData, error) {
	var d taxesData
	err := b.db.View(func(txn *badger.Txn) error {
		i, err := txn.Get([]byte(keyForTaxes(n)))
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("could not get key %s: %w", keyForTaxes(n), err)
		}
		v, err := i.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("could not read value for key %s: %w", keyForTaxes(n), err)
		}
		if err := json.Unmarshal(v, &d); err != nil {
			return fmt.Errorf("could not parse taxes: %w", err)
		}
		return nil
	})
	if err != nil {
		return taxesData{}, fmt.Errorf("error getting taxes for %s: %w", n, err)
	}
	return d, nil
}

type badgerStorage struct {
	db   *badger.DB
	path string
}

func (b *badgerStorage) close() error {
	b.db.Close()
	if err := os.RemoveAll(b.path); err != nil {
		return fmt.Errorf("error cleaning up badger storage directory: %w", err)
	}
	return nil
}

type badgerLogger struct{}

func (*badgerLogger) Errorf(string, ...interface{})   {}
func (*badgerLogger) Warningf(string, ...interface{}) {}
func (*badgerLogger) Infof(string, ...interface{})    {}
func (*badgerLogger) Debugf(string, ...interface{})   {}

func newBadgerStorage() (*badgerStorage, error) {
	d, err := os.MkdirTemp("", badgerFilePrefix)
	if err != nil {
		return nil, fmt.Errorf("error creating temporary key-value storage: %w", err)
	}
	if os.Getenv("DEBUG") != "" {
		log.Output(1, fmt.Sprintf("Creating temporary key-value storage at %s", d))
	}
	o := badger.DefaultOptions(d)
	o.Logger = &badgerLogger{}
	db, err := badger.Open(o)
	if err != nil {
		return nil, fmt.Errorf("error creating badger key-value object: %w", err)
	}
	return &badgerStorage{db: db, path: d}, nil
}
