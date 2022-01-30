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
		{venues, []string{filepath.Join(testdata, "K3241.K03200Y8.D11009.ESTABELE.zip")}},
		{motives, []string{filepath.Join(testdata, "F.K03200$Z.D11009.MOTICSV.zip")}},
		{base, []string{
			filepath.Join(testdata, "K3241.K03200Y5.D11009.EMPRECSV.zip"),
			filepath.Join(testdata, "K3241.K03200Y8.D11009.EMPRECSV.zip"),
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
