package db

import (
	"os"
	"testing"
)

func TestMongoDB(t *testing.T) {
	id := "33683111000280"
	json := `{"qsa": [{"name": 42}, {"name": "forty-two"}], "answer": 42}`

	u := os.Getenv("TEST_MONGODB_URL")
	if u == "" {
		t.Errorf("expected a posgres uri at TEST_DATABASE_URL, found nothing")
		return
	}
	db, err := NewMongoDB("minhareceita")
	if err != nil {
		t.Errorf("expected no error connecting to postgres, got %s", err)
		return
	}
	if err := db.DropCollection(); err != nil {
		t.Errorf("expected no error dropping the table, got %s", err)
	}
	defer func() {
		if err := db.DropCollection(); err != nil {
			t.Errorf("expected no error dropping the table, got %s", err)
		}
		db.Close()
	}()

	if err := db.CreateCollection(); err != nil {
		t.Errorf("expected no error creating the table, got %s", err)
	}

	if err := db.CreateCompanies([][]string{{id, json}}); err != nil {
		t.Errorf("expected no error saving a company, got %s", err)
	}
	if err := db.CreateCompanies([][]string{{id, json}}); err != nil {
		t.Errorf("expected no error saving a duplicated company, got %s", err)
	}
	got, err := db.GetCompany("33683111000280")
	if err != nil {
		t.Errorf("expected no error getting a company, got %s", err)
	}
	if got != json {
		t.Errorf("expected json to be %s, got %s", json, got)
	}
	got, err = db.GetCompany("33683111000280")
	if err != nil {
		t.Errorf("expected no error getting a company, got %s", err)
	}
	if got != json {
		t.Errorf("expected json to be %s, got %s", json, got)
	}
	if err := db.MetaSave("answer", "42"); err != nil {
		t.Errorf("expected no error writing to the metadata table, got %s", err)
	}
	metadata, err := db.MetaRead("answer")
	if err != nil {
		t.Errorf("expected no error getting metadata, got %s", err)
	}
	if metadata != "42" {
		t.Errorf("expected 42 as the answer, got %s", metadata)
	}
	if err := db.MetaSave("answer", "forty-two"); err != nil {
		t.Errorf("expected no error re-writing to the metadata table, got %s", err)
	}
	metadata2, err := db.MetaRead("answer")
	if err != nil {
		t.Errorf("expected no error getting metadata for the second time, got %s", err)
	}
	if metadata2 != "forty-two" {
		t.Errorf("expected foruty-two as the answer, got %s", metadata2)
	}
}
