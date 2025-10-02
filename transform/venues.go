package transform

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/cuducos/go-cnpj"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/sync/errgroup"
)

type venuesTask struct {
	source    *source
	lookups   *lookups
	kv        kvStorage
	privacy   bool
	dir       string
	db        database
	batchSize int
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

func (t *venuesTask) consumeRows(ctx context.Context, q <-chan []string, done chan<- int) error {
	ch := make(chan int)
	errs := make(chan error, 1)
	defer close(errs)
	go func() {
		var b []Company
		for {
			select {
			case <-ctx.Done():
				return
			case r, ok := <-q:
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
			done <- n
		case err := <-errs:
			close(ch)
			return err
		}
	}
}

func (t *venuesTask) run(m int) error {
	bar := progressbar.Default(int64(t.source.total))
	bar.Describe("Creating the JSON data for each CNPJ")
	defer func() {
		if err := t.source.close(); err != nil {
			slog.Warn("could not close source files", "error", err)
		}
	}()
	if err := bar.RenderBlank(); err != nil {
		return fmt.Errorf("error rendering the progress bar: %w", err)
	}
	if err := t.db.PreLoad(); err != nil {
		return fmt.Errorf("error preparing the database: %w", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var g errgroup.Group
	q := make(chan []string)
	g.Go(func() error {
		if err := t.source.sendTo(ctx, q); err != nil {
			return fmt.Errorf("error reading %s: %w", t.source.kind, err)
		}
		close(q)
		return nil
	})
	ch := make(chan int)
	close(ch)
	for range m {
		g.Go(func() error {
			return t.consumeRows(ctx, q, ch)
		})
	}
	errs := make(chan error, 1)
	defer close(errs)
	go func() {
		if err := g.Wait(); err != nil {
			errs <- err
		}
	}()
	for {
		select {
		case err := <-errs:
			return err
		case n := <-ch:
			if err := bar.Add(n); err != nil {
				return err
			}
			if bar.IsFinished() {
				return nil
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
	}
	return &t, nil
}
