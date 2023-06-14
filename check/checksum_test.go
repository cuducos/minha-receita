package check

import (
	"crypto/md5"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func checksumTestdata(t *testing.T) string {
	out := t.TempDir()
	src := filepath.Join("..", "testdata")
	ls, err := os.ReadDir(src)
	if err != nil {
		t.Fatalf("expected no error reading testdata directory, got %s", err)
	}
	for _, f := range ls {
		if filepath.Ext(f.Name()) == ".md5" {
			continue
		}
		func() {
			r, err := os.Open(filepath.Join(src, f.Name()))
			if err != nil {
				t.Fatalf("expected no error opening %s in testdata directory, got %s", f.Name(), err)
			}
			defer r.Close()

			w, err := os.Create(filepath.Join(out, f.Name()))
			if err != nil {
				t.Fatalf("expected no error creating %s in tmp testdata directory, got %s", f.Name(), err)
			}
			defer w.Close()

			if _, err := io.Copy(w, r); err != nil {
				t.Fatalf("expected no error writing %s, got %s", f.Name(), err)
			}
		}()
	}
	return out
}

func TestCreate(t *testing.T) {
	testCases := []struct {
		desc     string
		src      string
		expected int
		err      bool
	}{
		{
			"Src directory does not exists",
			filepath.Join("..", "no-dir"),
			0,
			true,
		},
		{
			"Create checksum files",
			checksumTestdata(t),
			16,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			err := CreateChecksum(tc.src)

			if !tc.err && err != nil {
				t.Errorf("expected no error creating checksum files, got %s", err)
			}

			ls, err := filepath.Glob(filepath.Join(tc.src, "*.md5"))
			if err != nil {
				t.Errorf("expected no error reading dir %s, got %s", tc.src, err)
			}

			var got int
			for range ls {
				got++
			}

			if got != tc.expected {
				t.Errorf("expected %d files in the sample directory, got %d", tc.expected, got)
			}
		})
	}
}

func createChecksumFilesInTestDirectory(t *testing.T) string {
	out := checksumTestdata(t)

	if err := CreateChecksum(out); err != nil {
		t.Errorf("expected no error creating testdata checksums, got %s", err)
	}
	return out
}

func testdataWithMissingFiles(t *testing.T) string {
	src := checksumTestdata(t)

	if err := os.Remove(filepath.Join(src, "Empresas0.zip")); err != nil {
		t.Errorf("expected no error removing file from testdata, got %s", err)
	}

	return src
}

func testdataWithInvalidChecksums(t *testing.T) string {
	src := createChecksumFilesInTestDirectory(t)

	fh := md5.New()
	fh.Write([]byte("different data"))

	if err := os.WriteFile(filepath.Join(src, "Empresas0.zip.md5"), fh.Sum(nil), 0755); err != nil {
		t.Errorf("expected no error creating %s checksum file in directory, got %s", src, err)
	}

	return src
}

func TestCheck(t *testing.T) {
	testCases := []struct {
		desc   string
		src    string
		target string
		err    bool
	}{
		{
			"Src directory has no checksum files",
			t.TempDir(),
			t.TempDir(),
			true,
		},
		{
			"Missing file in target directory",
			checksumTestdata(t),
			testdataWithMissingFiles(t),
			true,
		},
		{
			"Checksums match",
			createChecksumFilesInTestDirectory(t),
			createChecksumFilesInTestDirectory(t),
			false,
		},
		{
			"Checksums does not match",
			checksumTestdata(t),
			testdataWithInvalidChecksums(t),
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			err := CheckChecksum(tc.src, tc.target)

			if !tc.err && err != nil {
				t.Errorf("expected no error checking checksums, got %s", err)
			}
		})
	}
}

func TestChecksumFor(t *testing.T) {
	_, err := checksumFor(filepath.Join("..", "no-dir", "Estabelecimentos0.zip"))

	if err == nil {
		t.Error("expected error getting file checksum, got nil")
	}

	src := filepath.Join("..", "testdata")

	sb, err := checksumFor(filepath.Join(src, "Estabelecimentos0.zip"))
	if err != nil {
		t.Errorf("expected no error getting file checksum, got %s", err)
	}

	tb, err := checksumFor(filepath.Join(src, "response.json"))
	if err != nil {
		t.Errorf("expected no error getting file checksum, got %s", err)
	}

	if sb == tb {
		t.Errorf("expected different checksums for files, but got equal ones")
	}

	r, err := os.Open(filepath.Join(src, "Estabelecimentos0.zip.md5"))
	if err != nil {
		t.Errorf("expected no error opening template checksum file, got %s", err)
	}

	ssb, err := io.ReadAll(r)
	if err != nil {
		t.Errorf("expected no error reading template checksum file, got %s", err)
	}

	fssb := strings.Trim(string(ssb), "\n")

	if sb != fssb {
		t.Errorf("expected equal checksums for the same file, but got different ones")
	}
}
