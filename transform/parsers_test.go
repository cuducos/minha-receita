package transform

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fixtures struct {
	company string
	partner string
	cnae    string
}

func newFixture(t *testing.T) fixtures {
	f, err := filepath.Abs(filepath.Join("..", "testdata", "fixed-width-sample"))
	if err != nil {
		t.Errorf("Could understand path %s", f)
	}
	b, err := ioutil.ReadFile(f)
	if err != nil {
		t.Errorf("Could not read from %s", f)
	}
	ls := strings.Split(string(b), "\n")
	return fixtures{company: ls[0], partner: ls[1], cnae: ls[2]}
}

func arraysEqual(t *testing.T, a, b []string) {
	if len(a) != len(b) {
		t.Errorf("1st array has %d items, 2nd array has %d items", len(a), len(b))
	}
	for i := 0; i < len(b); i++ {
		if a[i] != b[i] {
			t.Errorf("Item #%d in the 1st array is %s, but it is %s in the 2nd", i, a[i], b[i])
		}
	}

}
func TestIsNumeric(t *testing.T) {
	cases := []struct {
		value   string
		expects bool
	}{
		{"123", true},
		{"a2c", false},
	}
	for _, c := range cases {
		if r := isNumeric(c.value); r != c.expects {
			t.Errorf("Expected isNumeric to be %v for %s, got %v", c.expects, c.value, r)
		}
	}
}
func TestCleanCompanyName(t *testing.T) {
	cases := []struct {
		value   string
		expects string
	}{
		{"12345678901", "12345678901"},
		{"PESSOA LTDA", "PESSOA LTDA"},
		{"PESSOA - CPF 123", "PESSOA - CPF 123"},
		{"PESSOA - CPF 12345678901", "PESSOA"},
		{"PESSOA 12345678901", "PESSOA"},
		{"PESSOA -", "PESSOA"},
		{"PESSOA - 12345678901", "PESSOA"},
		{"PESSOA CPF 12345678901", "PESSOA"},
		{"PESSOA - CPF", "PESSOA"},
	}
	for _, c := range cases {
		if r := cleanCompanyName(c.value); r != c.expects {
			t.Errorf("Expected cleaned %s to be %s, got %s", c.value, c.expects, r)
		}
	}
}

func TestCleanLine(t *testing.T) {
	c := cleanLine([]string{" foo", "bar  ", "   foo bar\t"})
	e := []string{"foo", "bar", "foo bar"}
	arraysEqual(t, e, c)
}

func TestParsePartner(t *testing.T) {
	f := newFixture(t)
	p := parsePartner(f.partner)
	e := []string{
		"19131243000197",
		"2",
		"FERNANDA CAMPAGNUCCI PEREIRA",
		"***690948**",
		"16",
		"",
		"2019-10-25",
		"",
		"",
		"",
	}
	arraysEqual(t, e, p)
}

func TestParseCNAE(t *testing.T) {
	f := newFixture(t)
	c := parseCNAE(f.cnae)
	e := [][]string{
		{"19131243000197", "9493600"},
		{"19131243000197", "9499500"},
		{"19131243000197", "8599699"},
		{"19131243000197", "8230001"},
		{"19131243000197", "6204000"},
	}
	if len(c) != len(e) {
		t.Errorf("Expected CNAE to have %d items, got %d", len(e), len(c))
	}
	for i := 0; i < len(e); i++ {
		arraysEqual(t, e[i], c[i])
	}
}

func TestParseCompany(t *testing.T) {
	f := newFixture(t)
	c := parseCompany(f.company)
	e := []string{
		"19131243000197",
		"1",
		"OPEN KNOWLEDGE BRASIL",
		"REDE PELO CONHECIMENTO LIVRE",
		"2",
		"2013-10-03",
		"",
		"",
		"3999",
		"2013-10-03",
		"9430800",
		"AVENIDA",
		"PAULISTA 37",
		"37",
		"ANDAR 4",
		"BELA VISTA",
		"01311902",
		"SP",
		"7107",
		"SAO PAULO",
		"11  23851939",
		"",
		"",
		"16",
		"",
		"5",
		"false",
		"",
		"",
		"false",
		"",
		"",
	}
	arraysEqual(t, e, c)
}

func TestParse(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	src, err := ioutil.ReadFile(filepath.Join("..", "testdata", "DADOS_ABERTOS_CNPJ_01.zip"))
	if err != nil {
		t.Error(err)
	}

	err = ioutil.WriteFile(filepath.Join(dir, "DADOS_ABERTOS_CNPJ_01.zip"), src, 0644)
	if err != nil {
		t.Error(err)
	}

	Parse(dir)
	for _, n := range []string{"empresa", "socio", "cnae_secundaria"} {
		f := fmt.Sprintf("%s.csv.gz", n)
		expected, err := os.Stat(filepath.Join("..", "testdata", f))
		if err != nil {
			t.Error(err)
		}

		got, err := os.Stat(filepath.Join(dir, f))
		if err != nil {
			t.Error(err)
		}

		if expected.Size() != got.Size() {
			t.Errorf("Expected generated %s to have %d bytes, but it has %d", f, expected.Size(), got.Size())
		}
	}
}
