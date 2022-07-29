package transform

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/charmap"
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

	var a *archivedCSV
	t := strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))
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
		a = &archivedCSV{p, c, []io.Closer{f, r}}
		break
	}

	if a == nil {
		return nil, fmt.Errorf("could not find file %s in the archive %s", t, p)
	}

	return a, nil
}

func removeNulChar(r rune) rune {
	if r == '\x00' {
		return -1
	}
	return r
}

func (a *archivedCSV) read() ([]string, error) {
	ls, err := a.reader.Read()
	if err == io.EOF {
		return []string{}, err
	}
	if err != nil {
		return []string{}, fmt.Errorf("error reading archived csv line from %s: %w", a.path, err)
	}
	for i, l := range ls {
		ls[i], err = charmap.ISO8859_1.NewDecoder().String(l)
		if err != nil {
			return []string{}, fmt.Errorf("encoding error in text %s from %s: %w", l, a.path, err)
		}
		ls[i] = strings.Map(removeNulChar, ls[i])
	}
	return ls, nil
}

func (a *archivedCSV) close() error {
	for _, i := range a.toClose {
		if err := i.Close(); err != nil {
			return fmt.Errorf("error closing resource from archive %s: %w", a.path, err)
		}
	}
	return nil
}

func (a *archivedCSV) toLookup() (lookup, error) {
	m := make(map[int]string)
	for {
		l, err := a.read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return m, fmt.Errorf("error reading CSV from %s: %w", a.path, err)
		}
		i, err := strconv.Atoi(l[0])
		if err != nil {
			return m, fmt.Errorf("error converting key %s to int in %s: %w", l[0], a.path, err)
		}
		m[i] = l[1]
	}
	return m, nil
}
