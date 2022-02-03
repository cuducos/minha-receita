package transform

import (
	"fmt"
	"io"
	"sync/atomic"
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
	shutdown   uint32
}

func (s *source) createReaders() error {
	s.readers = make([]*archivedCSV, len(s.files))
	for i, p := range s.files {
		r, err := newArchivedCSV(p, separator)
		if err != nil {
			return fmt.Errorf("error reading %s: %w", p, err)
		}
		s.readers[i] = r
	}
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

func (s *source) countLinesFor(a *archivedCSV, count chan<- int64, errs chan<- error) {
	var t int64
	for {
		if atomic.LoadUint32(&s.shutdown) == 1 {
			return
		}
		_, err := a.read()
		if err == io.EOF {
			break
		}
		if err != nil {
			atomic.StoreUint32(&s.shutdown, 1)
			errs <- err
			return
		}
		t++
	}
	if atomic.LoadUint32(&s.shutdown) == 1 {
		return
	}
	count <- t
}

func (s *source) countLines() error {
	count := make(chan int64)
	errs := make(chan error)
	for _, r := range s.readers {
		go s.countLinesFor(r, count, errs)
	}
	defer func() {
		s.resetReaders()
		close(count)
		close(errs)
	}()
	var done int
	for {
		select {
		case err := <-errs:
			return fmt.Errorf("error counting lines: %w", err)
		case n := <-count:
			s.totalLines += n
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
