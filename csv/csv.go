// Package csv handles the creation of a CSV file for PostgreSQL copy command.
package csv

import (
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cuducos/minha-receita/transform"
	"github.com/schollz/progressbar/v3"
)

const (
	// Path is the name of CSV file ready for PostgresSQL copy command.
	Path = "cnpj.csv.gz"

	// IDFieldName is the name of the primary key column in PostgreSQL, i.e. the CNPJ.
	IDFieldName = "id"

	// JSONFieldName is the name of the column in PostgreSQL with the JSON content.
	JSONFieldName = "json"

	// MaxReadDirEntries is the maximum number of files or directories
	// returned at each attempt to read a directory.
	MaxReadDirEntries = 32
)

type task struct {
	dir    string
	rows   chan []string
	errors chan error
	wg     *sync.WaitGroup
}

func (t *task) readDir(dir string) {
	if t.dir != dir { // we keep track of recursive calls in a WaitGroup
		defer t.wg.Done()
	} else { // wait for the recursive calls to finish and close the channel
		defer func(t *task) {
			t.wg.Wait()
			close(t.rows)
			close(t.errors)
		}(t)
	}

	d, err := os.Open(dir)
	if err != nil {
		t.errors <- fmt.Errorf("error opening directory %s: %w", t.dir, err)
		return
	}
	defer d.Close()

	for {
		ls, err := d.Readdirnames(MaxReadDirEntries)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.errors <- fmt.Errorf("error reading diretcory %s: %w", dir, err)
			return
		}
		for _, p := range ls {
			p := filepath.Join(d.Name(), p)
			f, err := os.Open(p)
			if err != nil {
				t.errors <- fmt.Errorf("error opening path %s: %w", p, err)
				continue
			}
			s, err := f.Stat()
			if err != nil {
				t.errors <- fmt.Errorf("error getting stat for %s: %w", p, err)
				continue
			}
			if s.IsDir() {
				t.wg.Add(1)
				go t.readDir(p)
				continue
			}
			if !strings.HasSuffix(p, ".json") {
				continue
			}
			n, err := transform.CNPJForPath(p)
			if err != nil {
				log.Output(2, fmt.Sprintf("Invalid JSON file path for a CNPJ %s", p))
				continue
			}
			j, err := ioutil.ReadFile(p)
			if err != nil {
				t.errors <- fmt.Errorf("error reading %s: %w", p, err)
				continue
			}
			t.rows <- []string{n, strings.TrimSpace(string(j))}
		}
	}
	return
}

func newTask(dir string) task {
	var wg sync.WaitGroup
	return task{
		dir:    dir,
		rows:   make(chan []string),
		errors: make(chan error),
		wg:     &wg,
	}
}

// CreateCSV creates a GZipped CSV file for PostgreSQ copy command.
func CreateCSV(dir string) error {
	p := filepath.Join(dir, Path)
	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("error while creating the CSV file %s: %w", p, err)
	}
	defer f.Close()

	z := gzip.NewWriter(f)
	defer z.Close()

	w := csv.NewWriter(z)
	w.Write([]string{IDFieldName, JSONFieldName})
	defer w.Flush()

	t := newTask(dir)
	go t.readDir(dir)

	bar := progressbar.Default(-1, fmt.Sprintf("Writing CNPJ data to %s", p))
	for {
		select {
		case err := <-t.errors:
			return fmt.Errorf("error running create csv: %w", err)
		case r, ok := <-t.rows:
			w.Write(r)
			bar.Add(1)
			if !ok {
				return nil
			}
		}
	}
}
