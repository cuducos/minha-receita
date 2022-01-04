package transform

import (
	"fmt"
	"io"
	"log"

	"github.com/cuducos/go-cnpj"
	"github.com/schollz/progressbar/v3"
)

func addSimplesToCompany(db database, r []string) error {
	strs, err := db.ListCompanies(r[0])
	if err != nil {
		return fmt.Errorf("error loading companies with base %s: %w", r[0], err)
	}
	if len(strs) == 0 {
		log.Output(2, fmt.Sprintf("No company found for CNPJ base %s", r[0]))
		return nil
	}
	for _, s := range strs {
		c, err := companyFromString(s)
		if err != nil {
			return fmt.Errorf("error loading company: %w", err)
		}
		c.OpcaoPeloSimples = toBool(r[1])
		c.DataOpcaoPeloSimples, err = toDate(r[2])
		if err != nil {
			return fmt.Errorf("error parsing DataOpcaoPeloSimples %s: %w", r[2], err)
		}
		c.DataExclusaoDoSimples, err = toDate(r[3])
		if err != nil {
			return fmt.Errorf("error parsing DataExclusaoDoSimples %s: %w", r[3], err)
		}
		c.OpcaoPeloMEI = toBool(r[4])
		c.DataOpcaoPeloMEI, err = toDate(r[5])
		if err != nil {
			return fmt.Errorf("error parsing DataOpcaoPeloMEI %s: %w", r[5], err)
		}
		c.DataExclusaoDoMEI, err = toDate(r[6])
		if err != nil {
			return fmt.Errorf("error parsing DataExclusaoDoMEI %s: %w", r[6], err)
		}
		if err = c.Save(db); err != nil {
			return fmt.Errorf("error saving %s: %w", cnpj.Mask(c.CNPJ), err)
		}
	}
	return nil
}

type simplesTask struct {
	db      database
	queues  []chan []string
	success chan struct{}
	errors  chan error
	bar     *progressbar.ProgressBar
}

func (t *simplesTask) consumeShard(n int) {
	for r := range t.queues[n] {
		if err := addSimplesToCompany(t.db, r); err != nil {
			t.errors <- fmt.Errorf("error processing simples %v: %w", r, err)
			continue
		}
		t.success <- struct{}{}
	}
}

func addSimplesToCompanies(dir string, db database) error {
	s, err := newSource(simple, dir)
	if err != nil {
		return fmt.Errorf("error creating source for simples: %w", err)
	}
	defer s.close()
	if s.totalLines == 0 {
		return nil
	}

	t := simplesTask{
		db:      db,
		success: make(chan struct{}),
		errors:  make(chan error),
		bar:     progressbar.Default(s.totalLines),
	}
	t.bar.Describe("Adding Simples and MEI")
	for i := 0; i < numOfShards; i++ {
		t.queues = append(t.queues, make(chan []string))
	}
	for i := 0; i < numOfShards; i++ {
		go t.consumeShard(i)
	}
	for _, r := range s.readers {
		go func(a *archivedCSV, q []chan []string, e chan error) {
			defer a.close()
			for {
				r, err := a.read()
				if err == io.EOF {
					break
				}
				if err != nil {
					e <- fmt.Errorf("error reading line %v: %w", r, err)
					break // do not proceed in case of errors.
				}
				s, err := shard(r[0])
				if err != nil {
					e <- fmt.Errorf("error getting shard for %s: %w", r[0], err)
					break // do not proceed in case of errors.
				}
				t.queues[s] <- r
			}
		}(r, t.queues, t.errors)
	}

	defer func() {
		for _, q := range t.queues {
			close(q)
		}
		close(t.success)
		close(t.errors)
	}()

	for {
		select {
		case err := <-t.errors:
			return err
		case <-t.success:
			t.bar.Add(1)
			if t.bar.IsFinished() {
				return nil
			}
		}
	}
}
