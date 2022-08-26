package sample

import (
	"os"
	"path/filepath"
	"testing"
)

var testdata = filepath.Join("..", "testdata")

func TestSample(t *testing.T) {
	tmp := t.TempDir()
	if err := Sample(testdata, tmp, 42); err != nil {
		t.Fatalf("expected no error running sample, got %s", err)
	}
	ls, err := os.ReadDir(tmp)
	if err != nil {
		t.Errorf("expected no error readind dir %s, got %s", tmp, err)
	}

	expected := 12
	var got []string
	for _, f := range ls {
		if !f.IsDir() {
			got = append(got, f.Name())
		}
	}
	if len(got) != expected {
		t.Errorf("expected %d files in the sample directory, got %d: %v", expected, len(got), got)
	}
	for _, f := range got {
		_, err := os.Stat(filepath.Join(testdata, f))
		if err != nil {
			t.Errorf("expected %s to exist in %s, but it does not", f, testdata)
		}
	}
}
