package transform

import (
	"fmt"
	"io"
)

type sourceType string

const (
	venue         sourceType = "ESTABELE"
	motive                   = "MOTICSV"
	main                     = "EMPRECSV"
	city                     = "MUNICCSV"
	cnae                     = "CNAECSV"
	country                  = "PAISCSV"
	nature                   = "NATJUCSV"
	partner                  = "SOCIOCSV"
	qualification            = "QUALSCSV"
	simple                   = "SIMPLES"
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
	counter    chan lineCount
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

func (s *source) countLinesFor(a *archivedCSV) {
	var t int64
	for {
		_, err := a.read()
		if err == io.EOF {
			break
		}
		if err != nil {
			s.counter <- lineCount{0, err}
			return
		}
		t++
	}
	s.counter <- lineCount{t, nil}
}

func (s *source) countLines() error {
	for _, r := range s.readers {
		go s.countLinesFor(r)
	}

	for range s.readers {
		r := <-s.counter
		if r.err != nil {
			return fmt.Errorf("error counting lines: %w", r.err)
		}
		s.totalLines += r.total
	}

	close(s.counter)
	s.resetReaders()
	return nil
}

func newSource(t sourceType, d string) (*source, error) {
	ls, err := PathsForSource(t, d)
	if err != nil {
		return nil, fmt.Errorf("error getting files for %s in %s: %w", string(t), d, err)
	}

	s := source{dir: d, files: ls, counter: make(chan lineCount)}
	s.createReaders()
	s.countLines()
	return &s, nil
}
