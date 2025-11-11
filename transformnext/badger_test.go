package transformnext

import (
	"testing"
)

func TestSerializeDeserialize(t *testing.T) {
	kv, err := newBadger(t.TempDir(), false)
	if err != nil {
		t.Errorf("expected no error opening badger, got %s", err)
	}
	for _, tc := range []struct {
		name string
		row  []string
	}{
		{"normal", []string{"um", "dois", "três"}},
		{"empty", []string{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, err := kv.serialize(tc.row)
			if err != nil {
				t.Errorf("expected no error serializing, got %s", err)
			}
			got, err := kv.deserialize(s)
			if err != nil {
				t.Errorf("expected no error deserializing, got %s", err)
			}
			for idx := range len(tc.row) {
				if got[idx] != tc.row[idx] {
					t.Errorf("expected element %d to be %s, got %s", idx+1, tc.row[idx], got[idx])
				}
			}
		})
	}
}

func TestPutGet(t *testing.T) {
	src := &source{prefix: "test"}
	kv, err := newBadger(t.TempDir(), false)
	if err != nil {
		t.Errorf("expected no error opening badger, got %s", err)
	}
	defer func() {
		if err := kv.db.Close(); err != nil {
			t.Errorf("expected no error closing badger, got %s", err)
		}
	}()
	for _, tc := range []struct {
		name string
		id   string
		row  []string
	}{
		{"normal", "1", []string{"um", "dois", "três"}},
		{"empty", "2", []string{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := kv.put(src, tc.id, tc.row)
			if err != nil {
				t.Errorf("expected no error putting row, got %s", err)
			}
			k := src.keyFor(tc.id)
			got, err := kv.get(k)
			if err != nil {
				t.Errorf("expected no error getting row, got %s", err)
			}
			if len(tc.row) == 0 {
				if got != nil {
					t.Errorf("expected value to be nil, got %v", got)
				}
			} else {
				for idx := range tc.row {
					if got[idx] != tc.row[idx] {
						t.Errorf("expected element %d to be %s, got %s", idx+1, tc.row[idx], got[idx])
					}
				}
			}
		})
	}
}
