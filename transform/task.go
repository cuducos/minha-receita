package transform

import (
	"fmt"
	"io"

	"github.com/schollz/progressbar/v3"
)

type task struct {
	source  *source
	queue   chan []string
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
		go t.rowsFrom(r)
	}
}

func (t *task) rowsFrom(a *archivedCSV) {
	for {
		r, err := a.read()
		if err == io.EOF {
			break
		}
		t.queue <- r
	}
}

func (t *task) consumeRows() {
	for r := range t.queue {
		t.consumeRow(r)
	}
}

func (t *task) consumeRow(r []string) {
	c, err := newCompany(r, t.motives)
	if err != nil {
		t.errors <- fmt.Errorf("error parsing company from %q: %w", r, err)
		return
	}

	_, err = c.toJSON(t.source.dir)
	if err != nil {
		t.errors <- fmt.Errorf("error getting the JSON bytes for %v: %w", c, err)
		return
	}
	t.errors <- nil
}

func (t *task) run(m int) error {
	defer t.source.close()
	go t.produceRows()
	for i := 0; i < m; i++ {
		go t.consumeRows()
	}
	for i := int64(0); i < t.source.totalLines; i++ {
		if err := <-t.errors; err != nil {
			return fmt.Errorf("error while consuming row: %w", err)
		}
	}
	return nil
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
