package transform

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/cuducos/go-cnpj"
	"github.com/schollz/progressbar/v3"
)

type venuesTask struct {
	source    *source
	lookups   *lookups
	kv        kvStorage
	privacy   bool
	dir       string
	db        database
	batchSize int
	rows      chan []string
	saved     chan int
	errors    chan error
	consumers sync.WaitGroup
	bar       *progressbar.ProgressBar
}

func (t *venuesTask) saveBatch(b []Company) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	s := make([][]string, len(b))
	for i, c := range b {
		j, err := c.JSON()
		if err != nil {
			return 0, fmt.Errorf("error getting company %s as json: %w", cnpj.Mask(c.CNPJ), err)
		}
		s[i] = []string{c.CNPJ, j}
	}
	if err := t.db.CreateCompanies(s); err != nil {
		return 0, fmt.Errorf("error saving companies: %w", err)
	}
	return len(s), nil
}

func (t *venuesTask) consumeRows(ctx context.Context) error {
	ch := make(chan int)
	errs := make(chan error, 1)
	defer func() {
		t.consumers.Done()
		close(errs)
	}()
	go func() {
		var b []Company
		for {
			select {
			case <-ctx.Done():
				return
			case r, ok := <-t.rows:
				if !ok {
					n, err := t.saveBatch(b)
					if err != nil {
						errs <- err
						return
					}
					ch <- n
					return
				}
				c, err := newCompany(r, t.lookups, t.kv, t.privacy)
				if err != nil {
					errs <- fmt.Errorf("error parsing company from %q: %w", r, err)
					return
				}
				b = append(b, c)
				if len(b) < t.batchSize {
					continue
				}
				n, err := t.saveBatch(b)
				if err != nil {
					errs <- err
					return
				}
				ch <- n
				b = []Company{}
			}
		}
	}()
	for {
		select {
		case <-ctx.Done():
			close(ch)
			return nil
		case n, ok := <-ch:
			if !ok {
				return nil
			}
			t.saved <- n
		case err := <-errs:
			close(ch)
			return err
		}
	}
}

func (t *venuesTask) run(m int) error {
	defer t.source.close()
	if err := t.bar.RenderBlank(); err != nil {
		return fmt.Errorf("error rendering the progress bar: %w", err)
	}
	if err := t.db.PreLoad(); err != nil {
		return fmt.Errorf("error preparing the database: %w", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := t.source.sendTo(ctx, t.rows); err != nil {
			t.errors <- fmt.Errorf("error reading %s: %w", t.source.kind, err)
		}
		close(t.rows)
	}()
	for range m {
		t.consumers.Add(1)
		go t.consumeRows(ctx)
	}
	defer func() {
		t.consumers.Wait()
		close(t.saved)
		close(t.errors)
	}()
	for {
		select {
		case err := <-t.errors:
			cancel()
			return err
		case n := <-t.saved:
			t.bar.Add(n)
			if t.bar.IsFinished() {
				cancel()
				log.Output(1, "Consolidating the database…")
				return t.db.PostLoad()
			}
		}
	}
}

func createJSONRecordsTask(dir string, db database, l *lookups, kv kvStorage, b int, p bool) (*venuesTask, error) {
	v, err := newSource(context.Background(), venues, dir)
	if err != nil {
		return nil, fmt.Errorf("error creating a source for venues from %s: %w", dir, err)
	}
	t := venuesTask{
		source:    v,
		lookups:   l,
		kv:        kv,
		privacy:   p,
		dir:       dir,
		db:        db,
		batchSize: b,
		rows:      make(chan []string),
		saved:     make(chan int),
		errors:    make(chan error, 1),
		bar:       progressbar.Default(int64(v.total)),
	}
	t.bar.Describe("Creating the JSON data for each CNPJ")
	return &t, nil
}
