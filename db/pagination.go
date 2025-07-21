package db

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/cuducos/minha-receita/transform"
)

func parseURLParams(q []string) []string {
	var r []string
	for _, v := range q {
		for s := range strings.SplitSeq(v, ",") {
			s = strings.ToUpper(strings.TrimSpace(s))
			if s != "" {
				r = append(r, s)
			}
		}
	}
	return r
}

type Query struct {
	UF         []string
	CNAE       []string
	CNAEFiscal []string
	Cursor     *string
	Limit      uint16
}

func (q *Query) Empty() bool {
	return len(q.UF) == 0 && len(q.CNAE) == 0 && len(q.CNAEFiscal) == 0
}
func (q *Query) CursorAsInt() (int, error) {
	if q.Cursor == nil {
		return 0, nil
	}
	c := *q.Cursor
	if c == "" {
		return 0, nil
	}
	return strconv.Atoi(c)
}

func NewQuery(v url.Values, limit uint16) *Query {
	q := Query{
		UF:         parseURLParams(v["uf"]),
		CNAE:       parseURLParams(v["cnae"]),
		CNAEFiscal: parseURLParams(v["cnae_fiscal"]),
		Cursor:     nil,
		Limit:      limit,
	}
	if q.Empty() {
		return nil
	}
	if c := v.Get("cursor"); c != "" {
		q.Cursor = &c
	}
	return &q
}

type page struct {
	Data   []transform.Company `json:"data"`
	Cursor *string             `json:"cursor"`
}

func newPage(cs []transform.Company, c string) page {
	var p page
	if len(cs) > 0 {
		p.Data = cs
	} else {
		p.Data = []transform.Company{}
	}
	if c != "" {
		p.Cursor = &c
	}
	return p
}
