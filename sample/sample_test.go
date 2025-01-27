package sample

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cuducos/minha-receita/download"
)

var testdata = filepath.Join("..", "testdata")

func testdataWithoutUpdatedAt(t *testing.T) string {
	tmp := t.TempDir()
	ls, err := filepath.Glob(filepath.Join(testdata, "*"))
	if err != nil {
		t.Fatalf("could not read %s", testdata)
	}
	for _, f := range ls {
		if filepath.Base(f) == download.FederalRevenueUpdatedAt {
			continue
		}
		s, err := os.Stat(f)
		if err != nil {
			t.Fatalf("could not read %s", f)
		}
		if s.IsDir() || !s.Mode().IsRegular() {
			continue
		}
		func() {
			r, err := os.Open(f)
			if err != nil {
				t.Fatalf("could not open %s", f)
			}
			defer r.Close()
			d := filepath.Join(tmp, filepath.Base(f))
			w, err := os.Create(d)
			if err != nil {
				t.Fatalf("could not create %s", d)
			}
			defer w.Close()
			_, err = io.Copy(w, r)
			if err != nil {
				t.Fatalf("could not copy %s to %s", f, d)
			}
		}()
	}
	return tmp
}

func TestSample(t *testing.T) {
	testCases := []struct {
		desc      string
		src       string
		updatedAt string
		expected  int
		err       bool
	}{
		{
			"Copy existing updated_at.txt file",
			testdata,
			"",
			13,
			false,
		},
		{
			"Ignore updated_at.txt file",
			testdataWithoutUpdatedAt(t),
			"",
			12,
			false,
		},
		{
			"updated_at.txt file with invalid date format",
			testdataWithoutUpdatedAt(t),
			"17-10-2022",
			12,
			true,
		},
		{
			"updated_at.txt from user input",
			testdataWithoutUpdatedAt(t),
			time.Now().Format("2006-01-02"),
			13,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			out := t.TempDir()
			err := Sample(tc.src, out, 2, tc.updatedAt)
			if !tc.err && err != nil {
				t.Fatalf("expected no error running sample, got %s", err)
			}
			ls, err := os.ReadDir(out)
			if err != nil {
				t.Errorf("expected no error reading dir %s, got %s", out, err)
			}
			var got int
			for _, f := range ls {
				if !f.IsDir() {
					got++
				}
			}
			if got != tc.expected {
				t.Errorf("expected %d files in the sample directory, got %d", tc.expected, got)
			}
		})
	}
}
