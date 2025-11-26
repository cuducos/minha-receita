package transformnext

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"golang.org/x/sync/errgroup"
)

func stringsFromKV(srcs map[string]*source, kv *kv, prefix string, id string) ([]string, error) {
	src, ok := srcs[prefix]
	if !ok {
		return nil, fmt.Errorf("could not find lookup %s", prefix)
	}
	k := src.keyFor(id)
	v, err := kv.get(k)
	if err != nil {
		return nil, fmt.Errorf("could not find %s", string(k))
	}
	return v, nil
}

func stringFromKV(srcs map[string]*source, kv *kv, prefix string, id string, idx uint) (*string, error) {
	v, err := stringsFromKV(srcs, kv, prefix, id)
	if err != nil {
		return nil, err
	}
	if len(v) <= int(idx) {
		return nil, fmt.Errorf("value for %s %s has %d items, cannot load index %d: %v", id, prefix, len(v), idx, v)
	}
	return &v[idx], nil
}

func (c *Company) base(srcs map[string]*source, kv *kv) error {
	var err error
	row, err := stringsFromKV(srcs, kv, "emp", c.CNPJ[:8])
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}
		return err
	}
	if len(row) != 6 {
		return fmt.Errorf("expected exactly 6 columns for base data, got %d", len(row))
	}
	c.RazaoSocial = row[0]
	c.CodigoNaturezaJuridica, err = toInt(row[1])
	if err != nil {
		return fmt.Errorf("could not parse CodigoNaturezaJuridica for %s: %w", c.CNPJ, err)
	}
	c.NaturezaJuridica, err = stringFromKV(srcs, kv, "nat", row[1], 0)
	if err != nil {
		return fmt.Errorf("could not parse NaturezaJuridica for %s: %w", c.CNPJ, err)
	}
	c.QualificacaoDoResponsavel, err = toInt(row[2])
	if err != nil {
		return fmt.Errorf("could not parse QualificacaoDoResponsavel for %s: %w", c.CNPJ, err)
	}
	c.CapitalSocial, err = toFloat(row[3])
	if err != nil {
		return fmt.Errorf("could not parse CapitalSocial for %s: %w", c.CNPJ, err)
	}
	c.CodigoPorte, err = toInt(row[4])
	if err != nil {
		return fmt.Errorf("could not parse CodigoParse for %s: %w", c.CNPJ, err)
	}
	if err := c.porte(); err != nil {
		return err
	}
	c.EnteFederativoResponsavel = row[5]
	return nil
}

func (c *Company) simples(srcs map[string]*source, kv *kv) error {
	var err error
	row, err := stringsFromKV(srcs, kv, "sim", c.CNPJ[:8])
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}
		return err
	}
	if len(row) != 4 {
		return fmt.Errorf("expected exactly 4 columns for simples data, got %d", len(row))
	}
	c.OpcaoPeloSimples = toBool(row[0])
	c.DataOpcaoPeloSimples, err = toDate(row[1])
	if err != nil {
		return fmt.Errorf("could not parse DataOpcaoPeloSimples for %s: %w", c.CNPJ, err)
	}
	c.DataExclusaoDoSimples, err = toDate(row[2])
	if err != nil {
		return fmt.Errorf("could not parse DataExclusaoDoSimples for %s: %w", c.CNPJ, err)
	}
	c.OpcaoPeloMEI = toBool(row[3])
	c.DataOpcaoPeloMEI, err = toDate(row[4])
	if err != nil {
		return fmt.Errorf("could not parse DataOpcaoPeloMEI for %s: %w", c.CNPJ, err)
	}
	c.DataExclusaoDoMEI, err = toDate(row[5])
	if err != nil {
		return fmt.Errorf("could not parse DataExclusaoDoMEI for %s: %w", c.CNPJ, err)
	}
	return nil
}

