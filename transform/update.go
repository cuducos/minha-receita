package transform

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/avast/retry-go/v4"
	"github.com/schollz/progressbar/v3"
)

type line struct {
	content []string
	source  sourceType
}

type updateTask struct {
	db         database
	sources    []*source
	totalLines int
	lookups    *lookups
	batchSize  int
	queues     []chan line
	updated    chan int
	errors     chan error
	shutdown   int32
	bar        *progressbar.ProgressBar
}

// optimize batch merges updates of the same base cnpj in the a single update
func optimizeBatch(b *[][]string) [][]string {
	m := make(map[string]string)
	for _, u := range *b {
		json, exists := m[u[0]]
		if exists { // append to json array, or merge json objects
			json = json[:len(json)-1] + ", " + u[1][1:]
		} else {
			json = u[1]
		}
		m[u[0]] = json
	}
	var c int
	n := make([][]string, len(m))
	for base, json := range m {
		n[c] = []string{base, json}
		delete(m, base)
		c++
	}
	*b = [][]string{}
	return n
}

func (t *updateTask) sendBatch(s sourceType, b *[][]string) {
	n := len(*b)
	if n == 0 {
		return
	}
	var f func([][]string) error
	if s == partners {
		f = t.db.AddPartners
	} else {
		f = t.db.UpdateCompanies
	}
	err := retry.Do(
		func() error { return f(optimizeBatch(b)) },
		retry.Attempts(32),
	)
	if err != nil {
		if atomic.LoadInt32(&t.shutdown) != 1 {
			t.errors <- fmt.Errorf("error sending a batch of updates: %w", err)
			atomic.StoreInt32(&t.shutdown, 1)
		}
		return
	}
	if atomic.LoadInt32(&t.shutdown) != 1 {
		t.updated <- n
	}
}

func (t *updateTask) consumeShard(n int) {
	var batches struct {
		companies [][]string
		partners  [][]string
	}
	for l := range t.queues[n] {
		if atomic.LoadInt32(&t.shutdown) == 1 {
			return
		}
		var h func(*lookups, []string) ([]string, error)
		switch l.source {
		case base:
			h = addBase
		case partners:
			h = addPartners
		case taxes:
			h = addTax
		}
		u, err := h(t.lookups, l.content)
		if err != nil {
			if atomic.LoadInt32(&t.shutdown) != 1 {
				t.errors <- fmt.Errorf("error processing %v: %w", l.content, err)
				atomic.StoreInt32(&t.shutdown, 1)
			}
			return
		}
		var b *[][]string
		if l.source == partners {
			b = &batches.partners
		} else {
			b = &batches.companies
		}
		*b = append(*b, u)
		if len(*b) >= t.batchSize {
			t.sendBatch(l.source, b)
		}
	}
	t.sendBatch(base, &batches.companies)
	t.sendBatch(partners, &batches.partners)
}

func (t *updateTask) sendLinesToShards(s *source) {
	var wg sync.WaitGroup
	for _, a := range s.readers {
		wg.Add(1)
		go func(a *archivedCSV) {
			defer wg.Done()
			for {
				if atomic.LoadInt32(&t.shutdown) == 1 {
					return
				}
				r, err := a.read()
				if err == io.EOF {
					return
				}
				if err != nil {
					if atomic.LoadInt32(&t.shutdown) != 0 {
						t.errors <- fmt.Errorf("error reading line %v: %w", r, err)
						atomic.StoreInt32(&t.shutdown, 1)
					}
					return
				}
				n, err := shard(r[0])
				if err != nil {
					if atomic.LoadInt32(&t.shutdown) != 0 {
						t.errors <- fmt.Errorf("error getting shard number for %s: %w", r[0], err)
						atomic.StoreInt32(&t.shutdown, 1)
					}
					return
				}
				if atomic.LoadInt32(&t.shutdown) != 1 {
					t.queues[n] <- line{r, s.kind}
				}
			}
		}(a)
	}
	wg.Wait()
}

func (t *updateTask) closeReaders() {
	for _, q := range t.queues {
		close(q)
	}
	for _, s := range t.sources {
		s.close()
	}
}

func (t *updateTask) run() error {
	if t.totalLines == 0 {
		return nil
	}
	if err := t.bar.RenderBlank(); err != nil {
		return fmt.Errorf("error rendering the progress bar: %w", err)
	}
	for n := 0; n < numOfShards; n++ {
		t.queues[n] = make(chan line)
		go t.consumeShard(n)
	}
	var wg sync.WaitGroup
	for _, s := range t.sources {
		wg.Add(1)
		go func(s *source) {
			t.sendLinesToShards(s)
			wg.Done()
		}(s)
	}
	go func() {
		defer t.closeReaders()
		wg.Wait()
	}()
	defer func() {
		close(t.updated)
		close(t.errors)
	}()
	for {
		select {
		case err := <-t.errors:
			return err
		case n := <-t.updated:
			t.bar.Add(n)
			if t.bar.IsFinished() {
				return nil
			}
		}
	}
}

func newUpdateTask(dir string, db database, b int, l *lookups) (*updateTask, error) {
	srcs := make([]*source, 3)
	for i, t := range []sourceType{base, partners, taxes} {
		s, err := newSource(t, dir)
		if err != nil {
			return nil, fmt.Errorf("error creating source for base cnpj: %w", err)
		}
		srcs[i] = s
	}
	var t int
	for _, s := range srcs {
		t += s.totalLines
	}
	u := updateTask{
		db:         db,
		sources:    srcs,
		totalLines: t,
		lookups:    l,
		batchSize:  b,
		queues:     make([]chan line, numOfShards),
		updated:    make(chan int),
		errors:     make(chan error),
		bar:        progressbar.Default(int64(t)),
	}
	u.bar.Describe("Adding base CNPJ, partners and taxes info")
	return &u, nil
}
