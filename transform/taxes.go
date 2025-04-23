package transform

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type simpleTaxesData struct {
	OpcaoPeloSimples      *bool `json:"opcao_pelo_simples" bson:"opcao_pelo_simples"`
	DataOpcaoPeloSimples  *date `json:"data_opcao_pelo_simples" bson:"data_opcao_pelo_simples"`
	DataExclusaoDoSimples *date `json:"data_exclusao_do_simples" bson:"data_exclusao_do_simples"`
	OpcaoPeloMEI          *bool `json:"opcao_pelo_mei" bson:"opcao_pelo_mei"`
	DataOpcaoPeloMEI      *date `json:"data_opcao_pelo_mei" bson:"data_opcao_pelo_mei"`
	DataExclusaoDoMEI     *date `json:"data_exclusao_do_mei" bson:"data_exclusao_do_mei"`
}

type TaxRegime struct {
	Ano                       int     `json:"ano" bson:"ano"`
	CNPJDaSCP                 *string `json:"cnpj_da_scp" bson:"cnpj_da_scp"`
	FormaDeTributação         string  `json:"forma_de_tributacao" bson:"forma_de_tributacao"`
	QuantidadeDeEscrituracoes int     `json:"quantidade_de_escrituracoes" bson:"quantidade_de_escrituracoes"`
}

type TaxRegimes []TaxRegime

func (t TaxRegimes) Len() int           { return len(t) }
func (t TaxRegimes) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t TaxRegimes) Less(i, j int) bool { return t[i].Ano < t[j].Ano }

func newSimpleTaxesData(r []string) (simpleTaxesData, error) {
	var err error
	d := simpleTaxesData{
		OpcaoPeloSimples: toBool(r[1]),
		OpcaoPeloMEI:     toBool(r[4]),
	}
	d.DataOpcaoPeloSimples, err = toDate(r[2])
	if err != nil {
		return simpleTaxesData{}, fmt.Errorf("error parsing DataOpcaoPeloSimples %s: %w", r[2], err)
	}
	d.DataExclusaoDoSimples, err = toDate(r[3])
	if err != nil {
		return simpleTaxesData{}, fmt.Errorf("error parsing DataExclusaoDoSimples %s: %w", r[3], err)
	}
	d.DataOpcaoPeloMEI, err = toDate(r[5])
	if err != nil {
		return simpleTaxesData{}, fmt.Errorf("error parsing DataOpcaoPeloMEI %s: %w", r[5], err)
	}
	d.DataExclusaoDoMEI, err = toDate(r[6])
	if err != nil {
		return simpleTaxesData{}, fmt.Errorf("error parsing DataExclusaoDoMEI %s: %w", r[6], err)
	}
	return d, nil
}

func newTaxRegimeData(r []string) (TaxRegime, error) {
	y, err := strconv.Atoi(r[0])
	if err != nil {
		return TaxRegime{}, fmt.Errorf("error reading year %s for tax data: %w", r[0], err)
	}
	q, err := strconv.Atoi(r[4])
	if err != nil {
		return TaxRegime{}, fmt.Errorf("error reading quantity %s for tax data: %w", r[4], err)
	}
	var c *string
	if r[2] != "" && r[2] != "0" {
		c = &r[2]
	}
	return TaxRegime{y, c, r[3], q}, nil
}

func loadSimpleTaxesRow(_ *lookups, r []string) ([]byte, error) {
	t, err := newSimpleTaxesData(r)
	if err != nil {
		return nil, fmt.Errorf("error parsing taxes line: %w", err)
	}
	v, err := json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("error while marshaling base: %w", err)
	}
	return v, nil
}

func loadTaxRow(_ *lookups, r []string) ([]byte, error) {
	t, err := newTaxRegimeData(r)
	if err != nil {
		return nil, fmt.Errorf("error parsing tax line data: %w", err)
	}
	b, err := json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("error while marshaling tax data: %w", err)
	}
	return b, nil
}
