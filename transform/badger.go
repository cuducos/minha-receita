package transform

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/cuducos/go-cnpj"
	"github.com/dgraph-io/badger/v4"
)

func keyForPartners(n string) string    { return fmt.Sprintf("p-%s", n) }
func keyForBase(n string) string        { return fmt.Sprintf("b-%s", n) }
func keyForSimpleTaxes(n string) string { return fmt.Sprintf("st-%s", n) }
func keyForTaxRegime(n string) string   { return fmt.Sprintf("tr-%s", cnpj.Unmask(n)) }

// functions to read data from Badger

func partnersOf(db *badger.DB, n string) ([]PartnerData, error) {
	p := []PartnerData{}
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

func simpleTaxesOf(db *badger.DB, n string) (simpleTaxesData, error) {
	var d simpleTaxesData
	err := db.View(func(txn *badger.Txn) error {
		i, err := txn.Get([]byte(keyForSimpleTaxes(n)))
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("could not get key %s: %w", keyForSimpleTaxes(n), err)
		}
		v, err := i.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("could not read value for key %s: %w", keyForSimpleTaxes(n), err)
		}
		if err := json.Unmarshal(v, &d); err != nil {
			return fmt.Errorf("could not parse taxes: %w", err)
		}
		return nil
	})
	if err != nil {
		return simpleTaxesData{}, fmt.Errorf("error getting taxes for %s: %w", n, err)
	}
	return d, nil
}

func taxRegimeOf(db *badger.DB, n string) (TaxRegimes, error) {
	var d TaxRegimes
	err := db.View(func(txn *badger.Txn) error {
		i, err := txn.Get([]byte(keyForTaxRegime(n)))
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("could not get key %s: %w", keyForSimpleTaxes(n), err)
		}
		v, err := i.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("could not read value for key %s: %w", keyForSimpleTaxes(n), err)
		}
		if err := json.Unmarshal(v, &d); err != nil {
			return fmt.Errorf("could not parse taxes: %w", err)
		}
		return nil
	})
	if err != nil {
		return TaxRegimes{}, fmt.Errorf("error getting tax regimes for %s: %w", n, err)
	}
	sort.Sort(TaxRegimes(d))
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
	qsa := []PartnerData{}
	if curr != nil {
		if err := json.Unmarshal(curr, &qsa); err != nil {
			return nil, fmt.Errorf("could not parse partners: %w", err)
		}
	}
	var p PartnerData
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

func mergeTaxRegimes(db *badger.DB, k, b []byte) ([]byte, error) {
	curr := []byte("[]")
	err := db.View(func(tx *badger.Txn) error {
		i, err := tx.Get(k)
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error getting tax regime key: %w", err)
		}
		curr, err = i.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("error reading tax regime value: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error getting current tax regimes: %w", err)
	}
	ts := TaxRegimes{}
	if curr != nil {
		if err := json.Unmarshal(curr, &ts); err != nil {
			return nil, fmt.Errorf("could not parse tax regimes: %w", err)
		}
	}
	var t TaxRegime
	if err := json.Unmarshal(b, &t); err != nil {
		return nil, fmt.Errorf("could not parse tax regime: %w", err)
	}
	ts = append(ts, t)
	j, err := json.Marshal(&ts)
	if err != nil {
		return nil, fmt.Errorf("could not convert tax regime to json: %w", err)
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
	if s == realProfit || s == presumedProfit || s == arbitratedProfit || s == noTaxes {
		v, err = mergeTaxRegimes(db, k, v)
		if err != nil {
			return fmt.Errorf("error merging taxes: %w", err)
		}
	}
	return db.Update(func(tx *badger.Txn) error { return tx.Set(k, v) })
}
