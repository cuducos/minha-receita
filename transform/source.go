package transform

import (
	"context"
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
	if len(ls) == 0 {
		return []string{}, fmt.Errorf("could not find any file matching %s in %s", string(t), dir)
	}
	return ls, nil
}

type sourceType string

const (
	venues           sourceType = "Estabelecimentos"
	motives          sourceType = "Motivos"
	base             sourceType = "Empresas"
	cities           sourceType = "Municipios"
	cnaes            sourceType = "Cnaes"
	countries        sourceType = "Paises"
	natures          sourceType = "Naturezas"
	partners         sourceType = "Socios"
	qualifications   sourceType = "Qualificacoes"
	simpleTaxes      sourceType = "Simples"
	realProfit       sourceType = "Lucro Real"
	presumedProfit   sourceType = "Lucro Presumido"
	arbitratedProfit sourceType = "Lucro Arbitrado"
	noTaxes          sourceType = "Imunes e Isentas"
)

type source struct {
	kind     sourceType
	dir      string
	files    []string
	readers  []*archivedCSVs
	total    int
	shutdown int32
}

func (s *source) createReaders() error {
	s.readers = make([]*archivedCSVs, len(s.files))
	for i, p := range s.files {
		var h bool
		var sep rune
		switch s.kind {
		case realProfit:
			sep = ','
			h = true
		case presumedProfit:
			sep = ','
			h = true
		case arbitratedProfit:
			sep = ','
			h = true
		case noTaxes:
			sep = ','
			h = true
		default:
			sep = ';'
			h = false
		}
		r, err := newArchivedCSV(p, sep, h)
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

func (s *source) countCSVRows(a *archivedCSVs, count chan<- int, errs chan<- error) {
	var t int
	for {
		_, err := a.read()
		if err == io.EOF {
			break
		}
		if err != nil {
			if atomic.CompareAndSwapInt32(&s.shutdown, 0, 1) {
				errs <- err
			}
		}
		t++
	}
	count <- t
}

func (s *source) countLines() error {
	count := make(chan int)
	errs := make(chan error)
	for _, r := range s.readers {
		go s.countCSVRows(r, count, errs)
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
			s.total += n
			done++
			if done == len(s.readers) {
				return nil
			}
		}
	}
}

func newSource(ctx context.Context, t sourceType, d string) (*source, error) {
	log.Output(1, fmt.Sprintf("Loading %s files…", string(t)))
	var s source
	ch := make(chan error, 1)
	done := atomic.Bool{}
	defer close(ch)
	go func() {
		ls, err := pathsForSource(t, d)
		if err != nil {
			if !done.Load() {
				ch <- fmt.Errorf("error getting files for %s in %s: %w", string(t), d, err)
			}
			return
		}
		s = source{kind: t, dir: d, files: ls}
		s.createReaders()
		if err = s.countLines(); err != nil {
			if !done.Load() {
				ch <- fmt.Errorf("error counting lines for %s in %s: %w", string(t), d, err)
			}
		}
		if !done.Load() {
			ch <- nil
		}
	}()
	select {
	case <-ctx.Done():
		done.Store(true)
		return nil, ctx.Err()
	case err := <-ch:
		return &s, err
	}
}

func newSources(dir string, kinds []sourceType) ([]*source, error) {
	srcs := []*source{}
	ok := make(chan *source)
	errs := make(chan error, 1)
	done := atomic.Bool{}
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		close(ok)
		close(errs)
	}()
	for _, s := range kinds {
		go func(s sourceType) {
			src, err := newSource(ctx, s, dir)
			if err != nil {
				if !done.Load() {
					errs <- fmt.Errorf("could not load source %s: %w", string(s), err)
				}
				return
			}
			if !done.Load() {
				ok <- src
			}
		}(s)
	}
	for {
		select {
		case err := <-errs:
			done.Store(true)
			return nil, fmt.Errorf("error loading sources: %w", err)
		case src := <-ok:
			srcs = append(srcs, src)
			log.Output(1, fmt.Sprintf("[%d/%d] %s loaded!", len(srcs), len(kinds), src.kind))
			if len(srcs) == len(kinds) {
				return srcs, nil
			}
		}
	}
}
