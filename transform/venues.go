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
			return 0, fmt.Errorf("could not convert cnpj %s to int: %w", c.CNPJ, err)
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
	kv                kvStorage
	privacy           bool
	dir               string
	db                database
	batchSize         int
	sentToBatches     int32
	rows              chan []string
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
				if atomic.LoadInt32(&t.shutdown) == 1 {
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
		if int(atomic.AddInt32(&t.sentToBatches, 1)) == t.source.totalLines {
			close(t.rows)
		}
		c, err := newCompany(r, t.lookups, t.kv, t.privacy)
		if err != nil { // initiate graceful shutdown.
			t.errors <- fmt.Errorf("error parsing company from %q: %w", r, err)
			atomic.StoreInt32(&t.shutdown, 1)
			return
		}
		b = append(b, c)
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
	if err := t.db.PreLoad(); err != nil {
		return fmt.Errorf("error preparing the database: %w", err)
	}
	t.produceRows()
	for i := 0; i < m; i++ {
		t.shutdownWaitGroup.Add(1)
		go t.consumeRows()
	}
	defer func() {
		if t.source.totalLines != int(t.sentToBatches) {
			close(t.rows)
		}
		if atomic.LoadInt32(&t.shutdown) == 1 {
			t.shutdownWaitGroup.Wait()
		}
		close(t.saved)
		close(t.errors)
	}()
	for {
		select {
		case err := <-t.errors:
			return err
		case n := <-t.saved:
			t.bar.Add(n)
			if t.bar.IsFinished() {
				return t.db.PostLoad()
			}
		}
	}
}

func createJSONRecordsTask(dir string, db database, l *lookups, kv kvStorage, b int, p bool) (*venuesTask, error) {
	v, err := newSource(venues, dir)
	if err != nil {
		return nil, fmt.Errorf("error creating a source for venues from %s: %w", dir, err)
	}
	t := venuesTask{
		source:        v,
		lookups:       l,
		kv:            kv,
		privacy:       p,
		dir:           dir,
		db:            db,
		batchSize:     b,
		sentToBatches: 0,
		rows:          make(chan []string),
		saved:         make(chan int),
		errors:        make(chan error),
		bar:           progressbar.Default(int64(v.totalLines)),
	}
	t.bar.Describe("Creating the JSON data for each CNPJ")
	return &t, nil
}
