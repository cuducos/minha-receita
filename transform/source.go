package transform

import (
	"fmt"
	"io"
)

type sourceType string

const (
	venues         sourceType = "ESTABELE"
	motives                   = "MOTICSV"
	base_cpnj                 = "EMPRECSV"
	cities                    = "MUNICCSV"
	cnaes                     = "CNAECSV"
	countries                 = "PAISCSV"
	natures                   = "NATJUCSV"
	partners                  = "SOCIOCSV"
	qualifications            = "QUALSCSV"
	simple                    = "SIMPLES"
)

type lineCount struct {
	total int64
	err   error
}

type source struct {
	dir        string
	files      []string
	readers    []*archivedCSV
	totalLines int64
}

func (s *source) createReaders() error {
	var as []*archivedCSV
	for _, p := range s.files {
		r, err := newArchivedCSV(p, separator)
		if err != nil {
			return fmt.Errorf("error reading %s: %w", p, err)
		}
		as = append(as, r)
	}
	s.readers = as
	return nil
}

func (s *source) close() error {
	for _, r := range s.readers {
		if err := r.close(); err != nil {
			return fmt.Errorf("error closing %s: %w", r.path, err)
		}
	}
	return nil
}

func (s *source) resetReaders() error {
	if err := s.close(); err != nil {
		return fmt.Errorf("error closing readers: %w", err)
	}
	if err := s.createReaders(); err != nil {
		return fmt.Errorf("error creating readers: %w", err)
	}
	return nil
}

func (s *source) countLinesFor(a *archivedCSV, q chan<- lineCount) {
	var t int64
	for {
		_, err := a.read()
		if err == io.EOF {
			break
		}
		if err != nil {
			q <- lineCount{0, err}
			return
		}
		t++
	}
	q <- lineCount{t, nil}
}

func (s *source) countLines() error {
	q := make(chan lineCount)
	for _, r := range s.readers {
		go s.countLinesFor(r, q)
	}

	for range s.readers {
		r := <-q
		if r.err != nil {
			return fmt.Errorf("error counting lines: %w", r.err)
		}
		s.totalLines += r.total
	}

	close(q)
	s.resetReaders()
	return nil
}

func newSource(t sourceType, d string) (*source, error) {
	ls, err := PathsForSource(t, d)
	if err != nil {
		return nil, fmt.Errorf("error getting files for %s in %s: %w", string(t), d, err)
	}

	s := source{dir: d, files: ls}
	s.createReaders()
	s.countLines()
	return &s, nil
}
