package transform

import (
	"errors"

	"path/filepath"
	"strings"
	"sync"
)

var testdata = filepath.Join("..", "testdata")

type mockDB struct {
	storage map[string]string
	mu      sync.Mutex
}

func (db *mockDB) SaveCompany(k, v string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.storage[k] = v
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
