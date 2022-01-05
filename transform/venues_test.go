package transform

import (
	"testing"
)

func TestTaskRun(t *testing.T) {
	db := newMockDB()
	r, err := createJSONRecordsTask(testdata, &db, 2)
	if err != nil {
		t.Errorf("expected no error creating task, got %s", err)
	}
	if err = r.run(2); err != nil {
		t.Errorf("expected no error running task, got %s", err)
	}
	if len(db.storage) != 1 {
		t.Errorf("expected 1 company, got %d", len(db.storage))
	}
}
