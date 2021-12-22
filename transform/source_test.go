package transform

import (
	"path/filepath"
	"testing"
)

func TestSource(t *testing.T) {
	s, err := newSource(base_cpnj, filepath.Join("..", "testdata"))

	if err != nil {
		t.Errorf("expected no error creating a source, got: %s", err)
	}
	if len(s.files) != 2 {
		t.Errorf("expected a source with 2 files, got %d", len(s.files))
	}
	if len(s.readers) != 2 {
		t.Errorf("expected a source with 2 readers, got %d", len(s.readers))
	}
	if s.totalLines != 2 {
		t.Errorf("expected a source with 2 lines, got %d", s.totalLines)
	}
}
