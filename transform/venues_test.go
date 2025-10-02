package transform

import "testing"

func TestTaskRun(t *testing.T) {
	db := newTestDB()
	kv, err := newBadgerStorage(t.TempDir(), false)
	if err != nil {
		t.Errorf("expected no error creating badger, got %s", err)
	}
	defer func() {
		if err := kv.close(); err != nil {
			t.Errorf("expected no error closing key-value storage, got %s", err)
		}
	}()
	lookups, err := newLookups(testdata)
	if err != nil {
		t.Errorf("expected no errors creating look up tables, got %v", err)
	}
	if err := kv.load(testdata, &lookups, 1024); err != nil {
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
