package transform

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

const badgerFilePrefix = "minha-receita-badger-"

func keyForPartners(n string) string { return fmt.Sprintf("partners%s", n) }
func keyForBase(n string) string     { return fmt.Sprintf("base%s", n) }
func keyForTaxes(n string) string    { return fmt.Sprintf("taxes%s", n) }

// functions to read data from Badger

func partnersOf(db *badger.DB, n string) ([]partnerData, error) {
	p := []partnerData{}
	err := db.View(func(txn *badger.Txn) error {
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

func baseOf(db *badger.DB, n string) (baseData, error) {
	var d baseData
	err := db.View(func(txn *badger.Txn) error {
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

func taxesOf(db *badger.DB, n string) (taxesData, error) {
	var d taxesData
	err := db.View(func(txn *badger.Txn) error {
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

// functions to write data to Badger

func mergePartners(db *badger.DB, k, b []byte) ([]byte, error) {
	curr := []byte("[]")
	err := db.View(func(tx *badger.Txn) error {
		i, err := tx.Get(k)
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error getting partner key: %w", err)
		}
		curr, err = i.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("error reading partner value: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error getting current partners: %w", err)
	}
	qsa := []partnerData{}
	if curr != nil {
		if err := json.Unmarshal(curr, &qsa); err != nil {
			return nil, fmt.Errorf("could not parse partners: %w", err)
		}
	}
	var p partnerData
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, fmt.Errorf("could not parse partner: %w", err)
	}
	qsa = append(qsa, p)
	j, err := json.Marshal(&qsa)
	if err != nil {
		return nil, fmt.Errorf("could not convert partner to json: %w", err)
	}
	return j, nil
}

func saveItem(db *badger.DB, s sourceType, k, v []byte) (err error) {
	if s == partners {
		v, err = mergePartners(db, k, v)
		if err != nil {
			return fmt.Errorf("error merging partners: %w", err)
		}
	}
	return db.Update(func(tx *badger.Txn) error { return tx.Set(k, v) })
}
