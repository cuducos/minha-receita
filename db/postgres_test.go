package db

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cuducos/minha-receita/transform"
)

func TestPostgresDB(t *testing.T) {
	expected := `{"answer":42}`
	cnpj := "33683111000280"

	// create the JSON file fixture
	pth, err := transform.PathForCNPJ(cnpj)
	if err != nil {
		t.Errorf("expected no errors getting the path for the fixture file, got %s", err)
		return
	}
	dir := t.TempDir()
	pth = filepath.Join(dir, pth)
	if err := os.MkdirAll(filepath.Dir(pth), 0755); err != nil {
		t.Errorf("expected no errors creating fixture directory %s, got %s", filepath.Dir(pth), err)
		return
	}
	if err := ioutil.WriteFile(pth, []byte(expected), 0755); err != nil {
		t.Errorf("expected no error writing fixture file %s, got %s", pth, err)
		return
	}

	// connect to the tyest database
	u := os.Getenv("TEST_POSTGRES_URI")
	if u == "" {
		t.Errorf("expected a posgres uri at TEST_POSTGRES_URI, found nothing")
		return
	}
	pg, err := NewPostgreSQL(u, "public")
	if err != nil {
		t.Errorf("expected no error connecting to postgres, got %s", err)
		return
	}
	defer pg.Close()

	// assertions
	if err := pg.CreateTable(); err != nil {
		t.Errorf("expected no error creating the table, got %s", err)
	}
	if err := pg.ImportData(dir); err != nil {
		t.Errorf("expected no error importing data, got %s", err)
	}
	got, err := pg.GetCompany("33683111000280")
	if err != nil {
		t.Errorf("expected no error getting a company, got %s", err)
	}
	if got != expected {
		t.Errorf("expected json to be %s, got %s", expected, got)
	}
	if err := pg.DropTable(); err != nil {
		t.Errorf("expected no error dropping the table, got %s", err)
	}
}
