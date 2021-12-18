package csv

import (
	"compress/gzip"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateCSV(t *testing.T) {
	tmp := t.TempDir()
	d := filepath.Join(tmp, "33", "683", "111")
	if err := os.MkdirAll(d, 0755); err != nil {
		t.Errorf("expected no error creating %s, got %s", d, err)
	}
	p := filepath.Join(d, "000280.json")
	if err := ioutil.WriteFile(p, []byte(`{"answer":42}`), 0755); err != nil {
		t.Errorf("expected no error creating %s, got %s", p, err)
	}

	if err := CreateCSV(tmp); err != nil {
		t.Errorf("expected no error creating csv, got %s", err)
	}

	f, err := os.Open(filepath.Join(tmp, Path))
	if err != nil {
		t.Errorf("expected no error opening csv file, got %s", err)
	}
	defer f.Close()

	r, err := gzip.NewReader(f)
	if err != nil {
		t.Errorf("expected no errors unarchiving the csv file, got %s", err)
	}
	defer r.Close()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		t.Errorf("expected no errors reading the csv file, got %s", err)
	}

	got := strings.TrimSpace(string(b))
	expected := "id,json\n33683111000280,\"{\"\"answer\"\":42}\""
	if string(got) != expected {
		t.Errorf("\nexpected:\n\n%s\n\ngot:\n\n%s\n\n", expected, got)
	}
}
