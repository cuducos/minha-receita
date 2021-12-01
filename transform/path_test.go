package transform

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"
)

func TestPathsForSource(t *testing.T) {
	dir := filepath.Join("..", "testdata")
	tc := []struct {
		source   sourceType
		expected []string
	}{
		{venue, []string{filepath.Join(dir, "K3241.K03200Y8.D11009.ESTABELE.zip")}},
		{motive, []string{filepath.Join(dir, "F.K03200$Z.D11009.MOTICSV.zip")}},
		{main, []string{
			filepath.Join(dir, "K3241.K03200Y5.D11009.EMPRECSV.zip"),
			filepath.Join(dir, "K3241.K03200Y8.D11009.EMPRECSV.zip"),
		}},
	}
	for _, c := range tc {
		got, err := PathsForSource(c.source, dir)
		if err != nil {
			t.Errorf("expected no error for %s, got %s", c.source, err)
		}
		if !reflect.DeepEqual(got, c.expected) {
			t.Errorf("expected %q for %s, got %q", c.expected, c.source, got)
		}
	}
}

func TestPathFor(t *testing.T) {
	tt := []struct {
		cnpj string
		path string
		err  error
	}{
		{"19131243000197", "19131243/000197.json", nil},
		{"19.131.243/0001-97", "19131243/000197.json", nil},
		{"foobar", "", InvalidCNPJError},
		{"12345678901234", "", InvalidCNPJError},
		{"12.345.678/9012-34", "", InvalidCNPJError},
	}
	for _, c := range tt {
		got, err := PathForCNPJ(c.cnpj)
		if got != c.path {
			t.Errorf("expected path for %s to be %s, got %s", c.cnpj, c.path, got)
		}
		if !errors.Is(err, c.err) {
			t.Errorf("expected %v as an error for %s, got %v", c.err, c.cnpj, err)
		}
	}
}

func TestCNPJFor(t *testing.T) {
	tt := []struct {
		path string
		cnpj string
		err  error
	}{
		{"19131243/000197.json", "19131243000197", nil},
		{"/home/user/data/19131243/000197.json", "19131243000197", nil},
		{"19/131.json", "", InvalidPathError},
	}
	for _, c := range tt {
		got, err := CNPJForPath(c.path)
		if got != c.cnpj {
			t.Errorf("expected cnpj for %s to be %s, got %s", c.path, c.cnpj, got)
		}
		if !errors.Is(err, c.err) {
			t.Errorf("expected %v as an error for %s, got %v", c.err, c.path, err)
		}
	}
}
