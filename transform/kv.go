package transform

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/cuducos/go-cnpj"
	"github.com/dgraph-io/badger/v4"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/sync/errgroup"
)

type item struct {
	key, value []byte
	kind       sourceType
}

func checksumFor(r []string) string {
	b := []byte(strings.Join(r, ""))
	h := md5.New()
	return hex.EncodeToString(h.Sum(b))
}

func newKVItem(s sourceType, l *lookups, r []string) (i item, err error) {
	var k string
	var h func(l *lookups, r []string) ([]byte, error)
	switch s {
	case partners:
		k = keyForPartners(r[0])
		h = loadPartnerRow
	case base:
		k = keyForBase(r[0])
		h = loadBaseRow
	case simpleTaxes:
		k = keyForSimpleTaxes(r[0])
		h = loadSimpleTaxesRow
	case realProfit:
		k = keyForTaxRegime(r[1])
		h = loadTaxRow
	case presumedProfit:
		k = keyForTaxRegime(r[1])
		h = loadTaxRow
	case arbitratedProfit:
		k = keyForTaxRegime(r[1])
		h = loadTaxRow
	case noTaxes:
		k = keyForTaxRegime(r[1])
		h = loadTaxRow
	default:
		return item{}, fmt.Errorf("unknown source type %s", string(s))
	}
	if s.isAccumulative() {
		k = k + ":" + checksumFor(r)
	}
	i.key = []byte(k)
	i.value, err = h(l, r)
	if err != nil {
		return item{}, fmt.Errorf("error loading value from source: %w", err)
	}
	i.kind = s
	return i, nil
}

type badgerStorage struct {
	db   *badger.DB
	path string
}

func (kv *badgerStorage) garbageCollect() {
	for {
		err := kv.db.RunValueLogGC(0.5)
		if err == badger.ErrRejected { // db already closed or more than one gc running
			return
		}
		if err == badger.ErrNoRewrite { // no garbage to collect
			return
		}
		if err != nil {
			slog.Error("Error running garbage collection", "error", err)
			return
		}
	}
}

func (kv *badgerStorage) loadRow(r []string, s sourceType, l *lookups) error {
	i, err := newKVItem(s, l, r)
	if err != nil {
		return fmt.Errorf("error creating an %s item: %w", s, err)
	}
	if err := kv.db.Update(func(tx *badger.Txn) error { return tx.Set(i.key, i.value) }); err != nil {
		return fmt.Errorf("could not save key-value: %w", err)
	}
	return nil
}

func (kv *badgerStorage) loadSource(ctx context.Context, s *source, l *lookups, bar *progressbar.ProgressBar, m int) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(m)
	ch := make(chan []string)
	g.Go(func() error {
		defer close(ch)
		err := s.sendTo(ctx, ch)
		if err == io.EOF {
			return nil
		}
		return err
	})
	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case r, ok := <-ch:
				if !ok {
					return nil
				}
				g.Go(func() error {
					if err := kv.loadRow(r, s.kind, l); err != nil {
						return err
					}
					bar.Add(1)
					return nil
				})
			}

		}
	})
	return g.Wait()
}

func (kv *badgerStorage) load(dir string, l *lookups, m int) error {
	srcs, t, err := newSources(dir, []sourceType{
		base,
		partners,
		simpleTaxes,
		noTaxes,
		presumedProfit,
		realProfit,
		arbitratedProfit,
	})
	if err != nil {
		return fmt.Errorf("could not load sources: %w", err)
	}
	tic := time.NewTicker(3 * time.Minute)
	defer tic.Stop()
	go func() {
		for range tic.C {
			kv.garbageCollect()
		}
	}()
	bar := progressbar.Default(t, "Processing base CNPJ, partners and taxes")
	defer bar.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)
	for _, src := range srcs {
		g.Go(func() error {
			return kv.loadSource(ctx, src, l, bar, m)
		})
	}
	return g.Wait()
}

func (kv *badgerStorage) enrichCompany(c *Company) error {
	n := cnpj.Base(c.CNPJ)
	ps := make(chan []PartnerData)
	bs := make(chan baseData)
	st := make(chan simpleTaxesData)
	tr := make(chan TaxRegimes)
	errs := make(chan error)
	go func() {
		p, err := partnersOf(kv.db, n)
		if err != nil {
			errs <- err
		}
		ps <- p
	}()
	go func() {
		v, err := baseOf(kv.db, n)
		if err != nil {
			errs <- err
		}
		bs <- v
	}()
	go func() {
		t, err := simpleTaxesOf(kv.db, n)
		if err != nil {
			errs <- err
		}
		st <- t
	}()
	go func() {
		t, err := taxRegimeOf(kv.db, c.CNPJ)
		if err != nil {
			errs <- err
		}
		tr <- t
	}()
	for range [4]int{} {
		select {
		case p := <-ps:
			c.QuadroSocietario = p
		case v := <-bs:
			c.CodigoPorte = v.CodigoPorte
			c.Porte = v.Porte
			c.RazaoSocial = v.RazaoSocial
			c.CodigoNaturezaJuridica = v.CodigoNaturezaJuridica
			c.NaturezaJuridica = v.NaturezaJuridica
			c.QualificacaoDoResponsavel = v.QualificacaoDoResponsavel
			c.CapitalSocial = v.CapitalSocial
			c.EnteFederativoResponsavel = v.EnteFederativoResponsavel
		case t := <-st:
			c.OpcaoPeloSimples = t.OpcaoPeloSimples
			c.DataOpcaoPeloSimples = t.DataOpcaoPeloSimples
			c.DataExclusaoDoSimples = t.DataExclusaoDoSimples
			c.OpcaoPeloMEI = t.OpcaoPeloMEI
			c.DataOpcaoPeloMEI = t.DataOpcaoPeloMEI
			c.DataExclusaoDoMEI = t.DataExclusaoDoMEI
		case t := <-tr:
			c.RegimeTributario = t
		case err := <-errs:
			return fmt.Errorf("error enriching company: %w", err)
		}
	}
	return nil
}

func (b *badgerStorage) close() error {
	return b.db.Close()
}

type noLogger struct{}

func (*noLogger) Errorf(string, ...any)   {}
func (*noLogger) Warningf(string, ...any) {}
func (*noLogger) Infof(string, ...any)    {}
func (*noLogger) Debugf(string, ...any)   {}

func newBadgerStorage(dir string, ro bool) (*badgerStorage, error) {
	opt := badger.DefaultOptions(dir).WithReadOnly(ro)
	slog.Debug("Creating temporary key-value storage", "path", dir)
	if os.Getenv("DEBUG") == "" {
		opt = opt.WithLogger(&noLogger{})
	}
	db, err := badger.Open(opt)
	if err != nil {
		return nil, fmt.Errorf("error creating badger key-value object: %w", err)
	}
	return &badgerStorage{db: db, path: dir}, nil
}
