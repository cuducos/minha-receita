package db

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompany(t *testing.T) {
	f, err := filepath.Abs(filepath.Join("..", "testdata", "response.json"))
	if err != nil {
		t.Errorf("Could understand path %s", f)
	}
	b, err := ioutil.ReadFile(f)
	if err != nil {
		t.Errorf("Could not read from %s", f)
	}
	expected := strings.TrimSpace(string(b))

	c := Company{CNPJ: "19131243000197"}
	if j, _ := c.JSON(); j != expected {
		t.Errorf("\nExpected JSON to be:\n\t%s\nGot:\n\t%s", expected, j)
	}
	if s := c.String(); s != "19.131.243/0001-97" {
		t.Errorf("Expected company to be 19.131.243/0001-97, but got %s", s)
	}
}
