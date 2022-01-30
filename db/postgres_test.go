package db

import (
	"os"
	"testing"
)

func TestPostgresDB(t *testing.T) {
	id := "33683111000280"
	json := `{"answer":42}`
	newJSON := `{"answer":"fourty-two"}`

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

	if err := pg.CreateTable(); err != nil {
		t.Errorf("expected no error creating the table, got %s", err)
	}
	if err := pg.CreateCompanies([][]string{{id, json}}); err != nil {
		t.Errorf("expected no error saving a company, got %s", err)
	}
	got, err := pg.GetCompany("33683111000280")
	if err != nil {
		t.Errorf("expected no error getting a company, got %s", err)
	}
	if got != json {
		t.Errorf("expected json to be %s, got %s", json, got)
	}
	if err := pg.UpdateCompany(id, newJSON); err != nil {
		t.Errorf("expected no error updating a company, got %s", err)
	}
	got, err = pg.GetCompany("33683111000280")
	if err != nil {
		t.Errorf("expected no error getting a company, got %s", err)
	}
	if got != newJSON {
		t.Errorf("expected json to be %s, got %s", newJSON, got)
	}
	list, err := pg.ListCompanies(id[:8])
	if err != nil {
		t.Errorf("expected no error listing companies, got %s", err)
	}
	if len(list) != 1 {
		t.Errorf("expected list to have 1 company, got %d", len(list))
	}
	if err := pg.DropTable(); err != nil {
		t.Errorf("expected no error dropping the table, got %s", err)
	}
}
