package download

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestGetURLs(t *testing.T) {
	for _, tc := range []struct {
		name     string
		fixture  string
		handler  getURLsHandler
		expected int
	}{
		{"federal revenue", "cadastro-nacional-de-pessoa-juridica-cnpj.json", federalRevenueGetURLs, 37},
		{"national treasure", "national-treasure.json", nationalTreasureGetURLs, 1},
	} {
		ts := httpTestServer(t, tc.fixture)
		defer ts.Close()
		got, err := getURLs(ts.Client(), getFilesConfig{tc.handler, ts.URL}, t.TempDir())
		if err != nil {
			t.Errorf("expected to run withour errors, got: %v:", err)
			return
		}
		if len(got) != tc.expected {
			t.Errorf("expected %d url(s) from the %s, got %d", tc.expected, tc.name, len(got))
		}
	}
}

func TestGetFiles(t *testing.T) {
	ts := httpTestServer(t, "cadastro-nacional-de-pessoa-juridica-cnpj.json")
	defer ts.Close()
	tmp := t.TempDir()
	got, err := getFiles(ts.Client(), getFilesConfig{federalRevenueGetURLs, ts.URL}, tmp, false)
	if err != nil {
		t.Errorf("Expected getFiles to run withour errors, got: %v:", err)
		return
	}
	expected := 37
	if expected != len(got) {
		t.Errorf("Expected getFiles to return %d files, got %d", expected, len(got))
	}
	for _, f := range got {
		if g := filepath.Dir(f.path); g != tmp {
			t.Errorf("Expected %s parent to be %s, got %s", f.path, tmp, g)
		}
		if g := filepath.Base(f.path); !strings.HasSuffix(f.url, g) {
			t.Errorf("Unexpected file name for %s: %s", f.url, g)
		}
	}
}

func TestGetSizes(t *testing.T) {
	f := "cadastro-nacional-de-pessoa-juridica-cnpj.json"
	ts := httpTestServer(t, f)
	defer ts.Close()
	url := ts.URL + "/" + f
	fs := []file{{url: url}}
	got, err := getSizes(ts.Client(), fs, false)
	if err != nil {
		t.Errorf("Expected getSizes to run withour errors, got: %v:", err)
		return
	}
	expected := int64(25398)
	for _, g := range got {
		if g.url == url && g.size != expected {
			t.Errorf("Expected %s size to be %d, got: %d", f, expected, g.size)
		}
	}
}
