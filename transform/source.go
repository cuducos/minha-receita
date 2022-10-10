package transform

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"sync/atomic"
)

type sourceType string

const (
	venues         sourceType = "Estabelecimentos"
	motives        sourceType = "Motivos"
	base           sourceType = "Empresas"
	cities         sourceType = "Municipios"
	cnaes          sourceType = "Cnaes"
	countries      sourceType = "Paises"
	natures        sourceType = "Naturezas"
	partners       sourceType = "Socios"
	qualifications sourceType = "Qualificacoes"
	taxes          sourceType = "Simples"
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
	buf := make([]byte, 32*1024)
	for {
		c, err := a.file.Read(buf)
		t += int64(bytes.Count(buf[:c], []byte{'\n'}))
		if err == io.EOF {
			break
		}
		if err != nil {
			if atomic.CompareAndSwapUint32(&s.shutdown, 0, 1) {
				errs <- err
			}
			return
		}
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
	log.Output(2, fmt.Sprintf("Loading %s files…", string(t)))
	ls, err := PathsForSource(t, d)
	if err != nil {
		return nil, fmt.Errorf("error getting files for %s in %s: %w", string(t), d, err)
	}
	s := source{kind: t, dir: d, files: ls}
	s.createReaders()
	if err = s.countLines(); err != nil {
		return nil, fmt.Errorf("error counting lines for %s in %s: %w", string(t), d, err)
	}
	return &s, nil
}
