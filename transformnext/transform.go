package transformnext

import (
	"context"
	"log/slog"
	"sync"

	"golang.org/x/sync/errgroup"
)

func Transform(dir string) error {
	var g errgroup.Group
	var wg sync.WaitGroup
	ch := make(chan []string)
	for _, p := range []string{"tabmun", "Empresas1"} {
		wg.Add(1)
		g.Go(func() error {
			defer wg.Done()
			return readCSVs(
				context.Background(),
				dir,
				p,
				';',
				false,
				ch,
			)
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
