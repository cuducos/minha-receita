package transform

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTaskRun(t *testing.T) {
	d := t.TempDir()
	for _, src := range []sourceType{venues, motives, cities, countries, cnaes, qualifications, natures} {
		ls, err := PathsForSource(src, filepath.Join("..", "testdata"))
		if err != nil {
			t.Errorf("expected no error finding paths for %s, got %s", string(src), err)
		}
		for _, f := range ls {
			copyFile(f, d)
		}
	}
	p, err := newTask(d, d)
	if err != nil {
		t.Errorf("expected no error creating task, got %s", err)
	}
	if err = p.run(2); err != nil {
		t.Errorf("expected no error running task, got %s", err)
	}

	var j []string
	err = filepath.WalkDir(d, func(p string, d os.DirEntry, err error) error {
		if strings.HasSuffix(p, ".json") {
			j = append(j, p)
		}
		return nil
	})
	if err != nil {
		t.Errorf("expected no error walking %s, got %s", d, err)
	}
	if len(j) != 1 {
		t.Errorf("expected 1 JSON files, got %d", len(j))
	}
}

func copyFile(src string, targetDir string) error {
	s, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("expected no error opening file %s, got %s", src, err)
	}
	defer s.Close()
	o := filepath.Join(targetDir, filepath.Base(src))
	t, err := os.Create(o)
	if err != nil {
		return fmt.Errorf("expected no error creating file %s, got %s", o, err)
	}
	defer t.Close()
	if _, err := io.Copy(t, s); err != nil {
		return fmt.Errorf("expected no error copying file %s to %s, got %s", src, o, err)
	}
	return nil
}
