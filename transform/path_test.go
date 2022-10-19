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
		got, err := PathsForSource(c.source, testdata)
		if err != nil {
			t.Errorf("expected no error for %s, got %s", c.source, err)
		}
		if !reflect.DeepEqual(got, c.expected) {
			t.Errorf("expected %q for %s, got %q", c.expected, c.source, got)
		}
	}
}
