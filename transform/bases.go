package transform

import (
	"encoding/json"
	"fmt"
)

type baseData struct {
	CodigoPorte               *int     `json:"codigo_porte"`
	Porte                     *string  `json:"porte"`
	RazaoSocial               string   `json:"razao_social"`
	CodigoNaturezaJuridica    *int     `json:"codigo_natureza_juridica"`
	NaturezaJuridica          *string  `json:"natureza_juridica"`
	QualificacaoDoResponsavel *int     `json:"qualificacao_do_responsavel"`
	CapitalSocial             *float32 `json:"capital_social"`
	EnteFederativoResponsavel string   `json:"ente_federativo_responsavel"`
}

func (d *baseData) porte(v string) error {
	i, err := toInt(v)
	if err != nil {
		return fmt.Errorf("error trying to parse CodigoPorte %s: %w", v, err)
	}
	if i == nil {
		return nil
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

	d.CodigoPorte = i
	if s != "" {
		d.Porte = &s
	}
	return nil
}

func (d *baseData) base(r []string, l *lookups) error {
	d.RazaoSocial = r[1]
	codigoNaturezaJuridica, err := toInt(r[2])
	if err != nil {
		return fmt.Errorf("error trying to parse CodigoNaturezaJuridica %s: %w", r[2], err)
	}
	d.CodigoNaturezaJuridica = codigoNaturezaJuridica
	qualificacaoDoResponsavel, err := toInt(r[3])
	if err != nil {
		return fmt.Errorf("error trying to parse QualificacaoDoResponsavel %s: %w", r[3], err)
	}
	d.QualificacaoDoResponsavel = qualificacaoDoResponsavel
	capitalSocial, err := toFloat(r[4])
	if err != nil {
		return fmt.Errorf("error trying to parse CapitalSocial %s: %w", r[4], err)
	}
	d.CapitalSocial = capitalSocial
	err = d.porte(r[5])
	if err != nil {
		return fmt.Errorf("error trying to parse Porte %s: %w", r[5], err)
	}
	d.EnteFederativoResponsavel = r[6]
	natures := l.natures[*d.CodigoNaturezaJuridica]
	if natures != "" {
		d.NaturezaJuridica = &natures
	}
	return nil
}

func newBaseData(l *lookups, r []string) (baseData, error) {
	var d baseData
	if err := d.base(r, l); err != nil {
		return baseData{}, fmt.Errorf("error handling base data for base cnpj %s: %w", r[0], err)
	}
	return d, nil
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
