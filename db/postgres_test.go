package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPostgresDB(t *testing.T) {
	pg := NewPostgreSQL(os.Getenv("TEST_POSTGRES_URI"))
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
