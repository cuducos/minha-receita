package transform

import (
	"encoding/json"
	"errors"
	"fmt"

	"path/filepath"
	"strings"
	"sync"

	"github.com/cuducos/go-cnpj"
)

var (
	testdata = filepath.Join("..", "testdata")
)

type mockDB struct {
	storage map[string]string
	mu      sync.Mutex
}

func (db *mockDB) hasKey(k string) bool {
	_, ok := db.storage[k]
	return ok
}

func (db *mockDB) UpdateCompany(k, v string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if !db.hasKey(k) {
		return fmt.Errorf("company %s not found", cnpj.Mask(k))
	}

	db.storage[k] = v
	return nil
}

func (db *mockDB) CreateCompanies(b [][]string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	for _, c := range b {
		if db.hasKey(c[0]) {
			return fmt.Errorf("company %s already exists", cnpj.Mask(c[0]))
		}
		db.storage[c[0]] = c[1]
	}
	return nil
}

func (db *mockDB) GetCompany(k string) (string, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	v, ok := db.storage[k]
	if !ok {
		return "", errors.New("not found")
	}
	return v, nil
}

func (db *mockDB) AddPartner(b, j string) error {
	var p partner
	err := json.Unmarshal([]byte(j), &p)
	if err != nil {
		return err
	}

	db.mu.Lock()
	defer db.mu.Unlock()
	for k, v := range db.storage {
		if strings.HasPrefix(k, b) {
			c, err := companyFromString(v)
			if err != nil {
				return err
			}
			c.QuadroSocietario = append(c.QuadroSocietario, p)
			s, err := c.JSON()
			if err != nil {
				return err
			}
			db.storage[k] = s
		}
	}
	return nil
}

func (db *mockDB) ListCompanies(b string) ([]string, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	var r []string
	for k, v := range db.storage {
		if strings.HasPrefix(k, b) {
			r = append(r, v)
		}
	}
	return r, nil
}

func newMockDB() mockDB {
	return mockDB{storage: make(map[string]string)}
}
