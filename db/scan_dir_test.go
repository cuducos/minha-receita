package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsValidJSON(t *testing.T) {
	tt := []struct {
		pth      string
		expected bool
	}{
		{filepath.Join("33", "683", "111", "000280.json"), true},
		{filepath.Join(".gitkeep"), false},
	}
	for _, tc := range tt {
		if got := isValidJSON(tc.pth); got != tc.expected {
			t.Errorf("expected %s to be %t, got %t", tc.pth, tc.expected, got)
		}
	}
}

func TestIsDir(t *testing.T) {
	tmp := t.TempDir()
	pth := filepath.Join(tmp, ".gitkeep")
	f, err := os.Create(pth)
	if err != nil {
		t.Errorf("expected no error writing fixture file %s, got %s", pth, err)
		return
	}
	f.Close()

	tt := []struct {
		pth      string
		expected bool
	}{
		{tmp, true},
		{pth, false},
	}
	for _, tc := range tt {
		got, err := isDir(tc.pth)
		if err != nil {
			t.Errorf("expected no error with %s, got %s", tc.pth, err)
		}
		if got != tc.expected {
			t.Errorf("expected %s to be %t, got %t", tc.pth, tc.expected, got)
		}
	}

}

func TestReadJSONFiles(t *testing.T) {
	tmp := t.TempDir()
	expected := filepath.Join(tmp, "33", "683", "111", "000280.json")
	paths := []string{
		expected,
		filepath.Join(tmp, "33", "683", "111", "XXXXXX.json"),
		filepath.Join(tmp, "33", "683", "XXX", "000280.json"),
		filepath.Join(tmp, ".gitkeep"),
	}
	for _, pth := range paths {
		if err := os.MkdirAll(filepath.Dir(pth), 0755); err != nil {
			t.Errorf("expected no errors creating fixture directory %s, got %s", filepath.Dir(pth), err)
			return
		}
		f, err := os.Create(pth)
		if err != nil {
			t.Errorf("expected no error writing fixture file %s, got %s", pth, err)
			return
		}
		f.Close()
	}

	done := make(chan string)
	errs := make(chan error)
	go allJSONFiles(tmp, done, errs)
	for {
		select {
		case err := <-errs:
			t.Errorf("expected no errors reading %s, got %s", tmp, err)
			return
		case f, ok := <-done:
			if !ok {
				return
			}
			if f != expected {
				t.Errorf("expected %s, got %s", expected, f)
			}
		}
	}
}
