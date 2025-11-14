package transformnext

import (
	"context"
	"fmt"
	"log/slog"
	"os"
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

func newProgressBar(label string, srcs []*source) (*progressbar.ProgressBar, error) {
	bar := progressbar.NewOptions(
		len(srcs), // it has a bug starting at zero, so we compensate for it later
		progressbar.OptionFullWidth(),
		progressbar.OptionSetDescription(label),
		progressbar.OptionUseANSICodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionShowTotalBytes(true),
	)
	return bar, bar.RenderBlank()
}

func Transform(dir string) error {
	srcs := sources()
	tmp, err := os.MkdirTemp("", fmt.Sprintf("minha-receita-%s-*", time.Now().Format("20060102150405")))
	if err != nil {
		return fmt.Errorf("could not create temporary directory for badger: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tmp); err != nil {
			slog.Warn("could not remove badger temporary directory", "path", tmp, "error", err)
		}
	}()
	kv, err := newBadger(tmp, false)
	if err != nil {
		return fmt.Errorf("could not create badger database: %w", err)
	}
	defer func() {
		if err := kv.db.Close(); err != nil {
			slog.Warn("could not close badger database", "error", err)
		}
	}()
	bar, err := newProgressBar("[Step 1 of 2] Loading data to key-value storage", srcs)
	if err != nil {
		return fmt.Errorf("could not create a progress bar: %w", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var g errgroup.Group
	for _, src := range srcs {
		src := src
		g.Go(func() error {
			return loadCSVs(ctx, dir, src, bar, kv)
		})
	}
	return g.Wait()
}
