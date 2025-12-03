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

	"github.com/cuducos/minha-receita/download"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/sync/errgroup"
)

const (
	// BatchSize determines the size of the batches used to create the initial JSON
	// data in the database.
	BatchSize = 8192

	// MaxParallelDBQueries is the default for maximum number of parallels save
	// queries sent to the database
	MaxParallelDBQueries = 8
)

var extraIndexes = [...]string{
	"cnae_fiscal",
	"cnaes_secundarios.codigo",
	"codigo_municipio",
	"codigo_municipio_ibge",
	"codigo_natureza_juridica",
	"qsa.cnpj_cpf_do_socio",
	"uf",
}

type database interface {
	PreLoad() error
	CreateCompanies([][]string) error
	PostLoad() error
	CreateExtraIndexes([]string) error
	MetaSave(string, string) error
}

func sources() map[string]*source { // all but Estabelecimentos (this one is loaded later on)
	srcs := []*source{
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
	m := make(map[string]*source)
	for _, src := range srcs {
		m[src.key] = src
	}
	return m
}

func newProgressBar(label string, srcs int) (*progressbar.ProgressBar, error) {
	bar := progressbar.NewOptions(
		srcs, // it has a bug starting At zero, so we compensate for it later
		progressbar.OptionFullWidth(),
		progressbar.OptionSetDescription(label),
		progressbar.OptionUseANSICodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionShowTotalBytes(true),
	)
	return bar, bar.RenderBlank()
}

func saveUpdatedAt(db database, dir string) error {
	slog.Info("Saving the updated at date to the database…")
	p := filepath.Join(dir, download.FederalRevenueUpdatedAt)
	v, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", p, err)

	}
	return db.MetaSave("updated-at", string(v))
}

func postLoad(db database) error {
	slog.Info("Consolidating the database…")
	if err := db.PostLoad(); err != nil {
		return err
	}
	slog.Info("Database consolidated!")
	slog.Info("Creating indexes…")
	if err := db.CreateExtraIndexes(extraIndexes[:]); err != nil {
		return err
	}
	slog.Info("Indexes created!")
	return nil
}

func Transform(dir string, db database, batch, maxDB int, privacy bool) error {
	if err := db.PreLoad(); err != nil {
		return err
	}
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
	bar, err := newProgressBar("[Step 1 of 2] Loading data to key-value storage", len(srcs))
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
	if err := writeJSONs(ctx, srcs, kv, db, maxDB, batch, dir, privacy); err != nil {
		return err
	}
	if err := postLoad(db); err != nil {
		return err
	}
	return saveUpdatedAt(db, dir)
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
