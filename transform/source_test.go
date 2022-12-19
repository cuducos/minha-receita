package transform

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestPathsForSource(t *testing.T) {
	tc := []struct {
		source   sourceType
		expected []string
	}{
		{venues, []string{
			filepath.Join(testdata, "Estabelecimentos0.zip"),
		}},
		{motives, []string{filepath.Join(testdata, "Motivos.zip")}},
		{base, []string{
			filepath.Join(testdata, "Empresas0.zip"),
			filepath.Join(testdata, "Empresas1.zip"),
		}},
	}
	for _, c := range tc {
		got, err := pathsForSource(c.source, testdata)
		if err != nil {
			t.Errorf("expected no error for %s, got %s", c.source, err)
		}
		if !reflect.DeepEqual(got, c.expected) {
			t.Errorf("expected %q for %s, got %q", c.expected, c.source, got)
		}
	}
}

func TestSource(t *testing.T) {
	s, err := newSource(base, testdata)

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