func (c *Company) cnaes(srcs map[string]*source, kv *kv, codes string) error {
	ch := make(chan CNAE)
	done := make(chan struct{}, 1)
	var g errgroup.Group
	for code := range strings.SplitSeq(codes, ",") {
		g.Go(func() error {
			d, err := stringFromKV(srcs, kv, "cna", code, 0)
			if err != nil {
				return err
			}
			n, err := toInt(*d)
			if err != nil {
				return fmt.Errorf("could not parse CNAESecundarios for %s: %w", c.CNPJ, err)
			}
			ch <- CNAE{*n, *d}
			return nil
		})
	}
	go func() {
		for p := range ch {
			c.CNAESecundarios = append(c.CNAESecundarios, p)
		}
		done <- struct{}{}
	}()
	err := g.Wait()
	close(ch)
	<-done
	return err
}

func (c *Company) partners(srcs map[string]*source, kv *kv) error {
	src, ok := srcs["soc"]
	if !ok {
		return errors.New("could not find lookup soc")
	}
	k := src.keyFor(c.CNPJ[:8])
	rows, err := kv.getPrefix(k)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}
		return fmt.Errorf("could not find %s", string(k))
	}
	ch := make(chan Partner)
	done := make(chan struct{}, 1)
	var g errgroup.Group
	for _, row := range rows {
		g.Go(func() error {
			var p Partner
			var err error
			p.IdentificadorDeSocio, err = toInt(row[0])
			if err != nil {
				return fmt.Errorf("could not parse IdentificadorDeSocio for %s: %w", c.CNPJ, err)
			}
			p.NomeSocio = row[1]
			p.CNPJCPFDoSocio = row[2]
			p.CodigoQualificacaoSocio, err = toInt(row[3])
			if err != nil {
				return fmt.Errorf("could not parse CodigoQualificacaoSocio for %s: %w", c.CNPJ, err)
			}
			p.QualificaoSocio, err = stringFromKV(srcs, kv, "qua", row[3], 0)
			if err != nil {
				return fmt.Errorf("could not parse QualificaoSocio for %s: %w", c.CNPJ, err)
			}
			p.DataEntradaSociedade, err = toDate(row[4])
			if err != nil {
				return fmt.Errorf("could not parse DataEntradaSociedade for %s: %w", c.CNPJ, err)
			}
			p.CodigoPais, err = toInt(row[5])
			if err != nil {
				return fmt.Errorf("could not parse CodigoPais for %s: %w", c.CNPJ, err)
			}
			p.Pais, err = stringFromKV(srcs, kv, "pai", row[5], 0)
			if err != nil {
				return fmt.Errorf("could not parse Pais for %s: %w", c.CNPJ, err)
			}
			p.CPFRepresentanteLegal = row[6]
			p.NomeRepresentanteLegal = row[7]
			p.CodigoQualificacaoRepresentanteLegal, err = toInt(row[8])
			if err != nil {
				return fmt.Errorf("could not parse CodigoQualificacaoRepresentanteLegal for %s: %w", c.CNPJ, err)
			}
			p.QualificacaoRepresentanteLegal, err = stringFromKV(srcs, kv, "qua", row[8], 0)
			if err != nil {
				return fmt.Errorf("could not parse QualificacaoRepresentanteLegal for %s: %w", c.CNPJ, err)
			}
			p.CodigoFaixaEtaria, err = toInt(row[9])
			if err != nil {
				return fmt.Errorf("could not parse CodigoFaixaEtaria for %s: %w", c.CNPJ, err)
			}
			if p.CodigoFaixaEtaria != nil {
				var f string
				switch *p.CodigoFaixaEtaria {
				case 1:
					f = "Entre 0 a 12 anos"
				case 2:
					f = "Entre 13 a 20 ano"
				case 3:
					f = "Entre 21 a 30 anos"
				case 4:
					f = "Entre 31 a 40 anos"
				case 5:
					f = "Entre 41 a 50 anos"
				case 6:
					f = "Entre 51 a 60 anos"
				case 7:
					f = "Entre 61 a 70 anos"
				case 8:
					f = "Entre 71 a 80 anos"
				case 9:
					f = "Maiores de 80 anos"
				case 0:
					f = "Não se aplica"
				default:
					return fmt.Errorf("unknown CodigoFaixaEtaria for %s: %d", c.CNPJ, *p.CodigoFaixaEtaria)
				}
				p.FaixaEtaria = &f
			}
			ch <- p
			return nil
		})
	}
	go func() {
		for p := range ch {
			c.QuadroSocietario = append(c.QuadroSocietario, p)
		}
		done <- struct{}{}
	}()
	err = g.Wait()
	close(ch)
	<-done
	return err
}

