package db

import (
	"os"
	"testing"
)

func TestPostgresDB(t *testing.T) {
	id := "33683111000280"
	json := `{"qsa": null, "answer": 42}`
	newJSON := `{"again": "fourty-two"}`
	partner1 := `[{"name": 42}]`
	partner2 := `[{"name":  "fourty-two"}]`
	expected := `{"qsa": [{"name": 42}, {"name": "fourty-two"}], "again": "fourty-two", "answer": 42}`

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
	defer func() {
		if err := pg.DropTable(); err != nil {
			t.Errorf("expected no error dropping the table, got %s", err)
		}
		pg.Close()
	}()

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
	if err := pg.UpdateCompanies([][]string{{id[:8], newJSON}}); err != nil {
		t.Errorf("expected no error updating a company, got %s", err)
	}
	if err := pg.AddPartners([][]string{{id[:8], partner1}, {id[:8], partner2}}); err != nil {
		t.Errorf("expected no error adding partners, got %s", err)
	}
	got, err = pg.GetCompany("33683111000280")
	if err != nil {
		t.Errorf("expected no error getting a company, got %s", err)
	}
	if got != expected {
		t.Errorf("expected json to be %s, got %s", expected, got)
	}
}
