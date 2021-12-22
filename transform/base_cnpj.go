package transform

import (
	"fmt"
	"io"
	"path/filepath"
)

func addBaseCPNJ(dir string, l *lookups) error {
	s, err := newSource(base_cpnj, dir)
	if err != nil {
		return fmt.Errorf("error creating source for partners: %w", err)
	}
	defer s.close()
	for _, r := range s.readers {
		for {
			t, err := r.read()
			if err == io.EOF {
				break
			}
			if err != nil {
				break // do not proceed in case of errors.
			}
			ls, err := filepath.Glob(filepath.Join(dir, t[0], "*.json"))
			if err != nil {
				return fmt.Errorf("error in the glob pattern: %w", err)
			}
			/*
				if len(ls) == 0 {
					return fmt.Errorf("No JSON file found for CNPJ base %s", t[0])
				}
			*/
			for _, f := range ls {
				c, err := companyFromJSON(f)
				if err != nil {
					return fmt.Errorf("error reading company from %s: %w", f, err)
				}
				err = c.fillMain(t)
				if err != nil {
					return fmt.Errorf("error filling company from %s: %w", f, err)
				}
				f, err = c.toJSON(dir)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (c *company) fillMain(data []string) error {
	c.RazaoSocial = data[1]
	codigoNaturezaJuridica, err := toInt(data[2])
	if err != nil {
		return fmt.Errorf("error trying to parse CodigoNaturezaJuridica %s: %w", data[2], err)
	}
	c.CodigoNaturezaJuridica = codigoNaturezaJuridica
	qualificacaoDoResponsavel, err := toInt(data[3])
	if err != nil {
		return fmt.Errorf("error trying to parse QualificacaoDoResponsavel %s: %w", data[2], err)
	}
	c.QualificacaoDoResponsavel = qualificacaoDoResponsavel
	capitalSocial, err := toFloat(data[4])
	if err != nil {
		return fmt.Errorf("error trying to parse CapitalSocial %s: %w", data[2], err)
	}
	c.CapitalSocial = capitalSocial
	err = c.porte(data[5])
	if err != nil {
		return fmt.Errorf("error trying to parse Porte %s: %w", data[5], err)
	}
	enteFederativoResponsavel, err := toInt(data[6])
	if err != nil {
		return fmt.Errorf("error trying to parse EnteFederativoResponsavel%s: %w", data[2], err)
	}
	c.EnteFederativoResponsavel = enteFederativoResponsavel
	return nil
}

func (c *company) porte(v string) error {
	i, err := toInt(v)
	if err != nil {
		return fmt.Errorf("error trying to parse CodigoPorte %s: %w", v, err)
	}

	var s string
	switch *i {
	case 0:
		s = "NÃO INFORMADO"
	case 1:
		s = "MICRO EMPRESA"
	case 3:
		s = "EMPRESA DE PEQUENO PORTE"
	case 5:
		s = "DEMAIS"
	}

	c.CodigoPorte = i
	if s != "" {
		c.Porte = &s
	}
	return nil
}
