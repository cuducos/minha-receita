package transformnext

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/sync/errgroup"
)

func sources() []*source { // all but Estabelecimentos (this one is loaded later on)
	return []*source{
		newSource("Cnaes", ';', false, false),
		newSource("Empresas", ';', false, false),
		newSource("Imunes e Isentas", ',', true, true),
		newSource("Lucro Arbitrado", ',', true, true),
		newSource("Lucro Presumido", ',', true, true),
		newSource("Lucro Real", ',', true, true),
		newSource("Motivos", ';', false, false),
		newSource("Municipios", ';', false, false),
		newSource("Naturezas", ';', false, false),
		newSource("Paises", ';', false, false),
		newSource("Qualificacoes", ';', false, false),
		newSource("Simples", ';', false, false),
		newSource("Socios", ';', false, true),
		newSource("tabmun", ';', false, false),
	}
}

type progress struct {
	bar *progressbar.ProgressBar
}

func (p *progress) update(srcs []*source) error {
	var tot, read int64
	for _, src := range srcs {
		tot += src.total.Load()
		read += src.done.Load()
	}
	if tot == 0 {
		return nil
	}
	p.bar.ChangeMax64(tot)
	return p.bar.Set64(read)
}

func newProgressBar(label string) *progress {
	return &progress{progressbar.NewOptions(
		-1,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetDescription(label),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowTotalBytes(true),
		progressbar.OptionThrottle(time.Duration(1*time.Second)),
		progressbar.OptionFullWidth(),
		progressbar.OptionShowCount(),
	)}
}

func Transform(dir string) error {
	srcs := sources()
	var g errgroup.Group
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := make(chan []string)
	for _, src := range srcs {
		wg.Add(1)
		g.Go(func() error {
			defer wg.Done()
			return readCSVs(ctx, dir, src, ch)
		})
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	bar := newProgressBar("[Step 1 of 2] Loading data to key-value storage")
	for range ch {
		if err := bar.update(srcs); err != nil {
			return fmt.Errorf("could not update the progress bar: %w", err)
		}
	}
	return g.Wait()
}
