package transform

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/dgraph-io/badger/v4"
)

const testBaseCNPJ = "12345678"

func newTestBadgerDB(t *testing.T, inMem bool) *badger.DB {
	var opt badger.Options
	if !inMem {
		opt = badger.DefaultOptions(t.TempDir())
	} else {
		opt = badger.DefaultOptions("").WithInMemory(inMem)
	}
	db, err := badger.Open(opt)
	if err != nil {
		t.Fatal("could not create a badger database")
	}
	return db
}

var (
	testIdentificacaoDoSocio1                 = 1
	testCodigoQualificacaoSocio1              = 2
	testQualificacaoSocio1                    = "Dois"
	testCodigoPais1                           = 3
	testPais1                                 = "Três"
	testCodigoQualificacaoRepresentanteLegal1 = 4
	testQualificacaoRepresentanteLegal1       = "Quatro"
	testCodigoFaixaEtaria1                    = 5
	testFaixaEtarua1                          = "Cinco"
	testPartner1                              = partnerData{
		&testIdentificacaoDoSocio1,
		"Nome da pessoa 1",
		"123",
		&testCodigoQualificacaoSocio1,
		&testQualificacaoSocio1,
		nil,
		&testCodigoPais1,
		&testPais1,
		"456",
		"Representante legal 1",
		&testCodigoQualificacaoRepresentanteLegal1,
		&testQualificacaoRepresentanteLegal1,
		&testCodigoFaixaEtaria1,
		&testFaixaEtarua1,
	}

	testIdentificacaoDoSocio2                 = 6
	testCodigoQualificacaoSocio2              = 7
	testQualificacaoSocio2                    = "Sete"
	testCodigoPais2                           = 8
	testPais2                                 = "Oito"
	testCodigoQualificacaoRepresentanteLegal2 = 9
	testQualificacaoRepresentanteLegal2       = "Nove"
	testCodigoFaixaEtaria2                    = 10
	testFaixaEtarua2                          = "Dez"
	testPartner2                              = partnerData{
		&testIdentificacaoDoSocio2,
		"Nome da pessoa 2",
		"789",
		&testCodigoQualificacaoSocio2,
		&testQualificacaoSocio2,
		nil,
		&testCodigoPais2,
		&testPais2,
		"012",
		"Representante legal 2",
		&testCodigoQualificacaoRepresentanteLegal2,
		&testQualificacaoRepresentanteLegal2,
		&testCodigoFaixaEtaria2,
		&testFaixaEtarua2,
	}
)

func toBytes(t *testing.T, i interface{}) []byte {
	b, err := json.Marshal(i)
	if err != nil {
		t.Fatalf("error marshaling %v: %s", i, err)
	}
	return b
}

func TestMergePartners(t *testing.T) {
	k := []byte(testBaseCNPJ)
	p := newTestPartner()
	v := toBytes(t, p)
	for _, inMem := range []bool{true, false} {
		for _, tc := range []struct {
			existing []partnerData
			expected []partnerData
		}{
			{nil, []partnerData{p}},
			{[]partnerData{testPartner1}, []partnerData{testPartner1, p}},
			{[]partnerData{testPartner1, testPartner2}, []partnerData{testPartner1, testPartner2, p}},
		} {
			n := "in disk"
			if inMem {
				n = "in memory"
			}
			t.Run(fmt.Sprintf("merging to %d partners %s", len(tc.existing), n), func(t *testing.T) {
				db := newTestBadgerDB(t, inMem)
				defer db.Close()
				if tc.existing != nil {
					db.Update(func(tx *badger.Txn) error {
						if err := tx.Set(k, toBytes(t, tc.existing)); err != nil {
							t.Fatalf("error setting existing partners %v: %s", tc.existing, err)
						}
						return nil
					})
				}
				m, err := mergePartners(db, k, v)
				if err != nil {
					t.Errorf("expected no error merging partners, got %s", err)
				}
				var got []partnerData
				if err := json.Unmarshal(m, &got); err != nil {
					t.Errorf("could not parse merged partners: %s", err)
				}
				if !reflect.DeepEqual(got, tc.expected) {
					t.Errorf("expected merged partners to be %v, got %v", tc.expected, got)
				}
			})
		}
	}
}

func TestSaveAndReadItems(t *testing.T) {
	for _, inMem := range []bool{true, false} {
		n := "in disk"
		if inMem {
			n = "in memory"
		}

		t.Run(fmt.Sprintf("partners %s", n), func(t *testing.T) {
			p := newTestPartner()
			db := newTestBadgerDB(t, inMem)
			defer db.Close()
			err := saveItem(
				db, partners,
				[]byte(keyForPartners(testBaseCNPJ)),
				toBytes(t, p),
			)
			if err != nil {
				t.Errorf("expected no error saving partner, got %s", err)
			}
			got, err := partnersOf(db, testBaseCNPJ)
			if err != nil {
				t.Errorf("expected no error reading partners, got %s", err)
			}
			if len(got) != 1 {
				t.Errorf("expected merged partnes to have 1 partger, got %d", len(got))
				return
			}
			if !reflect.DeepEqual(got[0], p) {
				t.Errorf("expected merged partner to be %v, got %v", p, got[0])
			}
		})

		t.Run(fmt.Sprintf("base %s", n), func(t *testing.T) {
			db := newTestBadgerDB(t, inMem)
			defer db.Close()
			d := newTestBaseCNPJ()
			v := toBytes(t, d)
			err := saveItem(db, base, []byte(keyForBase(testBaseCNPJ)), v)
			if err != nil {
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

		t.Run(fmt.Sprintf("taxes %s", n), func(t *testing.T) {
			db := newTestBadgerDB(t, inMem)
			defer db.Close()
			d := newTestTaxes()
			v := toBytes(t, d)
			err := saveItem(db, taxes, []byte(keyForTaxes(testBaseCNPJ)), v)
			if err != nil {
				t.Errorf("expected no error saving partner, got %s", err)
			}
			got, err := taxesOf(db, testBaseCNPJ)
			if err != nil {
				t.Errorf("expected no error reading taxes, got %s", err)
			}
			if !reflect.DeepEqual(got, d) {
				t.Errorf("expected %v, got %v", d, got)
			}
		})
	}
}
