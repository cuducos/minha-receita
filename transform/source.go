package transform

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
)

func pathsForSource(t sourceType, dir string) ([]string, error) {
	r, err := os.ReadDir(dir)
	if err != nil {
		return []string{}, err
	}
	var ls []string
	for _, f := range r {
		if f.IsDir() || filepath.Ext(f.Name()) == ".md5" {
			continue
		}
		if strings.Contains(strings.ToLower(f.Name()), strings.ToLower(string(t))) {
			ls = append(ls, filepath.Join(dir, f.Name()))
		}
	}
	return ls, nil
}

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
	totalLines int
	shutdown   int32
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

func (s *source) countLinesFor(a *archivedCSV, count chan<- int, errs chan<- error) {
	var t int
	buf := make([]byte, 32*1024)
	for {
		c, err := a.file.Read(buf)
		t += bytes.Count(buf[:c], []byte{'\n'})
		if err == io.EOF {
			break
		}
		if err != nil {
			if atomic.CompareAndSwapInt32(&s.shutdown, 0, 1) {
				errs <- err
			}
			return
		}
	}
	if atomic.LoadInt32(&s.shutdown) == 1 {
		return
	}
	count <- t
}

func (s *source) countLines() error {
	count := make(chan int)
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
	log.Output(1, fmt.Sprintf("Loading %s files…", string(t)))
	ls, err := pathsForSource(t, d)
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

func newSources(dir string, kinds []sourceType) ([]*source, error) {
	srcs := []*source{}
	done := make(chan *source)
	errs := make(chan error)
	defer func() {
		close(done)
		close(errs)
	}()
	for _, s := range kinds {
		go func(s sourceType) {
			src, err := newSource(s, dir)
			if err != nil {
				errs <- fmt.Errorf("could not load source %s: %w", string(s), err)
				return
			}
			done <- src
		}(s)
	}
	for {
		select {
		case err := <-errs:
			return nil, fmt.Errorf("error loading sources: %w", err)
		case src := <-done:
			srcs = append(srcs, src)
			if len(srcs) == len(kinds) {
				return srcs, nil
			}
		}
	}
}
