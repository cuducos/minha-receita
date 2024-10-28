package transform

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestTaskRun(t *testing.T) {
	db := newTestDB(t)
	tmp, err := os.MkdirTemp("", fmt.Sprintf("%s-%s", badgerFilePrefix, time.Now().Format("20060102150405")))
	if err != nil {
		t.Fatal("error creating temporary key-value storage: %w", err)
	}
	defer os.RemoveAll(tmp)
	kv, err := newBadgerStorage(tmp)
	if err != nil {
		t.Errorf("expected no error creating badger, got %s", err)
	}
	defer kv.close(false)
	lookups, err := newLookups(testdata)
	if err != nil {
		t.Errorf("expected no errors creating look up tables, got %v", err)
	}
	if err := kv.load(testdata, &lookups); err != nil {
		t.Errorf("expected no error loading values to badger, got %s", err)
	}
	r, err := createJSONRecordsTask(testdata, db, &lookups, kv, 2, false)
	if err != nil {
		t.Errorf("expected no error creating task, got %s", err)
	}
	if err = r.run(2); err != nil {
		t.Errorf("expected no error running task, got %s", err)
	}
	expected := "33683111000280"
	s, err := db.GetCompany(expected)
	if err != nil {
		t.Errorf("expected no error getting the created company, got %s", err)
	}
	c, err := companyFromString(s)
	if err != nil {
		t.Errorf("expected no error converting company's string to struct, got %s", err)
	}
	if c.CNPJ != expected {
		t.Errorf("expected cnpj to be %s, got %s", expected, c.CNPJ)
	}
}
