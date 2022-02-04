package transform

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/schollz/progressbar/v3"
)

type line struct {
	content []string
	source  sourceType
}

type updateTask struct {
	db                database
	sources           []*source
	totalLines        int64
	lookups           *lookups
	batchSize         int
	queues            []chan line
	updated           chan int
	errors            chan error
	readersWaitGroup  sync.WaitGroup
	shutdownWaitGroup sync.WaitGroup
	shutdown          int32
	bar               *progressbar.ProgressBar
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

func (t *updateTask) sendBatch(s sourceType, b *[][]string) (int, error) {
	var f func([][]string) error
	if s == partners {
		f = t.db.AddPartners
	} else {
		f = t.db.UpdateCompanies
	}
	n := len(*b)
	if err := f(optimizeBatch(b)); err != nil {
		return 0, fmt.Errorf("error updating %s: %w", string(s), err)
	}
	return n, nil
}

func (t *updateTask) consumeShard(n int) {
	defer t.shutdownWaitGroup.Done()
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
		if err != nil { // initiate graceful shutdown
			t.errors <- fmt.Errorf("error processing %v: %w", l.content, err)
			atomic.StoreInt32(&t.shutdown, 1)
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
			c, err := t.sendBatch(l.source, b)
			if err != nil {
				t.errors <- fmt.Errorf("error sending update batch: %w", err)
				atomic.StoreInt32(&t.shutdown, 1)
				return
			}
			if atomic.LoadInt32(&t.shutdown) == 1 {
				return
			}
			t.updated <- c
		}
	}
	for _, b := range []*[][]string{&batches.partners, &batches.companies} {
		if len(*b) == 0 {
			continue
		}
		var src sourceType
		if b == &batches.partners {
			src = partners
		} else {
			src = taxes // this includes base cnpj batch too
		}
		c, err := t.sendBatch(src, b)
		if err != nil {
			t.errors <- fmt.Errorf("error sending the remaining update batch: %w", err)
			atomic.StoreInt32(&t.shutdown, 1)
			return
		}
		if atomic.LoadInt32(&t.shutdown) == 1 {
			return
		}
		t.updated <- c
	}
}

func (t *updateTask) sendLinesToShards(a *archivedCSV, s sourceType) {
	defer func() {
		t.readersWaitGroup.Done()
		t.shutdownWaitGroup.Done()
		a.close()
	}()
	for {
		if atomic.LoadInt32(&t.shutdown) == 1 {
			return
		}
		r, err := a.read()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.errors <- fmt.Errorf("error reading line %v: %w", r, err)
			atomic.StoreInt32(&t.shutdown, 1)
			return
		}
		n, err := shard(r[0])
		if err != nil {
			t.errors <- fmt.Errorf("error getting shard number for %s: %w", r[0], err)
			atomic.StoreInt32(&t.shutdown, 1)
			return
		}
		t.queues[n] <- line{r, s}
	}
}

func (t *updateTask) closeReaders() {
	defer fmt.Printf("Closing readers…\n")
	t.readersWaitGroup.Wait()
	for _, q := range t.queues {
		close(q)
	}
	for _, s := range t.sources {
		s.close()
	}
}

func (t *updateTask) close() {
	if atomic.LoadInt32(&t.shutdown) == 1 {
		t.shutdownWaitGroup.Wait()
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
		t.shutdownWaitGroup.Add(1)
		t.queues[n] = make(chan line)
		go t.consumeShard(n)
	}
	for _, s := range t.sources {
		for _, r := range s.readers {
			t.readersWaitGroup.Add(1)
			t.shutdownWaitGroup.Add(1)
			go t.sendLinesToShards(r, s.kind)
		}
	}
	go t.closeReaders()
	defer t.close()
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
	var t int64
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
		bar:        progressbar.Default(t),
	}
	return &u, nil
}
