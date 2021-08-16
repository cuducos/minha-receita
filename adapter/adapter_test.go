package adapter

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestTransform(t *testing.T) {
	for _, c := range []string{"", "gz", "xz"} {
		tmp, err := ioutil.TempDir("", fmt.Sprintf("minha-receita-test-%s-", c))
		if err != nil {
			t.Errorf(err.Error())
		}
		defer os.RemoveAll(tmp)

		if err = copyTestDataTo(tmp); err != nil {
			t.Errorf(err.Error())
		}

		if err := Transform(tmp, c, true); err != nil {
			t.Errorf(err.Error())
		}

		for _, k := range []kind{company, facility, partner} {
			n := fmt.Sprintf("test adapter for %s", k)
			if c != "" {
				n = fmt.Sprintf("%s with compression %s", n, c)
			}

			t.Run(n, func(t *testing.T) {
				a := NewAdapter(k, tmp, c)
				p := filepath.Base(a.csvPath())

				expect, err := readCsv(filepath.Join("..", "testdata", p))
				if err != nil {
					t.Errorf(err.Error())
				}

				got, err := readCsv(filepath.Join(tmp, p))
				if err != nil {
					t.Errorf(err.Error())
				}

				if len(got) == 0 {
					t.Error("Expected resulting CSV to have lines")
				}

				if len(expect) != len(got) {
					t.Errorf("Expected resulting CSV to have %d lines, got %d", len(expect), len(got))
				}

				if reflect.DeepEqual(expect[0], got[0]) {
					t.Errorf("Expected resulting CSV header to be %q, got %q", expect[0], got[0])
				}

				if !csvFilesHaveSameElements(got, expect) {
					t.Errorf("CSV rows does not match.")
				}
			})
		}
	}
}

func copyTestDataTo(p string) error {
	ls, err := os.ReadDir(p)
	if err != nil {
		return err
	}

	for _, f := range ls {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".zip") {
			b := make([]byte, 4096)
			d, err := os.Create(filepath.Join(p, f.Name()))
			if err != nil {
				return err
			}

			for {
				h, err := os.Open(filepath.Join(p, f.Name()))
				if err != nil {
					return err
				}

				n, err := h.Read(b)
				if err == io.EOF || n == 0 {
					break
				}
				if err != nil {
					return err
				}
				if _, err := d.Write(b[:n]); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func readCsv(p string) ([][]string, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return csv.NewReader(f).ReadAll()
}

func csvFilesHaveSameElements(a1, a2 [][]string) bool {
	var m1 []string
	var m2 []string
	for _, i := range a1 {
		m1 = append(m1, strings.Join(i, ""))
	}
	for _, i := range a2 {
		m2 = append(m2, strings.Join(i, ""))
	}

	s1 := make(map[string]int)
	s2 := make(map[string]int)
	for _, i := range m1 {
		s1[i] += 1
	}
	for _, i := range m2 {
		s2[i] += 1
	}

	for _, i := range m1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	for _, i := range m2 {
		if s1[i] != s2[i] {
			return false
		}
	}

	return true
}
