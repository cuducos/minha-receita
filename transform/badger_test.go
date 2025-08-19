package transform

import (
	"encoding/json/v2"
	"reflect"
	"testing"

	"github.com/dgraph-io/badger/v4"
)

const testBaseCNPJ = "12345678"

func newTestBadgerDB(t *testing.T) *badger.DB {
	opt := badger.DefaultOptions(t.TempDir())
	db, err := badger.Open(opt)
	if err != nil {
		t.Fatal("could not create a badger database")
	}
	return db
}

func toBytes(t *testing.T, i any) []byte {
	b, err := json.Marshal(i)
	if err != nil {
		t.Fatalf("error marshaling %v: %s", i, err)
	}
	return b
}

func saveItem(t *testing.T, db *badger.DB, k string, v any) error {
	return db.Update(func(tx *badger.Txn) error {
		return tx.Set([]byte(k), toBytes(t, v))
	})
}

func TestReadItems(t *testing.T) {
	t.Run("partners", func(t *testing.T) {
		p := newTestPartner()
		db := newTestBadgerDB(t)
		defer db.Close()
		if err := saveItem(t, db, keyForPartners(testBaseCNPJ)+":md5hash", p); err != nil {
			t.Errorf("expected no error saving partner, got %s", err)
		}
		got, err := partnersOf(db, testBaseCNPJ)
		if err != nil {
			t.Errorf("expected no error reading partners, got %s", err)
		}
		if len(got) != 1 {
			t.Errorf("expected merged partners to have 1 partner, got %d", len(got))
			return
		}
		if !reflect.DeepEqual(got[0], p) {
			t.Errorf("expected merged partner to be %v, got %v", p, got[0])
		}
	})

	t.Run("base", func(t *testing.T) {
		db := newTestBadgerDB(t)
		defer db.Close()
		d := newTestBaseCNPJ()
		if err := saveItem(t, db, keyForBase(testBaseCNPJ), d); err != nil {
			t.Errorf("expected no error saving partner, got %s", err)
		}
		got, err := baseOf(db, testBaseCNPJ)
		if err != nil {
			t.Errorf("expected no error reading base, got %s", err)
		}
		if !reflect.DeepEqual(got, d) {
			t.Errorf("expected %v, got %v", d, got)
		}
	})

	t.Run("taxes", func(t *testing.T) {
		db := newTestBadgerDB(t)
		defer db.Close()
		d := newTestTaxes()
		if err := saveItem(t, db, keyForSimpleTaxes(testBaseCNPJ), d); err != nil {
			t.Errorf("expected no error saving partner, got %s", err)
		}
		got, err := simpleTaxesOf(db, testBaseCNPJ)
		if err != nil {
			t.Errorf("expected no error reading taxes, got %s", err)
		}
		if !reflect.DeepEqual(got, d) {
			t.Errorf("expected %v, got %v", d, got)
		}
	})

}
