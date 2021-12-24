package transform

import (
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
		{venues, []string{filepath.Join(dir, "K3241.K03200Y8.D11009.ESTABELE.zip")}},
		{motives, []string{filepath.Join(dir, "F.K03200$Z.D11009.MOTICSV.zip")}},
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

func TestPathForCNPJ(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		tt := []struct {
			cnpj     string
			expected string
		}{
			{"19131243000197", "19/131/243/000197.json"},
			{"19.131.243/0001-97", "19/131/243/000197.json"},
		}
		for _, c := range tt {
			got, err := PathForCNPJ(c.cnpj)
			if got != c.expected {
				t.Errorf("expected path for %s to be %s, got %s", c.cnpj, c.expected, got)
			}
			if err != nil {
				t.Errorf("expected no error for %s, got %s", c.cnpj, err)
			}
		}
	})
	t.Run("with error", func(t *testing.T) {
		tt := []string{
			"foobar",
			"12345678901234",
			"12.345.678/9012-34",
		}
		for _, c := range tt {
			_, err := PathForCNPJ(c)
			if err == nil {
				t.Errorf("expected an error for %s, got nil", c)
			}
		}
	})
}

func TestPathForBaseCNPJ(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		b := "19131243"
		expected := "19/131/243"
		got, err := pathForBaseCNPJ(b)
		if got != expected {
			t.Errorf("expected path for %s to be %s, got %s", b, expected, got)
		}
		if err != nil {
			t.Errorf("expected no error for %s, got %s", b, err)
		}
	})
	t.Run("with error", func(t *testing.T) {
		b := "19.131.243"
		_, err := pathForBaseCNPJ(b)
		if err == nil {
			t.Errorf("expected error for %s, got nil", b)
		}
	})
}

func TestCNPJForPath(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		tt := []struct {
			path     string
			expected string
		}{
			{"19/131/243/000197.json", "19131243000197"},
			{"/home/user/data/19/131/243/000197.json", "19131243000197"},
		}
		for _, c := range tt {
			got, err := CNPJForPath(c.path)
			if got != c.expected {
				t.Errorf("expected cnpj for %s to be %s, got %s", c.path, c.expected, got)
			}
			if err != nil {
				t.Errorf("expected no error for %s, got %v", c.path, err)
			}
		}
	})
	t.Run("with error", func(t *testing.T) {
		c := "19/131.json"
		_, err := CNPJForPath(c)
		if err == nil {
			t.Errorf("expected an error for %s, got nil", c)
		}
	})
}
