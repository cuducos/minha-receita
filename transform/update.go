package transform

import (
	"fmt"
	"io"

	"github.com/schollz/progressbar/v3"
)

type line struct {
	content []string
	source  sourceType
}

type updateTask struct {
	db         database
	sources    []*source
	totalLines int64
	lookups    *lookups
	queues     []chan line
	updated    chan struct{}
	errors     chan error
	bar        *progressbar.ProgressBar
}

type updateFunc func(string, database, *lookups) error
type shardConsumerHandler func(*lookups, database, []string) error

func (t *updateTask) consumeShard(n int) {
	for l := range t.queues[n] {
		var h shardConsumerHandler
		switch l.source {
		case base:
			h = addBase
		case partners:
			h = addPartner
		case taxes:
			h = addTax
		}
		if err := h(t.lookups, t.db, l.content); err != nil {
			t.errors <- fmt.Errorf("error processing %v: %w", l.content, err)
			continue
		}
		t.updated <- struct{}{}
	}
}

func (t *updateTask) sendLinesToShards(a *archivedCSV, s sourceType) {
	defer a.close()
	for {
		r, err := a.read()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.errors <- fmt.Errorf("error reading line %v: %w", r, err)
			break
		}
		n, err := shard(r[0])
		if err != nil {
			t.errors <- fmt.Errorf("error getting shard for %s: %w", r[0], err)
			break
		}
		t.queues[n] <- line{r, s}
	}
}

func (t *updateTask) close() {
	for _, s := range t.sources {
		s.close()
	}
	for _, q := range t.queues {
		close(q)
	}
	close(t.updated)
	close(t.errors)
}

func (t *updateTask) run() error {
	if t.totalLines == 0 {
		return nil
	}
	t.bar.Describe("Adding base CNPJ, partners and taxes info")
	if err := t.bar.RenderBlank(); err != nil {
		return fmt.Errorf("error rendering the progress bar: %w", err)
	}
	for n := 0; n < numOfShards; n++ {
		t.queues[n] = make(chan line)
		go t.consumeShard(n)
	}
	for _, s := range t.sources {
		for _, r := range s.readers {
			go t.sendLinesToShards(r, s.kind)
		}
	}
	defer t.close()
	for {
		select {
		case err := <-t.errors:
			return err
		case <-t.updated:
			t.bar.Add(1)
			if t.bar.IsFinished() {
				return nil
			}
		}
	}
}

func newUpdateTask(dir string, db database, l *lookups) (*updateTask, error) {
	srcs := make([]*source, 3)
	for i, t := range []sourceType{base, partners, taxes} {
		s, err := newSource(t, dir)
		if err != nil {
			return nil, fmt.Errorf("error creating source for base cnpj: %w", err)
		}
		srcs[i] = s
	}
	var t int64
	for _, s := range srcs {
		t += s.totalLines
	}
	u := updateTask{
		db:         db,
		sources:    srcs,
		totalLines: t,
		lookups:    l,
		queues:     make([]chan line, numOfShards),
		updated:    make(chan struct{}),
		errors:     make(chan error),
		bar:        progressbar.Default(t),
	}
	return &u, nil
}
