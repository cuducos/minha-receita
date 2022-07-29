package check

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

const badZipFile = "BAD_FILE.zip"

var testdata = filepath.Join("..", "testdata")

func TestCheckZipFiles(t *testing.T) {
	t.Run("failure", func(t *testing.T) {
		tmp := t.TempDir()
		if err := copyZipFiles(tmp); err != nil {
			t.Fatal("could not copy test files")
		}
		if err := createBadZipFile(tmp); err != nil {
			t.Fatal("could not create test files")
		}
		got, err := checkZipFiles(tmp)
		if err != nil {
			t.Errorf("expected no errors, got %s", err)
		}
		expected := []string{filepath.Join(tmp, badZipFile)}
		if len(got) != len(expected) {
			t.Errorf("expected %d files, got %d", len(expected), len(got))
		}
		var ok bool
		for pth := range got {
			ok = false
			for _, e := range expected {
				if pth == e {
					ok = true
					continue
				}
			}

			if !ok {
				t.Errorf("unexpected %s in the results: %v", pth, expected)
			}
		}
	})

	t.Run("success", func(t *testing.T) {
		tmp := t.TempDir()
		if err := copyZipFiles(tmp); err != nil {
			t.Fatal("could not copy test files")
		}
		got, err := checkZipFiles(t.TempDir())
		if err == nil {
			t.Error("expected error, got nil")
		}
		if len(got) != 0 {
			t.Errorf("expected no files, got %d", len(got))
		}
	})
}

func TestCheckZipFile(t *testing.T) {
	tmp := t.TempDir()
	badZipPath := filepath.Join(tmp, badZipFile)
	if err := copyZipFiles(tmp); err != nil {
		t.Fatal("could not copy test files")
	}
	if err := createBadZipFile(tmp); err != nil {
		t.Fatal("could not create test files")
	}
	tt := []struct {
		pth      string
		expected error
	}{
		{filepath.Join(testdata, "Simples.zip"), nil},
		{badZipPath, fmt.Errorf("error opening %s: zip: not a valid zip file", badZipPath)},
	}
	for _, tc := range tt {
		err := checkZipFile(tc.pth)
		if tc.expected == nil {
			if err != tc.expected {
				t.Errorf("expected nil, got %s", err)
			}
		}
		if tc.expected != nil {
			if err.Error() != tc.expected.Error() {
				t.Errorf("expected %s, got %s", tc.expected, err)
			}
		}
	}
}

func copyZipFiles(dir string) error {
	ls, err := filepath.Glob(filepath.Join(testdata, "*.zip"))
	if err != nil {
		return fmt.Errorf("error find zip files for the test: %w", err)
	}
	cp := func(f, dir string) error {
		src, err := os.Open(f)
		if err != nil {
			return fmt.Errorf("error opening %s: %w", f, err)
		}
		defer src.Close()
		pth := filepath.Join(dir, filepath.Base(f))
		dst, err := os.Create(pth)
		if err != nil {
			return fmt.Errorf("error creating %s: %w", pth, err)
		}
		defer dst.Close()
		_, err = io.Copy(dst, src)
		if err != nil {
			return fmt.Errorf("error copying %s to %s: %w", f, pth, err)
		}
		return nil
	}
	for _, f := range ls {
		if err := cp(f, dir); err != nil {
			return err
		}
	}
	return nil
}

func createBadZipFile(dir string) error {
	pth := filepath.Join(dir, badZipFile)
	dst, err := os.Create(pth)
	if err != nil {
		return fmt.Errorf("error creating %s: %w", pth, err)
	}
	defer dst.Close()
	return nil
}
