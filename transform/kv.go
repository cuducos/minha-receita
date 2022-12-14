package transform

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/cuducos/go-cnpj"
	"github.com/dgraph-io/badger/v3"
	"github.com/schollz/progressbar/v3"
)

type rowToBytesHandler func(l *lookups, r []string) ([]byte, error)

func loadPartnersRow(l *lookups, r []string) ([]byte, error) {
	p, err := newPartner(l, r)
	if err != nil {
		return nil, fmt.Errorf("error parsing taxes line: %w", err)
	}
	v, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("error while marshaling base: %w", err)
	}
	return v, nil
}

func loadBaseRow(l *lookups, r []string) ([]byte, error) {
	b, err := newBaseData(l, r)
	if err != nil {
		return nil, fmt.Errorf("error parsing base line: %w", err)
	}
	v, err := json.Marshal(b)
	if err != nil {
		return nil, fmt.Errorf("error while marshaling base: %w", err)
	}
	return v, nil
}

func loadTaxesRow(_ *lookups, r []string) ([]byte, error) {
	t, err := newTaxesData(r)
	if err != nil {
		return nil, fmt.Errorf("error parsing taxes line: %w", err)
	}
	v, err := json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("error while marshaling base: %w", err)
	}
	return v, nil
}

type item struct {
	key, value []byte
	kind       sourceType
}

func newKVItem(s sourceType, l *lookups, r []string) (i item, err error) {
	var h rowToBytesHandler
	switch s {
	case partners:
		i.key = []byte(keyForPartners(r[0]))
		h = loadPartnersRow
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

func mergePartners(kv *badgerStorage, k, b []byte) ([]byte, error) {
	curr := []byte("[]")
	err := kv.db.View(func(tx *badger.Txn) error {
		i, err := tx.Get(k)
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error getting partner key: %w", err)
		}
		curr, err = i.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("error reading partner value: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error getting current partners: %w", err)
	}
	qsa := []partner{}
	if curr != nil {
		if err := json.Unmarshal(curr, &qsa); err != nil {
			return nil, fmt.Errorf("could not parse partners: %w", err)
		}
	}
	var p partner
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, fmt.Errorf("could not parse partner: %w", err)
	}
	qsa = append(qsa, p)
	j, err := json.Marshal(&qsa)
	if err != nil {
		return nil, fmt.Errorf("could not convert partner to json: %w", err)
	}
	return j, nil
}

func loadKeyValues(kv *badgerStorage, l *lookups, dir string) error {
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
					if i.kind == partners {
						i.value, err = mergePartners(kv, i.key, i.value)
						if err != nil {
							if atomic.CompareAndSwapInt32(&shutdown, 0, 1) {
								errs <- fmt.Errorf("error merging partners: %w", err)
							}
							return
						}
					}
					err = kv.db.Update(func(tx *badger.Txn) error {
						return tx.Set(i.key, i.value)
					})
					if err != nil {
						if atomic.CompareAndSwapInt32(&shutdown, 0, 1) {
							errs <- fmt.Errorf("could save key-value storage: %w", err)
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
	return kv.db.Update(func(tx *badger.Txn) error {
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
	})
}

func enrichCompany(kv *badgerStorage, c *company) error {
	n := cnpj.Base(c.CNPJ)
	ps := make(chan []partner)
	bs := make(chan baseData)
	ts := make(chan taxesData)
	errs := make(chan error)
	go func() {
		p, err := partnersOf(kv, n)
		if err != nil {
			errs <- err
		}
		ps <- p
	}()
	go func() {
		v, err := baseOf(kv, n)
		if err != nil {
			errs <- err
		}
		bs <- v
	}()
	go func() {
		t, err := taxesOf(kv, n)
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
