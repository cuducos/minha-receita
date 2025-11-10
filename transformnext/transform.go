package transformnext

import (
	"context"
	"log/slog"
	"sync"

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

func Transform(dir string) error {
	srcs := sources()
	var g errgroup.Group
	var wg sync.WaitGroup
	ctx := context.Background()
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
	for r := range ch {
		slog.Info("got", "row", r)
	}
	return g.Wait()
}
