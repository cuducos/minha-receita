package db

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/cuducos/minha-receita/transform"
)

type database interface {
	Create() error
	Drop() error
	PreLoad() error
	PostLoad() error
	Close()

	CreateCompanies([][]string) error
	GetCompany(string) (string, error)

	CreateExtraIndexes([]string) error
	Search(context.Context, *Query) (string, error)

	MetaSave(string, string) error
	MetaRead(string) (string, error)
}

type testCase struct {
	params   url.Values
	expected int
}

func (tc *testCase) name(db database) string {
	return fmt.Sprintf("%T %s expecting %d", db, tc.params.Encode(), tc.expected)
}

type page struct {
	Data   []transform.Company `json:"data"`
	Cursor *string             `json:"cursor"`
}

func assertSearchCount(t *testing.T, s string, tc testCase) {
	var p page
	if err := json.Unmarshal([]byte(s), &p); err != nil {
		t.Errorf("expected no error deserializing JSON, got %s", err)
		return
	}
	if got := len(p.Data); got != tc.expected {
		t.Errorf("expected %d results for %v, got %d", tc.expected, tc.params.Encode(), got)
	}
}

func assertCompaniesAreEqual(t *testing.T, s1 string, s2 string) {
	toCompany := func(s string) transform.Company {
		var c transform.Company
		if err := json.Unmarshal([]byte(s), &c); err != nil {
			t.Errorf("expected no error unmarshalling company, got %s", err)
		}
		return c
	}
	c1 := toCompany(s1)
	c2 := toCompany(s2)
	if !reflect.DeepEqual(c1, c2) {
		t.Errorf("expected companies to be equal, got %s and %s", s1, s2)
	}
}

func TestRetrieve(t *testing.T) {
	id := "33683111000280"
	b, err := os.ReadFile(filepath.Join("..", "testdata", "response.json"))
	if err != nil {
		t.Error("error reading company JSON file")
	}
	c := string(b)
	pg, err := setUpPostgres(id, c)
	if err != nil {
		t.Errorf("expected no error setting up postgres, got %s", err)
		return
	}
	defer func() {
		if err := pg.Drop(); err != nil {
			t.Errorf("expected no error dropping the tables, got %s", err)
		}
		pg.Close()
	}()
	m, err := setUpMongo(id, c)
	if err != nil {
		t.Errorf("expected no error setting up mongo, got %s", err)
		return
	}
	defer func() {
		if err := m.Drop(); err != nil {
			t.Errorf("expected no error dropping the collections, got %s", err)
		}
		m.Close()
	}()
	for _, db := range []database{pg, m} {
		t.Run(fmt.Sprintf("%T", db), func(t *testing.T) {
			got, err := db.GetCompany("33683111000280")
			if err != nil {
				t.Errorf("expected no error getting a company, got %s", err)
			}
			assertCompaniesAreEqual(t, got, c)
			if err := db.MetaSave("answer", "42"); err != nil {
				t.Errorf("expected no error writing to the metadata table, got %s", err)
			}
			m1, err := db.MetaRead("answer")
			if err != nil {
				t.Errorf("expected no error getting metadata, got %s", err)
			}
			if m1 != "42" {
				t.Errorf("expected 42 as the answer, got %s", m1)
			}
			if err := db.MetaSave("answer", "forty-two"); err != nil {
				t.Errorf("expected no error re-writing to the metadata table, got %s", err)
			}
			m2, err := db.MetaRead("answer")
			if err != nil {
				t.Errorf("expected no error getting metadata for the second time, got %s", err)
			}
			if m2 != "forty-two" {
				t.Errorf("expected foruty-two as the answer, got %s", m2)
			}
			if err := db.CreateExtraIndexes([]string{"teste.index1"}); err == nil {
				t.Error("expected errors running extra indexes, got nil")
			}
		})
	}
}

func TestSearch(t *testing.T) {
	id := "33683111000280"
	b, err := os.ReadFile(filepath.Join("..", "testdata", "response.json"))
	if err != nil {
		t.Error("error reading company JSON file")
	}
	c := string(b)
	pg, err := setUpPostgres(id, c)
	if err != nil {
		t.Errorf("expected no error setting up postgres, got %s", err)
		return
	}
	defer func() {
		if err := pg.Drop(); err != nil {
			t.Errorf("expected no error dropping the tables, got %s", err)
		}
		pg.Close()
	}()
	m, err := setUpMongo(id, c)
	if err != nil {
		t.Errorf("expected no error setting up mongo, got %s", err)
		return
	}
	defer func() {
		if err := m.Drop(); err != nil {
			t.Errorf("expected no error dropping the collections, got %s", err)
		}
		m.Close()
	}()
	for _, tc := range []testCase{
		{map[string][]string{"uf": {"sc"}}, 0},
		{map[string][]string{"uf": {"sp"}}, 1},
		{map[string][]string{"uf": {"sp", "sc"}}, 1},
		{map[string][]string{"municipio": {"6105"}}, 0},
		{map[string][]string{"municipio": {"3500303"}}, 0},
		{map[string][]string{"municipio": {"7107"}}, 1},
		{map[string][]string{"municipio": {"3550308"}}, 1},
		{map[string][]string{"municipio": {"3550308", "3500303"}}, 1},
		{map[string][]string{"municipio": {"6105", "7107"}}, 1},
		{map[string][]string{"natureza_juridica": {"2143"}}, 0},
		{map[string][]string{"natureza_juridica": {"3999"}}, 1},
		{map[string][]string{"natureza_juridica": {"3999", "2143"}}, 1},
		{map[string][]string{"cnae_fiscal": {"6204000"}}, 0},
		{map[string][]string{"cnae_fiscal": {"9430800"}}, 1},
		{map[string][]string{"cnae_fiscal": {"9430800", "6204000"}}, 1},
		{map[string][]string{"cnae": {"722702"}}, 0},
		{map[string][]string{"cnae": {"6204000"}}, 1},
		{map[string][]string{"cnae": {"9430800", "6204000"}}, 1},
		{map[string][]string{"cnpf": {"21449073000135"}}, 0},
		{map[string][]string{"cnpf": {"***112108**"}}, 1},
		{map[string][]string{"cnpf": {"21449073000135", "***112108**"}}, 1},
	} {
		for _, db := range []database{pg, m} {
			t.Run(tc.name(db), func(t *testing.T) {
				q := NewQuery(tc.params)
				s, err := db.Search(context.Background(), q)
				if err != nil {
					t.Errorf("expected no error searching, got %s", err)
					return
				}
				assertSearchCount(t, s, tc)
			})
		}
	}
}
