package transform

import (
	"encoding/json/v2"
	"fmt"
	"path/filepath"
	"sync"
)

var (
	testdata = filepath.Join("..", "testdata")
)

func companyFromString(j string) (Company, error) {
	var c Company
	if err := json.Unmarshal([]byte(j), &c); err != nil {
		return Company{}, fmt.Errorf("error unmarshalling: %w", err)
	}
	return c, nil
}

type storage struct {
	data map[string]string
	lock sync.RWMutex
}

type inMemoryDB struct {
	cnpj *storage
	meta *storage
}

func (i inMemoryDB) PreLoad() error                    { return nil }
func (i inMemoryDB) PostLoad() error                   { return nil }
func (i inMemoryDB) CreateExtraIndexes([]string) error { return nil }

func (i inMemoryDB) CreateCompanies(cs [][]string) error {
	i.cnpj.lock.Lock()
	defer i.cnpj.lock.Unlock()
	for _, c := range cs {
		i.cnpj.data[c[0]] = c[1]
	}
	return nil
}

func (i inMemoryDB) MetaSave(k, v string) error {
	i.meta.lock.Lock()
	defer i.meta.lock.Unlock()
	i.meta.data[k] = v
	return nil
}

func (i inMemoryDB) GetCompany(n string) (string, error) {
	i.cnpj.lock.RLock()
	defer i.cnpj.lock.RUnlock()
	if c, ok := i.cnpj.data[n]; ok {
		return c, nil
	}
	return "", fmt.Errorf("company %s not found", n)
}

func newTestDB() inMemoryDB {
	return inMemoryDB{
		cnpj: &storage{data: make(map[string]string)},
		meta: &storage{data: make(map[string]string)},
	}
}
