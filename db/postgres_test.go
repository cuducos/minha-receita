package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/cuducos/minha-receita/testutils"
)

var postgresDefaultIndexes = []string{"cnpj_pkey", "cnpj_id"}

func setUpPostgres(id, c string) (*PostgreSQL, error) {
	u := os.Getenv("TEST_POSTGRES_URL")
	if u == "" {
		return nil, fmt.Errorf("expected a posgres uri at TEST_POSTGRES_URL, found nothing")
	}
	db, err := NewPostgreSQL(u, "public")
	if err != nil {
		return nil, fmt.Errorf("expected no error connecting to postgres, got %w", err)
	}
	if err := db.Drop(); err != nil {
		return nil, fmt.Errorf("expected no error dropping the tables, got %w", err)
	}
	if err := db.Create(); err != nil {
		return nil, fmt.Errorf("expected no error creating the tables, got %w", err)
	}
	if err := db.PreLoad(); err != nil {
		return nil, fmt.Errorf("expected no error pre load on postgres, got %w", err)
	}
	if err := db.CreateCompanies([][]string{{id, c}}); err != nil {
		return nil, fmt.Errorf("expected no error saving a company to postgres, got %w", err)
	}
	if err := db.PostLoad(); err != nil {
		return nil, fmt.Errorf("expected no error post load on postgres, got %w", err)
	}
	return &db, nil
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

func TestPostgresCreateIndexes(t *testing.T) {
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
	i := []string{"qsa.nome_socio"}
	if err := pg.CreateExtraIndexes(i); err != nil {
		t.Errorf("expected no errors running extra indexes, got %s", err)
	}
	testutils.AssertArraysHaveSameItems(t, i, listIndexesPostgres(t, pg))
}
