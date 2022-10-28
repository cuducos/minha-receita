package download

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	fileName = "minha.receita.zip"
	fixture  = "512\nminha.receita.zip:0010\n"
)

func TestRecoverFromScratch(t *testing.T) {
	r, err := newRecover(t.TempDir(), 512, false)
	if err != nil {
		t.Errorf("expected to error creating the recover, got %s", err)
	}
	r.addFile(fileName, 4)
	r.chunkDone(fileName, 2)
	for i := 0; i < 4; i++ {
		if i == 2 {
			continue
		}
		if !r.shouldDownload(fileName, i) {
			t.Errorf("expected chunk %d to be pending, but it says it's downloaded", i)
		}
	}
	if r.shouldDownload(fileName, 2) {
		t.Error("expected chunk 0 to be downloaded, but it says it's pending")
	}
	if err := r.save(); err != nil {
		t.Errorf("expected no errors saving the recover file, got %s", err)
	}
	b, err := os.ReadFile(r.path())
	if err != nil {
		t.Errorf("expected no errors reading the recover file, got %s", err)
	}
	if string(b) != fixture {
		t.Errorf("expected recovery file to be:\n%s\ngot:\n%s", fixture, string(b))
	}
}

func TestRecoverFromFile(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, recoverFileName), []byte(fixture), 0755); err != nil {
		t.Errorf("expected to error creating the recover file, got %s", err)
	}

	t.Run("valid chunk size", func(t *testing.T) {
		r, err := newRecover(tmp, 512, false)
		if err != nil {
			t.Errorf("expected to error creating the recover, got %s", err)
		}
		r.addFile(fileName, 4)
		for i := 0; i < 4; i++ {
			if i == 2 {
				continue
			}
			if !r.shouldDownload(fileName, i) {
				t.Errorf("expected chunk %d to be pending, but it says it's downloaded", i)
			}
		}
		if r.shouldDownload(fileName, 2) {
			t.Error("expected chunk 0 to be downloaded, but it says it's pending")
		}
	})

	t.Run("invalid chunk size", func(t *testing.T) {
		_, err := newRecover(tmp, 256, false)
		if err == nil {
			t.Error("expected error creating the recover with invalid chunk size, got nil")
		}
	})

	t.Run("invalid chunk size with restart", func(t *testing.T) {
		r, err := newRecover(tmp, 256, true)
		if err != nil {
			t.Errorf("expected to error creating the recover, got %s", err)
		}
		r.addFile(fileName, 4)
		r.chunkDone(fileName, 2)
		for i := 0; i < 4; i++ {
			if i == 2 {
				continue
			}
			if !r.shouldDownload(fileName, i) {
				t.Errorf("expected chunk %d to be pending, but it says it's downloaded", i)
			}
		}
		if r.shouldDownload(fileName, 2) {
			t.Error("expected chunk 0 to be downloaded, but it says it's pending")
		}
	})
}
