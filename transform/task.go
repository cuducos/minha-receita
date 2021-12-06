package transform

import (
	"fmt"
	"io"

	"github.com/schollz/progressbar/v3"
)

type task struct {
	source  *source
	queue   chan []string
	paths   chan string
	errors  chan error
	bar     *progressbar.ProgressBar
	motives map[int]string
}

func (t *task) loadMotives(d string, s rune) error {
	ls, err := PathsForSource(motive, d)
	if err != nil {
		return fmt.Errorf("cannot find files for motives: %w", err)
	}

	if len(ls) < 1 {
		return fmt.Errorf("cannot find files for motives: %w", err)
	}

	z, err := newArchivedCSV(ls[0], s)
	if err != nil {
		return fmt.Errorf("error loading archived CSV to build a map: %w", err)
	}
	defer z.close()
	t.motives, err = z.toMap()
	if err != nil {
		return fmt.Errorf("error creating motives lookup map: %w", err)
	}
	return nil
}

func (t *task) produceRows() {
	for _, r := range t.source.readers {
		go func(t *task, a *archivedCSV) {
			for {
				r, err := a.read()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.errors <- err
					continue
				}
				t.queue <- r
			}
		}(t, r)
	}
}

func (t *task) consumeRows() {
	for r := range t.queue {
		c, err := newCompany(r, t.motives)
		if err != nil {
			t.errors <- fmt.Errorf("error parsing company from %q: %w", r, err)
			continue
		}
		p, err := c.toJSON(t.source.dir)
		if err != nil {
			t.errors <- fmt.Errorf("error getting the JSON bytes for %v: %w", c, err)
			continue
		}
		t.paths <- p
	}
}

func (t *task) run(m int) error {
	defer t.source.close()
	t.produceRows()
	for i := 0; i < m; i++ {
		go t.consumeRows()
	}
	var c int64
	for {
		select {
		case err := <-t.errors:
			close(t.queue)
			close(t.paths)
			close(t.errors)
			return err
		case <-t.paths:
			c++
			if c == t.source.totalLines {
				close(t.queue)
				close(t.paths)
				close(t.errors)
				return nil
			}
		}
	}
}

func newTask(d string, t sourceType) (*task, error) {
	s, err := newSource(t, d)
	if err != nil {
		return nil, fmt.Errorf("error creating a source for %s from %s: %w", string(t), d, err)
	}
	o := task{
		source: s,
		bar:    progressbar.Default(s.totalLines),
	}
	o.loadMotives(d, separator)
	return &o, nil
}
