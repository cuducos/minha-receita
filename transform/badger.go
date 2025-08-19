package transform

import (
	"encoding/json/v2"
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
	var ts TaxRegimes
	pre := []byte(keyForTaxRegime(n))
	db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(pre); it.ValidForPrefix(pre); it.Next() {
			var t TaxRegime
			i := it.Item()
			err := i.Value(func(v []byte) error {
				if err := json.Unmarshal(v, &t); err != nil {
					return fmt.Errorf("could not parse tax regime: %w", err)
				}
				ts = append(ts, t)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	sort.Sort(TaxRegimes(ts))
	return ts, nil
}

func partnersOf(db *badger.DB, n string) ([]PartnerData, error) {
	var ps []PartnerData
	pre := []byte(keyForPartners(n))
	db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(pre); it.ValidForPrefix(pre); it.Next() {
			var p PartnerData
			i := it.Item()
			err := i.Value(func(v []byte) error {
				if err := json.Unmarshal(v, &p); err != nil {
					return fmt.Errorf("could not parse parter: %w", err)
				}
				ps = append(ps, p)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return ps, nil
}
