package transform

import (
	"fmt"
	"io"

	"github.com/schollz/progressbar/v3"
)

type venuesTask struct {
	source  *source
	lookups *lookups
	outDir  string
	queue   chan []string
	paths   chan string
	errors  chan error
	bar     *progressbar.ProgressBar
}

func (t *venuesTask) produceRows() {
	for _, r := range t.source.readers {
		go func(t *venuesTask, a *archivedCSV) {
			for {
				r, err := a.read()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.errors <- err
					break // do not proceed in case of errors.
				}
				t.queue <- r
			}
		}(t, r)
	}
}

func (t *venuesTask) consumeRows() {
	for r := range t.queue {
		c, err := newCompany(r, t.lookups)
		if err != nil {
			t.errors <- fmt.Errorf("error parsing company from %q: %w", r, err)
			break
		}
		p, err := c.toJSON(t.outDir)
		if err != nil {
			t.errors <- fmt.Errorf("error getting the JSON bytes for %v: %w", c, err)
			break
		}
		t.paths <- p
	}
}

func (t *venuesTask) run(m int) error {
	defer t.source.close()
	t.produceRows()
	for i := 0; i < m; i++ {
		go t.consumeRows()
	}
	defer func() {
		close(t.queue)
		close(t.paths)
		close(t.errors)
	}()
	for {
		select {
		case err := <-t.errors:
			return err
		case <-t.paths:
			t.bar.Add(1)
			if t.bar.IsFinished() {
				return nil
			}
		}
	}
}

func createJSONFiles(srcDir, outDir string) (*venuesTask, error) {
	v, err := newSource(venues, srcDir)
	if err != nil {
		return nil, fmt.Errorf("error creating a source for venues from %s: %w", srcDir, err)
	}
	l, err := newLookups(srcDir)
	if err != nil {
		return nil, fmt.Errorf("error creating look up tables from %s: %w", srcDir, err)
	}
	t := venuesTask{
		source:  v,
		outDir:  outDir,
		lookups: &l,
		bar:     progressbar.Default(v.totalLines),
		queue:   make(chan []string),
		paths:   make(chan string),
		errors:  make(chan error),
	}
	t.bar.Describe("Creating a JSON file for each CNPJ")
	return &t, nil
}
