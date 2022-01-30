package transform

import (
	"fmt"
	"strings"
)

type lookup map[int]string

func newLookup(p string) (lookup, error) {
	z, err := newArchivedCSV(p, separator)
	if err != nil {
		return nil, fmt.Errorf("error creating archivedCSV for %s: %w", p, err)
	}
	defer z.close()
	l, err := z.toLookup()
	if err != nil {
		return nil, fmt.Errorf("error creating lookup table from %s: %w", p, err)
	}
	return l, nil
}

type lookups struct {
	motives        lookup
	cities         lookup
	countries      lookup
	cnaes          lookup
	qualifications lookup
	natures        lookup
}

func newLookups(d string) (lookups, error) {
	var ls []lookup
	srcs := []sourceType{motives, cities, countries, cnaes, qualifications, natures}
	for _, src := range srcs {
		paths, err := PathsForSource(src, d)
		if err != nil {
			return lookups{}, fmt.Errorf("error finding sources for %s: %w", string(src), err)
		}
		for _, p := range paths {
			l, err := newLookup(p)
			if err != nil {
				return lookups{}, err
			}
			ls = append(ls, l)
		}
	}
	if len(ls) != len(srcs) {
		return lookups{}, fmt.Errorf("error creating look up tables, expected %d items, got %d", len(srcs), len(ls))
	}
	return lookups{ls[0], ls[1], ls[2], ls[3], ls[4], ls[5]}, nil
}

func (c *company) motivoSituacaoCadastral(l *lookups, v string) error {
	i, err := toInt(v)
	if err != nil {
		return fmt.Errorf("error trying to parse MotivoSituacaoCadastral %s: %w", v, err)
	}
	if i == nil {
		return nil
	}
	s := l.motives[*i]
	c.MotivoSituacaoCadastral = i
	if s != "" {
		c.DescricaoMotivoSituacaoCadastral = &s
	}
	return nil
}

func (c *company) pais(l *lookups, v string) error {
	i, err := toInt(v)
	if err != nil {
		return fmt.Errorf("error trying to parse CodigoPais %s: %w", v, err)
	}
	if i == nil {
		return nil
	}
	s := l.countries[*i]
	c.CodigoPais = i
	if s != "" {
		c.Pais = &s
	}
	return nil
}

func (c *company) municipio(l *lookups, v string) error {
	i, err := toInt(v)
	if err != nil {
		return fmt.Errorf("error trying to parse CodigoMunicipio %s: %w", v, err)
	}
	if i == nil {
		return nil
	}
	s := l.cities[*i]
	c.CodigoMunicipio = i
	if s != "" {
		c.Municipio = &s
	}
	return nil
}

type cnae struct {
	Codigo    int    `json:"codigo"`
	Descricao string `json:"descricao"`
}

func newCnae(l *lookups, v string) (cnae, error) {
	i, err := toInt(v)
	if err != nil {
		return cnae{}, fmt.Errorf("error trying to parse cnae %s: %w", v, err)
	}
	if i == nil {
		return cnae{}, nil
	}
	s := l.cnaes[*i]
	return cnae{Codigo: *i, Descricao: s}, nil
}

func (c *company) cnaes(l *lookups, p, s string) error {
	a, err := newCnae(l, p)
	if err != nil {
		return fmt.Errorf("error trying to parse CNAEFiscal %s: %w", p, err)
	}
	c.CNAEFiscal = &a.Codigo
	if a.Descricao != "" {
		c.CNAEFiscalDescricao = &a.Descricao
	}

	for _, n := range strings.Split(s, ",") {
		a, err := newCnae(l, n)
		if err != nil {
			return fmt.Errorf("error trying to parse CNAESecundarios %s: %w", n, err)
		}
		c.CNAESecundarios = append(c.CNAESecundarios, a)
	}
	return nil
}

func (p *partner) qualificacaoSocio(l *lookups, q, r string) error {
	i, err := toInt(q)
	if err != nil {
		return fmt.Errorf("error trying to parse CodigoQualificacaoSocio %s: %w", q, err)
	}
	j, err := toInt(r)
	if err != nil {
		return fmt.Errorf("error trying to parse CodigoQualificacaoRepresentanteLegal %s: %w", r, err)
	}
	if i == nil && j == nil {
		return nil
	}

	s := l.qualifications[*i]
	p.CodigoQualificacaoSocio = i
	if s != "" {
		p.QualificaoSocio = &s
	}

	t := l.qualifications[*j]
	p.CodigoQualificacaoRepresentanteLegal = j
	if t != "" {
		p.QualificacaoRepresentanteLegal = &s
	}
	return nil
}
