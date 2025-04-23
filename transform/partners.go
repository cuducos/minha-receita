package transform

import (
	"encoding/json"
	"fmt"
)

type PartnerData struct {
	IdentificadorDeSocio                 *int    `json:"identificador_de_socio" bson:"identificador_de_socio,omitempty"`
	NomeSocio                            string  `json:"nome_socio" bson:"nome_socio"`
	CNPJCPFDoSocio                       string  `json:"cnpj_cpf_do_socio" bson:"cnpj_cpf_do_socio"`
	CodigoQualificacaoSocio              *int    `json:"codigo_qualificacao_socio" bson:"codigo_qualificacao_socio,omitempty"`
	QualificaoSocio                      *string `json:"qualificacao_socio" bson:"qualificacao_socio,omitempty"`
	DataEntradaSociedade                 *date   `json:"data_entrada_sociedade" bson:"data_entrada_sociedade,omitempty"`
	CodigoPais                           *int    `json:"codigo_pais" bson:"codigo_pais,omitempty"`
	Pais                                 *string `json:"pais" bson:"pais,omitempty"`
	CPFRepresentanteLegal                string  `json:"cpf_representante_legal" bson:"cpf_representante_legal"`
	NomeRepresentanteLegal               string  `json:"nome_representante_legal" bson:"nome_representante_legal"`
	CodigoQualificacaoRepresentanteLegal *int    `json:"codigo_qualificacao_representante_legal" bson:"codigo_qualificacao_representante_legal,omitempty"`
	QualificacaoRepresentanteLegal       *string `json:"qualificacao_representante_legal" bson:"qualificacao_representante_legal,omitempty"`
	CodigoFaixaEtaria                    *int    `json:"codigo_faixa_etaria" bson:"codigo_faixa_etaria,omitempty"`
	FaixaEtaria                          *string `json:"faixa_etaria" bson:"faixa_etaria,omitempty"`
}


func (p *PartnerData) faixaEtaria(v string) {
	c, err := toInt(v)
	if err != nil {
		return
	}
	p.CodigoFaixaEtaria = c

	var s string
	switch *c {
	case 1:
		s = "para os intervalos entre 0 a 12 anos"
	case 2:
		s = "Entre 13 a 20 ano"
	case 3:
		s = "Entre 21 a 30 anos"
	case 4:
		s = "Entre 31 a 40 anos"
	case 5:
		s = "Entre 41 a 50 anos"
	case 6:
		s = "Entre 51 a 60 anos"
	case 7:
		s = "Entre 61 a 70 anos"
	case 8:
		s = "Entre 71 a 80 anos"
	case 9:
		s = "Maiores de 80 anos"
	case 0:
		s = "Não se aplica"
	}
	if s != "" {
		p.FaixaEtaria = &s
	}
}

func (p *PartnerData) pais(l *lookups, v string) error {
	i, err := toInt(v)
	if err != nil {
		return fmt.Errorf("error trying to parse CodigoPais %s: %w", v, err)
	}
	if i == nil {
		return nil
	}
	s := l.countries[*i]
	p.CodigoPais = i
	if s != "" {
		p.Pais = &s
	}
	return nil
}

func newPartnerData(l *lookups, r []string) (PartnerData, error) {
	identificadorDeSocio, err := toInt(r[1])
	if err != nil {
		return PartnerData{}, fmt.Errorf("error parsing IdentificadorDeSocio %s: %w", r[1], err)
	}

	dataEntradaSociedade, err := toDate(r[5])
	if err != nil {
		return PartnerData{}, fmt.Errorf("error parsing DataEntradaSociedade %s: %w", r[5], err)
	}

	p := PartnerData{
		IdentificadorDeSocio:   identificadorDeSocio,
		NomeSocio:              r[2],
		CNPJCPFDoSocio:         r[3],
		DataEntradaSociedade:   dataEntradaSociedade,
		CPFRepresentanteLegal:  r[7],
		NomeRepresentanteLegal: r[8],
	}
	p.pais(l, r[6])
	p.faixaEtaria(r[10])
	p.qualificacaoSocio(l, r[4], r[9])
	return p, nil
}

func loadPartnerRow(l *lookups, r []string) ([]byte, error) {
	p, err := newPartnerData(l, r)
	if err != nil {
		return nil, fmt.Errorf("error parsing taxes line: %w", err)
	}
	v, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("error while marshaling base: %w", err)
	}
	return v, nil
}
