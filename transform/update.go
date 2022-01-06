package transform

import (
	"fmt"

	"github.com/schollz/progressbar/v3"
)

type updateTask struct {
	db      database
	source  *source
	lookups *lookups
	queues  []chan []string
	updated chan struct{}
	errors  chan error
	bar     *progressbar.ProgressBar
}

type updateFunc func(string, database, *lookups) error
type shardConsumerHandler func(*lookups, database, []string) error

func (t *updateTask) consumeShard(n int, h shardConsumerHandler) {
	for r := range t.queues[n] {
		if err := h(t.lookups, t.db, r); err != nil {
			t.errors <- fmt.Errorf("error processing %v: %w", r, err)
			continue
		}
		t.updated <- struct{}{}
	}
}

func (t *updateTask) close() {
	t.source.close()
	for _, q := range t.queues {
		close(q)
	}
	close(t.updated)
	close(t.errors)
}

func (t *updateTask) run(d string, h shardConsumerHandler) error {
	if t.source.totalLines == 0 {
		return nil
	}
	t.bar.Describe(d)
	if err := t.bar.RenderBlank(); err != nil {
		return fmt.Errorf("error rendering the progress bar: %w", err)
	}
	for n := 0; n < numOfShards; n++ {
		t.queues[n] = make(chan []string)
		go t.consumeShard(n, h)
	}
	for _, r := range t.source.readers {
		go r.sendLinesToShards(t.queues, t.errors)
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

func newUpdateTask(srcType sourceType, dir string, db database, l *lookups) (*updateTask, error) {
	s, err := newSource(srcType, dir)
	if err != nil {
		return nil, fmt.Errorf("error creating source for base cnpj: %w", err)
	}
	t := updateTask{
		db:      db,
		source:  s,
		lookups: l,
		queues:  make([]chan []string, numOfShards),
		updated: make(chan struct{}),
		errors:  make(chan error),
		bar:     progressbar.Default(s.totalLines),
	}
	return &t, nil
}
