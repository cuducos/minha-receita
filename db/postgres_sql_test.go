package db

import (
	"strings"
	"testing"
)

func TestSource(t *testing.T) {
	cases := []struct {
		name    string
		path    string
		table   string
		index   string
		columns int
	}{
		{
			"empresa",
			"data/empresa.csv.gz",
			"empresas",
			"cnpj",
			32,
		},
		{
			"cnae",
			"data/CNAE_Subclasses_2_3_Estrutura_Detalhada.xlsx",
			"cnaes",
			"codigo",
			2,
		},
		{
			"cnae_secundaria",
			"data/cnae_secundaria.csv.gz",
			"cnae_secundarias",
			"cnpj",
			2,
		},
		{
			"socio",
			"data/socio.csv.gz",
			"socios",
			"cnpj",
			10,
		},
	}

	var s source
	for _, c := range cases {
		s = source{c.name}
		if p := s.path("data"); !strings.HasSuffix(p, c.path) {
			t.Errorf("Expected path to be %s, but got %s", c.path, p)
		}
		if tbl := s.tableName(); tbl != c.table {
			t.Errorf("Expected table name to be %s, but got %s", c.table, tbl)
		}
		if i := s.indexName(); i != c.index {
			t.Errorf("Expected index to be %s, but got %s", c.index, i)
		}
		if cols := strings.Split(s.columns(), ","); len(cols) != c.columns {
			t.Errorf("Expected %d columns,  but got %d", c.columns, len(cols))
		}
	}
}

func TestGetSources(t *testing.T) {
	s := getSources()
	if len(s) != 4 {
		t.Errorf("Expected to get 4 fources, but got %d: %v", len(s), s)
	}
}

func TestParseCNAE(t *testing.T) {
	cases := []struct {
		input    []string
		valid    bool
		expected CNAE
	}{
		{
			[]string{"", "", "", "", "4.2", "Quarenta e dois"},
			true,
			CNAE{42, "Quarenta e dois"},
		},
		{
			[]string{"", "", "", "", "", "Nope"},
			false,
			CNAE{},
		},
	}

	for _, c := range cases {
		p, err := parseCNAE(c.input)
		if err == nil {
			if !c.valid {
				t.Errorf("Expected the row %v to be valid, but it fails to parse", c.input)
			}
			if p.Codigo != c.expected.Codigo {
				t.Errorf("Expected the code to be %d, but got %d", c.expected.Codigo, p.Codigo)
			}
			if p.Descricao != c.expected.Descricao {
				t.Errorf("Expected the description to be %s, but got %s", c.expected.Descricao, p.Descricao)
			}
		} else {
			if c.valid {
				t.Errorf("Expected the row %v to be invalid, but it is valid.", c.input)
			}
		}
	}
}
