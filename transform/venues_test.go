package transform

import (
	"testing"
)

func TestTaskRun(t *testing.T) {
	db := newTestDB(t)
	r, err := createJSONRecordsTask(testdata, db, 2)
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
