package transform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTaskRun(t *testing.T) {
	outDir := t.TempDir()
	p, err := newTask(filepath.Join("..", "testdata"), outDir)
	if err != nil {
		t.Errorf("expected no error creating task, got %s", err)
	}
	if err = p.run(2); err != nil {
		t.Errorf("expected no error running task, got %s", err)
	}

	var j []string
	err = filepath.WalkDir(outDir, func(p string, d os.DirEntry, err error) error {
		if strings.HasSuffix(p, ".json") {
			j = append(j, p)
		}
		return nil
	})
	if err != nil {
		t.Errorf("expected no error walking %s, got %s", outDir, err)
	}
	if len(j) != 1 {
		t.Errorf("expected 1 JSON files, got %d", len(j))
	}
}
