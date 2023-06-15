package check

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const numTestFiles = 2

var (
	contents = []string{"This is the contents of file 0", "This is the contents of file 1"}
	hashes   = []string{"2e68b218b62624b52ee519a42e09f87d", "aa45a469cfb5b5b2ca3094568f0f2039"}
)

func checksumTestFiles(t *testing.T) string {
	tmp := t.TempDir()
	for n := 0; n < numTestFiles; n++ {
		f := filepath.Join(tmp, fmt.Sprintf("file%d", n))
		if err := os.WriteFile(f, []byte(contents[n]), 0755); err != nil {
			t.Fatalf("failed to write file %s: %s", f, err)
		}
		if err := os.WriteFile(f+".md5", []byte(hashes[n]), 0755); err != nil {
			t.Fatalf("failed to write checksum file %s: %s", f+".md5", err)
		}
	}
	return tmp
}

func TestCreate(t *testing.T) {
	t.Run("source directory does not exist", func(t *testing.T) {
		err := CreateChecksum(filepath.Join("..", "directory-does-not-exist"))
		if err == nil {
			t.Error("expected error creating checksum files, got nil")
		}
	})

	t.Run("create checksum files", func(t *testing.T) {
		src := checksumTestFiles(t)
		err := CreateChecksum(src)
		if err != nil {
			t.Errorf("expected no error creating checksum files, got %s", err)
		}
		ls, err := filepath.Glob(filepath.Join(src, "*.md5"))
		if err != nil {
			t.Errorf("expected no error reading dir %s, got %s", src, err)
		}
		var got int
		for range ls {
			got++
		}
		if got != numTestFiles {
			t.Errorf("expected %d files in the sample directory, got %d: %v", numTestFiles, got, ls)
		}
	})
}

func TestCheck(t *testing.T) {
	t.Run("source directory does not exist", func(t *testing.T) {
		tmp := t.TempDir()
		err := CheckChecksum(tmp, tmp)
		if err == nil {
			t.Error("expected error checking checksums, got nil")
		}
	})

	t.Run("missing source files", func(t *testing.T) {
		src := checksumTestFiles(t)
		out := checksumTestFiles(t)
		for n := 0; n < numTestFiles; n++ {
			if err := os.Remove(filepath.Join(src, fmt.Sprintf("file%d.md5", n))); err != nil {
				t.Fatalf("expected no error removing file from testdata, got %s", err)
			}
		}
		err := CheckChecksum(src, out)
		if err == nil {
			t.Error("expected error checking checksums, got nil")
		}
	})

	t.Run("match", func(t *testing.T) {
		src := checksumTestFiles(t)
		out := checksumTestFiles(t)
		if err := CheckChecksum(src, out); err != nil {
			t.Errorf("expected no error checking checksums, got %s", err)
		}
	})

	t.Run("no match", func(t *testing.T) {
		src := checksumTestFiles(t)
		out := checksumTestFiles(t)
		h := md5.New()
		h.Write([]byte("different data"))
		if err := os.WriteFile(filepath.Join(src, "file0.md5"), h.Sum(nil), 0755); err != nil {
			t.Fatalf("expected no error creating %s checksum file in directory, got %s", src, err)
		}
		if err := CheckChecksum(src, out); err == nil {
			t.Error("expected error checking checksums, got nil")
		}
	})
}

func TestChecksumFor(t *testing.T) {
	src := checksumTestFiles(t)
	f1 := filepath.Join(src, "file0")
	h1, err := checksumFor(f1)
	if err != nil {
		t.Errorf("expected no error getting %s checksum, got %s", f1, err)
	}
	if h1 != hashes[0] {
		t.Errorf("expected checksum %s for file %s, got %s", hashes[0], f1, h1)
	}

	f2 := filepath.Join(src, "file1")
	h2, err := checksumFor(f2)
	if err != nil {
		t.Errorf("expected no error getting %s checksum, got %s", f2, err)
	}
	if h2 != hashes[1] {
		t.Errorf("expected checksum %s for file %s, got %s", hashes[1], f2, h2)
	}
}
