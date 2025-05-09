package transform

import (
	"context"
	"fmt"
	"log"
	"os"
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

func newKVItem(s sourceType, l *lookups, r []string) (i item, err error) {
	var h func(l *lookups, r []string) ([]byte, error)
	switch s {
	case partners:
		i.key = []byte(keyForPartners(r[0]))
		h = loadPartnerRow
	case base:
		i.key = []byte(keyForBase(r[0]))
		h = loadBaseRow
	case simpleTaxes:
		i.key = []byte(keyForSimpleTaxes(r[0]))
		h = loadSimpleTaxesRow
	case realProfit:
		i.key = []byte(keyForTaxRegime(r[1]))
		h = loadTaxRow
	case presumedProfit:
		i.key = []byte(keyForTaxRegime(r[1]))
		h = loadTaxRow
	case arbitratedProfit:
		i.key = []byte(keyForTaxRegime(r[1]))
		h = loadTaxRow
	case noTaxes:
		i.key = []byte(keyForTaxRegime(r[1]))
		h = loadTaxRow
	default:
		return item{}, fmt.Errorf("unknown source type %s", string(s))
	}
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
			log.Output(1, fmt.Sprintf("Error running garbage collection: %v", err))
			return
		}
	}
}

func (kv *badgerStorage) loadRow(r []string, s sourceType, l *lookups) error {
	i, err := newKVItem(s, l, r)
	if err != nil {
		return fmt.Errorf("error creating an %s item: %w", s, err)
	}
	if err := saveItem(kv.db, i.kind, i.key, i.value); err != nil {
		return fmt.Errorf("could not save key-value: %w", err)
	}
	return nil
}

func (kv *badgerStorage) loadSource(ctx context.Context, s *source, l *lookups, bar *progressbar.ProgressBar) error {
	ch := make(chan []string)
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		defer close(ch)
		return s.sendTo(ctx, ch)
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
				if s.kind.isAccumulative() { // avoid concurrency
					if err := kv.loadRow(r, s.kind, l); err != nil {
						return err
					}
					bar.Add(1)
				} else {
					g.Go(func() error {
						if err := kv.loadRow(r, s.kind, l); err != nil {
							return err
						}
						bar.Add(1)
						return nil
					})
				}
			}
		}
	})
	return g.Wait()
}

func (kv *badgerStorage) load(dir string, l *lookups) error {
	srcs, err := newSources(dir, []sourceType{
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
	bar := progressbar.Default(0, "Processing base CNPJ, partners and taxes")
	defer bar.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)
	for _, src := range srcs {
		bar.AddMax(src.total)
		g.Go(func() error {
			return kv.loadSource(ctx, src, l, bar)
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
	if os.Getenv("DEBUG") != "" {
		log.Output(1, fmt.Sprintf("Creating temporary key-value storage at %s", dir))
	} else {
		opt = opt.WithLogger(&noLogger{})
	}
	db, err := badger.Open(opt)
	if err != nil {
		return nil, fmt.Errorf("error creating badger key-value object: %w", err)
	}
	return &badgerStorage{db: db, path: dir}, nil
}
