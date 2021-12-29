package transform

import (
	"fmt"
	"io"
	"log"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
)

func addSimplesToCompany(dir string, r []string) error {
	b, err := pathForBaseCNPJ(r[0])
	if err != nil {
		return fmt.Errorf("error getting the path for %s: %w", r[0], err)
	}
	ls, err := filepath.Glob(filepath.Join(dir, b, "*.json"))
	if err != nil {
		return fmt.Errorf("error in the glob pattern: %w", err)
	}
	if len(ls) == 0 {
		log.Output(2, fmt.Sprintf("No JSON file found for CNPJ base %s", r[0]))
		return nil
	}
	for _, f := range ls {
		c, err := companyFromJSON(f)
		if err != nil {
			return fmt.Errorf("error reading company from %s: %w", f, err)
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
		f, err = c.toJSON(dir)
		if err != nil {
			return fmt.Errorf("error updating json file for %s: %w", c.CNPJ, err)
		}
	}
	return nil
}

type simplesTask struct {
	dir     string
	queues  []chan []string
	success chan struct{}
	errors  chan error
	bar     *progressbar.ProgressBar
}

func (t *simplesTask) consumeShard(n int) {
	for r := range t.queues[n] {
		if err := addSimplesToCompany(t.dir, r); err != nil {
			t.errors <- fmt.Errorf("error processing simples %v: %w", r, err)
			continue
		}
		t.success <- struct{}{}
	}
}

func addSimplesToCompanies(dir string) error {
	s, err := newSource(simple, dir)
	if err != nil {
		return fmt.Errorf("error creating source for simples: %w", err)
	}
	defer s.close()
	if s.totalLines == 0 {
		return nil
	}

	t := simplesTask{
		dir:     dir,
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
