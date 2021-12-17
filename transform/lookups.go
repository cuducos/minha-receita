package transform

import "fmt"

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
	motives   lookup
	cities    lookup
	countries lookup
	cnaes     lookup
}

func newLookups(d string) (lookups, error) {
	var ls []lookup
	srcs := []sourceType{motives, cities, countries, cnaes}
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
	return lookups{ls[0], ls[1], ls[2], ls[3]}, nil
}

func (c *company) motivoSituacaoCadastral(l lookups, v string) error {
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

func (c *company) pais(l lookups, v string) error {
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

func (c *company) municipio(l lookups, v string) error {
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

func (c *company) cnae(l lookups, v string) error {
	cnae, err := newCnae(l, v)
	if err != nil {
		return err
	}
	c.CNAEFiscal = &cnae.Codigo
	if cnae.Descricao != "" {
		c.CNAEFiscalDescricao = &cnae.Descricao
	}
	return nil
}
