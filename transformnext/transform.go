package transformnext

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
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

func Cleanup() error {
	return filepath.WalkDir(os.TempDir(), func(pth string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			return nil
		}
		if !strings.HasPrefix(d.Name(), "minha-receita-") {
			return nil
		}
		part := strings.Split(d.Name(), "-")
		if len(part) != 4 {
			return nil
		}
		if _, err := time.Parse("20060102150405", part[2]); err != nil {
			return nil
		}
		fmt.Printf("Removing %s\n", pth)
		return os.RemoveAll(pth)
	})
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

	// Step 1: Load auxiliary data to key-value storage
	bar, err := newProgressBar("[Step 1 of 2] Loading auxiliary data to key-value storage", srcs)
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
	if err := g.Wait(); err != nil {
		return err
	}

	// Step 2: Process Estabelecimentos with enrichment and database writes
	return processEstabelecimentos(dir, kv)
}
