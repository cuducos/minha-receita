package db

import (
	"os"
	"path/filepath"
	"testing"
)

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
	if err := pg.CreateCompanies([][]string{{id, c}}); err != nil {
		t.Errorf("expected no error saving a duplicated company, got %s", err)
	}
	if err := pg.PostLoad(); err != nil {
		t.Errorf("expected no error post load, got %s", err)
	}
	got, err := pg.GetCompany("33683111000280")
	if err != nil {
		t.Errorf("expected no error getting a company, got %s", err)
	}
	assertCompaniesAreEqual(t, got, c)
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
}
