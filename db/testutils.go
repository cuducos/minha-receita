package db

import (
	"encoding/json/v2"
	"reflect"
	"testing"

	"github.com/cuducos/minha-receita/transform"
)

func assertCompaniesAreEqual(t *testing.T, s1 string, s2 string) {
	toCompany := func(s string) transform.Company {
		var c transform.Company
		if err := json.Unmarshal([]byte(s), &c); err != nil {
			t.Errorf("expected no error unmarshalling company, got %s", err)
		}
		return c
	}
	c1 := toCompany(s1)
	c2 := toCompany(s2)
	if !reflect.DeepEqual(c1, c2) {
		t.Errorf("expected companies to be equal, got %s and %s", s1, s2)
	}
}
