package transform

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTask(t *testing.T) {
	d := t.TempDir()
	for _, s := range []sourceType{venue, motive} {
		ls, err := PathsForSource(s, filepath.Join("..", "testdata"))
		if err != nil {
			t.Errorf("expected no error finding paths for %s, got %s", string(s), err)
		}
		for _, f := range ls {
			b, err := ioutil.ReadFile(f)
			if err != nil {
				t.Errorf("expected no error reading %s, got %s", f, err)
			}

			o := filepath.Join(d, filepath.Base(f))
			err = ioutil.WriteFile(o, b, 0644)
			if err != nil {
				t.Errorf("expected no error writing %s, got %s", o, err)
			}
		}
	}

	p, err := newTask(d, venue)
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
	if len(j) != 2 {
		t.Errorf("expected 2 JSON files, got %d", len(j))
	}
}
