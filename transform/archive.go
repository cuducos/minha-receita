package transform

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

const separator = ';'

type archivedCSV struct {
	path    string
	reader  *csv.Reader
	toClose []io.Closer
}

func newArchivedCSV(p string, s rune) (*archivedCSV, error) {
	r, err := zip.OpenReader(p)
	if err != nil {
		return nil, fmt.Errorf("error opening archive %s: %w", p, err)
	}

	for _, z := range r.File {
		if z.FileInfo().IsDir() {
			continue
		}

		f, err := z.Open()
		if err != nil {
			return nil, fmt.Errorf("error reading archived file %s in %s: %w", z.Name, p, err)
		}

		c := csv.NewReader(f)
		c.Comma = s
		return &archivedCSV{p, c, []io.Closer{f, r}}, nil
	}

	return nil, fmt.Errorf("could not find a file in the archive %s", p)
}

func (a *archivedCSV) Read() ([]string, error) {
	return a.reader.Read()
}

func (a *archivedCSV) close() error {
	for _, i := range a.toClose {
		if err := i.Close(); err != nil {
			return fmt.Errorf("error closing resource from archive %s: %w", a.path, err)
		}
	}
	return nil
}

func archivedCSVToMap(p string, s rune) (map[int]string, error) {
	m := make(map[int]string)
	z, err := newArchivedCSV(p, s)
	if err != nil {
		return m, fmt.Errorf("error reading archived CSV %s: %w", p, err)
	}
	defer z.close()

	for {
		l, err := z.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return m, fmt.Errorf("error reading CSV from %s: %w", p, err)
		}

		i, err := strconv.Atoi(l[0])
		if err != nil {
			return m, fmt.Errorf("error converting key %s to int in %s: %w", l[0], p, err)
		}

		m[i] = l[1]
	}
	return m, nil
}