func (c *Company) taxes(srcs map[string]*source, kv *kv) error {
	var g errgroup.Group
	ch := make(chan TaxRegime)
	done := make(chan struct{}, 1)
	for _, p := range []string{"arb", "imu", "pre", "rea"} {
		g.Go(func() error {
			src, ok := srcs[p]
			if !ok {
				return fmt.Errorf("could not find lookup %s", p)
			}
			k := src.keyFor(c.CNPJ[:8])
			rows, err := kv.getPrefix(k)
			if err != nil {
				if errors.Is(err, badger.ErrKeyNotFound) {
					return nil
				}
				return fmt.Errorf("could not find %s", string(k))
			}
			for _, row := range rows {
				var t TaxRegime
				y, err := toInt(row[0])
				if err != nil {
					return fmt.Errorf("could not parse Ano for %s: %w", string(k), err)
				}
				t.Ano = *y
				t.CNPJDaSCP = &row[1]
				t.FormaDeTributação = row[2]
				q, err := toInt(row[3])
				if err != nil {
					return fmt.Errorf("could not parse QuantidadeDeEscrituracoes for %s: %w", string(k), err)
				}
				t.QuantidadeDeEscrituracoes = *q
				ch <- t
			}
			return nil
		})
	}
	go func() {
		for t := range ch {
			c.RegimeTributario = append(c.RegimeTributario, t)
		}
		sort.Slice(c.RegimeTributario, func(i, j int) bool {
			return c.RegimeTributario[i].Ano < c.RegimeTributario[j].Ano
		})
		done <- struct{}{}
	}()
	err := g.Wait()
	close(ch)
	<-done
	return err
}

func (c *Company) descricaoMatrizFilial() error {
	if c.IdentificadorMatrizFilial == nil {
		return fmt.Errorf("company %s missing IdentificadorMatrizFilial", c.CNPJ)
	}
	var d string
	switch *c.IdentificadorMatrizFilial {
	case 1:
		d = "MATRIZ"
	case 2:
		d = "FILIAL"
	default:
		return fmt.Errorf("unknown IdentificadorMatrizFilial for %s: %d", c.CNPJ, *c.IdentificadorMatrizFilial)
	}
	c.DescricaoMatrizFilial = &d
	return nil
}

func (c *Company) descricaoSituacaoCadastral() error {
	if c.SituacaoCadastral == nil {
		return fmt.Errorf("company %s missing SituacaoCadastral", c.CNPJ)
	}
	var d string
	switch *c.SituacaoCadastral {
	case 1:
		d = "NULA"
	case 2:
		d = "ATIVA"
	case 3:
		d = "SUSPENSA"
	case 4:
		d = "INAPTA"
	case 8:
		d = "BAIXADA"
	default:
		return fmt.Errorf("unknown IdentificadorMatrizFilial for %s: %d", c.CNPJ, *c.IdentificadorMatrizFilial)
	}
	c.DescricaoSituacaoCadastral = &d
	return nil
}

func (c *Company) porte() error {
	if c.CodigoPorte == nil {
		return fmt.Errorf("company %s missing CodigoPorte", c.CNPJ)
	}
	var p string
	switch *c.CodigoPorte {
	case 0:
		p = "NÃO INFORMADO"
	case 1:
		p = "MICRO EMPRESA"
	case 3:
		p = "EMPRESA DE PEQUENO PORTE"
	case 5:
		p = "DEMAIS"
	default:
		return fmt.Errorf("unknown CodigoPorte for %s: %d", c.CNPJ, c.CodigoPorte)
	}
	c.Porte = &p
	return nil
}
