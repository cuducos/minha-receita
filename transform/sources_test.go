package transform

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestFilesFor(t *testing.T) {
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
		got, err := filesFor(c.source, dir)
		if err != nil {
			t.Errorf("expected no error for %s, got %s", c.source, err)
		}
		if !reflect.DeepEqual(got, c.expected) {
			t.Errorf("expected %q for %s, got %q", c.expected, c.source, got)
		}
	}
}
