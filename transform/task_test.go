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
	testdata := []struct {
		desc   string
		source sourceType
		nFiles int
	}{
		{"ESTABELE", venue, 1},
	}
	for _, test := range testdata {
		t.Run(test.desc, func(t *testing.T) {
			d := t.TempDir()
			ls, err := PathsForSource(test.source, filepath.Join("..", "testdata"))
			if err != nil {
				t.Errorf("expected no error finding paths for %s, got %s", string(test.source), err)
			}
			for _, f := range ls {
				s, err := os.Open(f)
				if err != nil {
					t.Errorf("expected no error opening file %s, got %s", f, err)
				}
				defer s.Close()
				o := filepath.Join(d, filepath.Base(f))
				d, err := os.Create(o)
				if err != nil {
					t.Errorf("expected no error creating file %s, got %s", o, err)
				}
				defer d.Close()
				if _, err := io.Copy(d, s); err != nil {
					t.Errorf("expected no error copying file %s to %s, got %s", f, o, err)
				}
				fmt.Println(test.desc, o)
			}
			p, err := newTask(d, test.source)
			if err != nil {
				t.Errorf("expected no error creating task, got %s", err)
			}
			fmt.Println("task created", p, "total lines:", p.source.totalLines)
			if err = p.run(test.nFiles); err != nil {
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
			if len(j) != test.nFiles {
				t.Errorf("expected %d JSON files (EMPRECSV), got %d", test.nFiles, len(j))
			}
		})
	}
}
