package db

import (
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cuducos/minha-receita/transform"
)

const (
	timeout      = time.Duration(30 * time.Second)
	defaultLimit = 256
	maxLimit     = 1024
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

func parseURLParamsToUInt(q []string) []uint16 {
	var r []uint16
	for _, v := range parseURLParams(q) {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			slog.Info("Ignoring invalid CNAE number", "cnae", v)
			continue
		}
		r = append(r, uint16(n))
	}
	return r
}

type Query struct {
	UF         []string
	CNAE       []uint16
	CNAEFiscal []uint16
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

func NewQuery(v url.Values) *Query {
	q := Query{
		UF:         parseURLParams(v["uf"]),
		CNAE:       parseURLParamsToUInt(v["cnae"]),
		CNAEFiscal: parseURLParamsToUInt(v["cnae_fiscal"]),
		Limit:      0,
		Cursor:     nil,
	}
	if q.Empty() {
		return nil
	}
	ls := parseURLParamsToUInt(v["limit"])
	if len(ls) == 0 {
		q.Limit = defaultLimit
	} else {
		if ls[0] > maxLimit {
			q.Limit = maxLimit
		} else {
			q.Limit = ls[0]
		}
	}
	if c := v.Get("cursor"); c != "" {
		q.Cursor = &c

	}
	return &q
}

type page struct {
	Data   []transform.Company `json:"data" bson:"data"`
	Cursor *string             `json:"cursor" bson:"cursor"`
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
