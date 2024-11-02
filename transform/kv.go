package transform

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync/atomic"

	"github.com/cuducos/go-cnpj"
	"github.com/dgraph-io/badger/v4"
	"github.com/schollz/progressbar/v3"
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
	case taxes:
		i.key = []byte(keyForTaxes(r[0]))
		h = loadTaxesRow
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

func (kv *badgerStorage) load(dir string, l *lookups) error {
	srcs, err := newSources(dir, []sourceType{base, partners, taxes})
	if err != nil {
		return fmt.Errorf("could not load sources: %w", err)
	}
	items := make(chan struct{})
	errs := make(chan error)
	var shutdown int32
	defer func() {
		close(items)
		close(errs)
	}()
	var t int
	for _, src := range srcs {
		t += src.totalLines
		for _, a := range src.readers {
			go func(s sourceType, a *archivedCSV) {
				for {
					r, err := a.read()
					if err == io.EOF {
						break
					}
					if err != nil {
						if atomic.CompareAndSwapInt32(&shutdown, 0, 1) {
							errs <- fmt.Errorf("error reading %s: %w", a.path, err)
						}
						return
					}
					i, err := newKVItem(s, l, r)
					if err != nil {
						if atomic.CompareAndSwapInt32(&shutdown, 0, 1) {
							errs <- fmt.Errorf("error creating an %s item: %w", string(s), err)
						}
						return
					}
					if err := saveItem(kv.db, i.kind, i.key, i.value); err != nil {
						if atomic.CompareAndSwapInt32(&shutdown, 0, 1) {
							errs <- fmt.Errorf("could not save key-value: %w", err)
						}
						return
					}
					if atomic.LoadInt32(&shutdown) == 0 {
						items <- struct{}{}
					}
				}
			}(src.kind, a)
		}
	}
	bar := progressbar.Default(int64(t), "Processing base CNPJ, partners and taxes")
	defer bar.Close()
	for {
		select {
		case <-items:
			bar.Add(1)
			if bar.IsFinished() {
				return nil
			}
		case err := <-errs:
			return fmt.Errorf("error creating key-value storage: %w", err)
		}
	}
}

func (kv *badgerStorage) enrichCompany(c *company) error {
	n := cnpj.Base(c.CNPJ)
	ps := make(chan []partnerData)
	bs := make(chan baseData)
	ts := make(chan taxesData)
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
		t, err := taxesOf(kv.db, n)
		if err != nil {
			errs <- err
		}
		ts <- t
	}()
	for i := 0; i < 3; i++ {
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
		case t := <-ts:
			c.OpcaoPeloSimples = t.OpcaoPeloSimples
			c.DataOpcaoPeloSimples = t.DataOpcaoPeloSimples
			c.DataExclusaoDoSimples = t.DataExclusaoDoSimples
			c.OpcaoPeloMEI = t.OpcaoPeloMEI
			c.DataOpcaoPeloMEI = t.DataOpcaoPeloMEI
			c.DataExclusaoDoMEI = t.DataExclusaoDoMEI
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

func (*noLogger) Errorf(string, ...interface{})   {}
func (*noLogger) Warningf(string, ...interface{}) {}
func (*noLogger) Infof(string, ...interface{})    {}
func (*noLogger) Debugf(string, ...interface{})   {}

func newBadgerStorage(dir string) (*badgerStorage, error) {
	opt := badger.DefaultOptions(dir)
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
