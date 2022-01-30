package transform

import (
	"fmt"
	"io"
)

type sourceType string

const (
	venues         sourceType = "ESTABELE"
	motives        sourceType = "MOTICSV"
	base           sourceType = "EMPRECSV"
	cities         sourceType = "MUNICCSV"
	cnaes          sourceType = "CNAECSV"
	countries      sourceType = "PAISCSV"
	natures        sourceType = "NATJUCSV"
	partners       sourceType = "SOCIOCSV"
	qualifications sourceType = "QUALSCSV"
	taxes          sourceType = "SIMPLES"
)

type source struct {
	kind       sourceType
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
		return fmt.Errorf("error re-creating readers: %w", err)
	}
	return nil
}

func (s *source) countLinesFor(a *archivedCSV, count chan<- int64, errs chan<- error, done chan<- struct{}) {
	defer func() { done <- struct{}{} }()
	var t int64
	for {
		_, err := a.read()
		if err == io.EOF {
			break
		}
		if err != nil {
			errs <- err
			return
		}
		t++
	}
	count <- t
}

func (s *source) countLines() error {
	var done int
	count := make(chan int64)
	errs := make(chan error)
	read := make(chan struct{})
	for _, r := range s.readers {
		go s.countLinesFor(r, count, errs, read)
	}
	defer func() {
		s.resetReaders()
		close(read)
		close(count)
		close(errs)
	}()
	for {
		select {
		case err := <-errs:
			return fmt.Errorf("error counting lines: %w", err)
		case n := <-count:
			s.totalLines += n
		case <-read:
			done++
			if done == len(s.readers) {
				return nil
			}
		}
	}
}

func newSource(t sourceType, d string) (*source, error) {
	ls, err := PathsForSource(t, d)
	if err != nil {
		return nil, fmt.Errorf("error getting files for %s in %s: %w", string(t), d, err)
	}
	s := source{kind: t, dir: d, files: ls}
	s.createReaders()
	s.countLines()
	return &s, nil
}
