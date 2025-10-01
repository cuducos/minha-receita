package db

import (
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
)

const (
	defaultLimit = 256
	maxLimit     = 1024
)

func isValid(p string) bool {
	if p == "" {
		return false
	}
	for _, c := range p {
		if (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '*' {
			return false
		}
	}
	return true
}

func parseURLParams(q []string) []string {
	var r []string
	for _, v := range q {
		for s := range strings.SplitSeq(v, ",") {
			s = strings.ToUpper(strings.TrimSpace(s))
			if isValid(s) {
				r = append(r, s)
			}
		}
	}
	return r
}

func parseURLParamsToUInt(q []string) []uint32 {
	var r []uint32
	for _, v := range parseURLParams(q) {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			slog.Info("Ignoring invalid CNAE number", "cnae", v)
			continue
		}
		r = append(r, uint32(n))
	}
	return r
}

type Query struct {
	CNAE             []uint32
	CNAEFiscal       []uint32
	CNPF             []string // CNPJ or CPF in the QSA
	Municipio        []uint32 // IBGE or SIAFI
	NaturezaJuridica []uint32
	UF               []string
	Cursor           *string
	Limit            uint32
}

func (q *Query) empty() bool {
	return len(q.CNAE) == 0 &&
		len(q.CNAEFiscal) == 0 &&
		len(q.CNPF) == 0 &&
		len(q.Municipio) == 0 &&
		len(q.NaturezaJuridica) == 0 &&
		len(q.UF) == 0
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
		UF:               parseURLParams(v["uf"]),
		Municipio:        parseURLParamsToUInt(v["municipio"]),
		CNPF:             parseURLParams(v["cnpf"]),
		CNAE:             parseURLParamsToUInt(v["cnae"]),
		CNAEFiscal:       parseURLParamsToUInt(v["cnae_fiscal"]),
		NaturezaJuridica: parseURLParamsToUInt(v["natureza_juridica"]),
		Limit:            defaultLimit,
		Cursor:           nil,
	}
	if q.empty() {
		return nil
	}
	for _, v := range parseURLParamsToUInt(v["limit"]) {
		if v > maxLimit {
			continue
		}
		q.Limit = v
		break
	}
	if c := v.Get("cursor"); c != "" {
		q.Cursor = &c

	}
	return &q
}

// builds a paginated search JSON response without depending on marshalling and
// unmarhsalling results from the database (the assumption for performance is
// that data coming from the database is valid JSON text).
func newPage(d []string, c string) string {
	ps := []string{fmt.Sprintf(`"data":[%s]`, strings.Join(d, ","))}
	if c != "" {
		ps = append(ps, fmt.Sprintf(`"cursor":"%s"`, c))
	} else {
		ps = append(ps, `"cursor":null`)
	}
	return fmt.Sprintf(`{%s}`, strings.Join(ps, ","))
}
