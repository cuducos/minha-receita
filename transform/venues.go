package transform

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"

	"github.com/cuducos/go-cnpj"
	"github.com/schollz/progressbar/v3"
)

func saveBatch(db database, b []company) (int, error) {
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
	if err := db.CreateCompanies(s); err != nil {
		return 0, fmt.Errorf("error saving companies: %w", err)
	}
	return len(s), nil
}

type venuesTask struct {
	source            *source
	lookups           *lookups
	privacy           bool
	dir               string
	db                database
	batchSize         int
	sentToBatches     int64
	rows              chan []string
	companies         chan struct{}
	saved             chan int
	errors            chan error
	cache             *cache
	bar               *progressbar.ProgressBar
	shutdown          int32
	shutdownWaitGroup sync.WaitGroup
}

func (t *venuesTask) produceRows() {
	for _, r := range t.source.readers {
		t.shutdownWaitGroup.Add(1)
		go func(t *venuesTask, a *archivedCSV) {
			defer t.shutdownWaitGroup.Done()
			for {
				if atomic.LoadInt32(&t.shutdown) == 1 { // check if must continue.
					return
				}
				r, err := a.read()
				if err == io.EOF {
					break
				}
				if err != nil { // initiate graceful shutdown.
					t.errors <- err
					atomic.StoreInt32(&t.shutdown, 1)
					return
				}
				t.rows <- r
			}
		}(t, r)
	}
}

func (t *venuesTask) consumeRows() {
	defer t.shutdownWaitGroup.Done()
	var b []company
	var s int
	for r := range t.rows {
		if atomic.LoadInt32(&t.shutdown) == 1 { // check if must continue.
			return
		}
		c, err := newCompany(r, t.lookups, t.privacy)
		if err != nil { // initiate graceful shutdown.
			t.errors <- fmt.Errorf("error parsing company from %q: %w", r, err)
			atomic.StoreInt32(&t.shutdown, 1)
			return
		}
		exist, err := t.cache.check(c.CNPJ)
		if err != nil {
			t.errors <- fmt.Errorf("error checking cache for company %s: %w", c.CNPJ, err)
			atomic.StoreInt32(&t.shutdown, 1)
			return
		}
		if exist {
			s++
		} else {
			b = append(b, c)
		}
		t.companies <- struct{}{}
		if len(b) >= t.batchSize {
			n, err := saveBatch(t.db, b)
			if err != nil { // initiate graceful shutdown.
				t.errors <- fmt.Errorf("error saving companies: %w", err)
				atomic.StoreInt32(&t.shutdown, 1)
				return
			}
			t.saved <- n + s
			b = []company{}
			s = 0
		}
	}
	if len(b) == 0 || atomic.LoadInt32(&t.shutdown) == 1 { // check if must continue.
		t.saved <- s
		return
	}
	// send the remaining items in the batch
	n, err := saveBatch(t.db, b)
	if err != nil { // initiate graceful shutdown.
		t.errors <- fmt.Errorf("error saving companies: %w", err)
		atomic.StoreInt32(&t.shutdown, 1)
		return
	}
	t.saved <- n + s
}

func (t *venuesTask) run(m int) error {
	defer t.source.close()
	if err := t.bar.RenderBlank(); err != nil {
		return fmt.Errorf("error rendering the progress bar: %w", err)
	}
	t.produceRows()

	tmp, err := os.MkdirTemp("", "minha-receita")
	if err != nil {
		return fmt.Errorf("error creating tmp dir for cache: %w", err)
	}
	defer os.RemoveAll(tmp)
	t.cache = &cache{dir: tmp}

	for i := 0; i < m; i++ {
		t.shutdownWaitGroup.Add(1)
		go t.consumeRows()
	}
	defer func() {
		if atomic.LoadInt32(&t.shutdown) == 1 {
			t.shutdownWaitGroup.Wait()
		}
		close(t.companies)
		close(t.saved)
		close(t.errors)
	}()
	for {
		select {
		case err := <-t.errors:
			return err
		case <-t.companies:
			t.sentToBatches++
			if t.source.totalLines == t.sentToBatches {
				close(t.rows)
			}
		case n := <-t.saved:
			t.bar.Add(n)
			if t.bar.IsFinished() {
				return nil
			}
		}
	}
}

func createJSONRecordsTask(dir string, db database, b int, p bool) (*venuesTask, error) {
	v, err := newSource(venues, dir)
	if err != nil {
		return nil, fmt.Errorf("error creating a source for venues from %s: %w", dir, err)
	}
	l, err := newLookups(dir)
	if err != nil {
		return nil, fmt.Errorf("error creating look up tables from %s: %w", dir, err)
	}
	t := venuesTask{
		source:        v,
		lookups:       &l,
		privacy:       p,
		dir:           dir,
		db:            db,
		batchSize:     b,
		sentToBatches: 0,
		rows:          make(chan []string),
		companies:     make(chan struct{}),
		saved:         make(chan int),
		errors:        make(chan error),
		bar:           progressbar.Default(v.totalLines),
	}
	t.bar.Describe("Creating the JSON data for each CNPJ")
	return &t, nil
}
