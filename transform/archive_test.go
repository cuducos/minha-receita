package transform

import (
	"io"
	"path/filepath"
	"reflect"
	"testing"
)

var path = filepath.Join("..", "testdata", "F.K03200$Z.D11009.MOTICSV.zip")

func TestArchivedCSV(t *testing.T) {
	t.Run("read", func(t *testing.T) {
		expected := [][]string{
			{"00", "SEM MOTIVO"},
			{"01", "EXTINCAO POR ENCERRAMENTO LIQUIDACAO VOLUNTARIA"},
		}

		var got [][]string
		z, err := newArchivedCSV(path, separator)
		if err != nil {
			t.Errorf("error creating archived CSV for the test: %s", err)
		}
		for {
			line, err := z.read()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Errorf("error reading archived CSV %s: %s", z.path, err)
			}
			got = append(got, line)
		}

		if !reflect.DeepEqual(expected, got) {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})

	t.Run("close", func(t *testing.T) {
		z, err := newArchivedCSV(path, separator)
		if err != nil {
			t.Errorf("error creating archived CSV for the test: %s", err)
		}

		if err := z.close(); err != nil {
			t.Errorf("expected no error when closing archived CSV, got %s", err)
		}
	})
}

func TestArchivedCSVToLookup(t *testing.T) {
	expected := make(lookup)
	expected[0] = "SEM MOTIVO"
	expected[1] = "EXTINCAO POR ENCERRAMENTO LIQUIDACAO VOLUNTARIA"

	z, err := newArchivedCSV(path, separator)
	if err != nil {
		t.Errorf("expected no error creating an archivedCSV with %s, got %s", path, err)
	}
	defer z.close()

	got, err := z.toLookup()
	if err != nil {
		t.Errorf("expected no error with %s, got %s", path, err)
	}

	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected %q, got %q", expected, got)
	}
}
