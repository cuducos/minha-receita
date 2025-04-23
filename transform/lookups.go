package transform

import (
	"fmt"
	"log"
	"strings"
)

var separator = ';'

type lookup map[int]string

func newLookup(p string) (lookup, error) {
	z, err := newArchivedCSV(p, separator, false)
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
	ibge           lookup
}

func newLookups(d string) (lookups, error) {
	var ls []lookup
	srcs := []sourceType{motives, cities, countries, cnaes, qualifications, natures}
	for _, src := range srcs {
		paths, err := pathsForSource(src, d)
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
	c, err := citiesLookup(d)
	if err != nil {
		return lookups{}, fmt.Errorf("error creating ibge lookup: %w", err)
	}
	return lookups{ls[0], ls[1], ls[2], ls[3], ls[4], ls[5], c}, nil
}

func (c *Company) motivoSituacaoCadastral(l *lookups, v string) error {
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

func (c *Company) pais(l *lookups, v string) error {
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

func (c *Company) municipio(l *lookups, v string) error {
	if c.UF == "EX" {
		return nil
	}
	i, err := toInt(v)
	if err != nil {
		return fmt.Errorf("error trying to parse CodigoMunicipio %s: %w", v, err)
	}
	if i == nil {
		return nil
	}
	c.CodigoMunicipio = i
	s, ok := l.cities[*i]
	if !ok {
		return nil
	}
	c.Municipio = &s
	ibge, ok := l.ibge[*i]
	if !ok {
		log.Output(1, fmt.Sprintf("Could not find IBGE city code for %s-%s (%d)", *c.Municipio, c.UF, *i))
		return nil
	}
	c.CodigoMunicipioIBGE, err = toInt(ibge)
	if err != nil {
		return fmt.Errorf("error trying to parse ibge code %s: %w", ibge, err)
	}
	return nil
}

type CNAE struct {
	Codigo    int    `json:"codigo"`
	Descricao string `json:"descricao"`
}

func newCnae(l *lookups, v string) (CNAE, error) {
	i, err := toInt(v)
	if err != nil {
		return CNAE{}, fmt.Errorf("error trying to parse cnae %s: %w", v, err)
	}
	if i == nil {
		return CNAE{}, nil
	}
	s := l.cnaes[*i]
	return CNAE{Codigo: *i, Descricao: s}, nil
}

func (c *Company) cnaes(l *lookups, p, s string) error {
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

func (p *PartnerData) qualificacaoSocio(l *lookups, q, r string) error {
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
		p.QualificacaoRepresentanteLegal = &t
	}
	return nil
}
