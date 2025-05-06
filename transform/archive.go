package transform

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

var multipleSpaces = regexp.MustCompile(`\s{2,}`)

type archivedCSVs struct {
	path      string
	zipReader io.Closer
	files     []io.ReadCloser
	readers   []*csv.Reader
	reading   int
}

func newArchivedCSV(p string, s rune, h bool) (*archivedCSVs, error) {
	r, err := zip.OpenReader(p)
	if err != nil {
		return nil, fmt.Errorf("error opening archive %s: %w", p, err)
	}
	var a *archivedCSVs
	var fs []io.ReadCloser
	var cs []*csv.Reader
	for _, z := range r.File {
		if z.FileInfo().IsDir() {
			continue
		}
		f, err := z.Open()
		if err != nil {
			return nil, fmt.Errorf("error reading archived file %s in %s: %w", z.Name, p, err)
		}
		c := csv.NewReader(charmap.ISO8859_15.NewDecoder().Reader(f))
		c.Comma = s
		if h {
			if _, err := c.Read(); err != nil {
				return nil, fmt.Errorf("error skipping header of %s in %s: %w", z.Name, p, err)
			}
		}
		fs = append(fs, f)
		cs = append(cs, c)
	}
	a = &archivedCSVs{p, r, fs, cs, 0}
	return a, nil
}

func removeNulChar(r rune) rune {
	if r == '\x00' {
		return -1
	}
	return r
}

func (a *archivedCSVs) read() ([]string, error) {
	if a.reading >= len(a.readers) {
		return []string{}, io.EOF
	}
	ls, err := a.readers[a.reading].Read()
	if err == io.EOF {
		a.reading += 1
		return a.read()
	}
	if err != nil {
		return []string{}, fmt.Errorf("error reading archived csv line from %s: %w", a.path, err)
	}
	for i := range ls {
		ls[i] = multipleSpaces.ReplaceAllString(strings.Map(removeNulChar, ls[i]), " ")
	}
	return ls, nil
}

func (a *archivedCSVs) sendTo(ctx context.Context, ch chan<- []string) error {
	e := make(chan error, 1)
	defer close(e)
	go func() {
		for {
			row, err := a.read()
			if err != nil {
				e <- err
				return
			}
			ch <- row
		}
	}()
	select {
	case <-ctx.Done():
		return nil
	case err := <-e:
		return err
	}
}

func (a *archivedCSVs) close() error {
	for _, f := range a.files {
		if err := f.Close(); err != nil {
			return fmt.Errorf("error closing resource from archive %s: %w", a.path, err)
		}
	}
	if err := a.zipReader.Close(); err != nil {
		return fmt.Errorf("error closing archive %s: %w", a.path, err)
	}
	return nil
}

func (a *archivedCSVs) toLookup() (lookup, error) {
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
