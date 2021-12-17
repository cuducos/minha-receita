package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPostgresDB(t *testing.T) {
	u := os.Getenv("TEST_POSTGRES_URI")
	if u == "" {
		t.Errorf("expected a posgres uri at TEST_POSTGRES_URI, found nothing")
		return
	}
	pg, err := NewPostgreSQL(u)
	if err != nil {
		t.Errorf("expected no error connecting to postgres, got %s", err)
		return
	}
	defer pg.Close()
	if err := pg.CreateTable(); err != nil {
		t.Errorf("expected no error creating the table, got %s", err)
	}
	dir := filepath.Join("..", "testdata")
	if err := pg.ImportData(dir); err != nil {
		t.Errorf("expected no error importing data, got %s", err)
	}
	got, err := pg.GetCompany("42")
	if err != nil {
		t.Errorf("expected no error getting a company, got %s", err)
	}
	expected := `{"fourty": "two"}`
	if got != expected {
		t.Errorf("expected json to be %s, got %s", expected, got)
	}
	if err := pg.DropTable(); err != nil {
		t.Errorf("expected no error dropping the table, got %s", err)
	}
}
