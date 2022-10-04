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

	u := os.Getenv("TEST_DATABASE_URL")
	if u == "" {
		t.Errorf("expected a posgres uri at TEST_DATABASE_URL, found nothing")
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
	if err := pg.CreateCompanies([][]string{{id, json}}); err != nil {
		t.Errorf("expected no error saving a duplicated company, got %s", err)
	}
	if err := pg.CreateIndex(); err != nil {
		t.Errorf("expected no error creating index, got %s", err)
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
	if err := pg.MetaSave("answer", "42"); err != nil {
		t.Errorf("expected no error writing to the metadata table, got %s", err)
	}
	if got := pg.MetaRead("answer"); got != "42" {
		t.Errorf("expected 42 as the answer, got %s", got)
	}
	if err := pg.MetaSave("answer", "fourty-two"); err != nil {
		t.Errorf("expected no error re-writing to the metadata table, got %s", err)
	}
	if got := pg.MetaRead("answer"); got != "fourty-two" {
		t.Errorf("expected foruty-two as the answer, got %s", got)
	}
}
