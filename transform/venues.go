package transform

import (
	"fmt"
	"io"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/cuducos/go-cnpj"
	"github.com/schollz/progressbar/v3"
)

func saveBatch(db database, b []company) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	s := make([][]any, len(b))
	for i, c := range b {
		j, err := c.JSON()
		if err != nil {
			return 0, fmt.Errorf("error getting company %s as json: %w", cnpj.Mask(c.CNPJ), err)
		}
		n, err := strconv.Atoi(c.CNPJ)
		if err != nil {
			return 0, fmt.Errorf("copuld not convert cnpj %s to int: %w", c.CNPJ, err)
		}
		s[i] = []any{n, j}
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
	sentToBatches     int
	rows              chan []string
	companies         chan struct{}
	saved             chan int
	errors            chan error
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
		b = append(b, c)
		t.companies <- struct{}{}
		if len(b) >= t.batchSize {
			n, err := saveBatch(t.db, b)
			if err != nil { // initiate graceful shutdown.
				t.errors <- fmt.Errorf("error saving companies: %w", err)
				atomic.StoreInt32(&t.shutdown, 1)
				return
			}
			t.saved <- n
			b = []company{}
		}
	}
	if len(b) == 0 || atomic.LoadInt32(&t.shutdown) == 1 { // check if must continue.
		return
	}
	// send the remaining items in the batch
	n, err := saveBatch(t.db, b)
	if err != nil { // initiate graceful shutdown.
		t.errors <- fmt.Errorf("error saving companies: %w", err)
		atomic.StoreInt32(&t.shutdown, 1)
		return
	}
	t.saved <- n
}

func (t *venuesTask) run(m int) error {
	defer t.source.close()
	if err := t.bar.RenderBlank(); err != nil {
		return fmt.Errorf("error rendering the progress bar: %w", err)
	}
	t.produceRows()
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
				return t.db.CreateIndex()
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
		bar:           progressbar.Default(int64(v.totalLines)),
	}
	t.bar.Describe("Creating the JSON data for each CNPJ")
	return &t, nil
}
