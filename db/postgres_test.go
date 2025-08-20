package db

import (
	"context"
	"encoding/json/v2"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/cuducos/minha-receita/testutils"
	"github.com/cuducos/minha-receita/transform"
)

var postgresDefaultIndexes = []string{"cnpj_pkey", "cnpj_id"}

type page struct {
	Data   []transform.Company `json:"data"`
	Cursor *string             `json:"cursor"`
}

func listIndexesPostgres(t *testing.T, pg *PostgreSQL) []string {
	q := `
		SELECT indexname
		FROM pg_indexes
		WHERE tablename = $1 AND schemaname = 'public'
	`
	c := context.Background()
	r, err := pg.pool.Query(c, q, pg.CompanyTableName)
	if err != nil {
		t.Errorf("expected no errors checking index list, got %s", err)
		return nil
	}
	defer r.Close()
	var i []string
	for r.Next() {
		var iname string
		if err := r.Scan(&iname); err != nil {
			t.Errorf("expected no error scanning index name, got %s", err)
			continue
		}
		if !slices.Contains(postgresDefaultIndexes, iname) {
			i = append(i, strings.TrimPrefix(iname, "idx_json."))
		}
	}
	return i
}

func TestPostgresDB(t *testing.T) {
	id := "33683111000280"
	b, err := os.ReadFile(filepath.Join("..", "testdata", "response.json"))
	if err != nil {
		t.Error("error reading company JSON file")
	}
	c := string(b)

	u := os.Getenv("TEST_POSTGRES_URL")
	if u == "" {
		t.Errorf("expected a posgres uri at TEST_POSTGRES_URL, found nothing")
		return
	}
	pg, err := NewPostgreSQL(u, "public")
	if err != nil {
		t.Errorf("expected no error connecting to postgres, got %s", err)
		return
	}
	if err := pg.Drop(); err != nil {
		t.Errorf("expected no error dropping the table, got %s", err)
	}
	defer func() {
		if err := pg.Drop(); err != nil {
			t.Errorf("expected no error dropping the table, got %s", err)
		}
		pg.Close()
	}()

	if err := pg.Create(); err != nil {
		t.Errorf("expected no error creating the table, got %s", err)
	}
	if err := pg.PreLoad(); err != nil {
		t.Errorf("expected no error pre load, got %s", err)
	}
	if err := pg.CreateCompanies([][]string{{id, c}}); err != nil {
		t.Errorf("expected no error saving a company, got %s", err)
	}
	if err := pg.PostLoad(); err != nil {
		t.Errorf("expected no error post load, got %s", err)
	}
	got, err := pg.GetCompany("33683111000280")
	if err != nil {
		t.Errorf("expected no error getting a company, got %s", err)
	}
	assertCompaniesAreEqual(t, got, c)
	var q Query
	q.Limit = 1
	q.Municipio = []uint32{3500303}
	sr, err := pg.Search(context.Background(), &q)
	if err != nil {
		t.Errorf("expected no error querying %#v, got %s", q, err)
	}
	var r page
	if err := json.Unmarshal([]byte(sr), &r); err != nil {
		t.Errorf("expected no error deserializing JSON, got %s", err)
	}
	if len(r.Data) != 0 {
		t.Errorf("expected error no result, got %#v", r)
	}
	q.Municipio = nil
	q.UF = []string{"SP"}
	sr, err = pg.Search(context.Background(), &q)
	if err != nil {
		t.Errorf("expected no error querying %#v, got %s", q, err)
	}
	if err := json.Unmarshal([]byte(sr), &r); err != nil {
		t.Errorf("expected no error deserializing JSON, got %s", err)
	}
	if len(r.Data) != 1 {
		t.Errorf("expected one result, got %d", len(r.Data))
	}
	if r.Data[0].UF != "SP" {
		t.Errorf("expected query result to be from SP, got %s", r.Data[0].UF)
	}
	if err := pg.MetaSave("answer", "42"); err != nil {
		t.Errorf("expected no error writing to the metadata table, got %s", err)
	}
	metadata, err := pg.MetaRead("answer")
	if err != nil {
		t.Errorf("expected no error getting metadata, got %s", err)
	}
	if metadata != "42" {
		t.Errorf("expected 42 as the answer, got %s", metadata)
	}
	if err := pg.MetaSave("answer", "forty-two"); err != nil {
		t.Errorf("expected no error re-writing to the metadata table, got %s", err)
	}
	metadata2, err := pg.MetaRead("answer")
	if err != nil {
		t.Errorf("expected no error getting metadata for the second time, got %s", err)
	}
	if metadata2 != "forty-two" {
		t.Errorf("expected foruty-two as the answer, got %s", metadata2)
	}
	if err := pg.CreateExtraIndexes([]string{"teste.index1"}); err == nil {
		t.Error("expected errors running extra indexes, got nil")
	}
	i := []string{"qsa.nome_socio"}
	if err := pg.CreateExtraIndexes(i); err != nil {
		t.Errorf("expected no errors running extra indexes, got %s", err)
	}
	testutils.AssertArraysHaveSameItems(t, i, listIndexesPostgres(t, &pg))
}
