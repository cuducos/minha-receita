package transform

import (
	"fmt"
	"io"

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
	source        *source
	lookups       *lookups
	dir           string
	db            database
	batchSize     int
	sentToBatches int64
	rows          chan []string
	companies     chan struct{}
	saved         chan int
	errors        chan error
	bar           *progressbar.ProgressBar
}

func (t *venuesTask) produceRows() {
	for _, r := range t.source.readers {
		go func(t *venuesTask, a *archivedCSV) {
			for {
				r, err := a.read()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.errors <- err
					break // do not proceed in case of errors.
				}
				t.rows <- r
			}
		}(t, r)
	}
}

func (t *venuesTask) consumeRows() {
	var b []company
	defer func() { // send the remaining items in the batch
		n, err := saveBatch(t.db, b)
		if err != nil {
			t.errors <- fmt.Errorf("error saving companies: %w", err)
		}
		t.saved <- n
	}()
	for r := range t.rows {
		c, err := newCompany(r, t.lookups)
		if err != nil {
			t.errors <- fmt.Errorf("error parsing company from %q: %w", r, err)
			break
		}
		b = append(b, c)
		t.companies <- struct{}{}
		if len(b) >= t.batchSize {
			n, err := saveBatch(t.db, b)
			if err != nil {
				t.errors <- fmt.Errorf("error saving companies: %w", err)
				break
			}
			t.saved <- n
			b = []company{}
		}
	}
}

func (t *venuesTask) run(m int) error {
	defer t.source.close()
	if err := t.bar.RenderBlank(); err != nil {
		return fmt.Errorf("error rendering the progress bar: %w", err)
	}
	t.produceRows()
	for i := 0; i < m; i++ {
		go t.consumeRows()
	}
	defer func() {
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

func createJSONRecordsTask(dir string, db database, b int) (*venuesTask, error) {
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
