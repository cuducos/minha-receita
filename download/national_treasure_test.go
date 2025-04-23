package download

import (
	"testing"

	"github.com/cuducos/minha-receita/testutils"
)

func TestNationalTreasureGetURLs(t *testing.T) {
	ts := httpTestServer(t, []string{"national-treasure.json"})
	defer ts.Close()
	got, err := nationalTreasureGetURLs(ts.URL)
	if err != nil {
		t.Errorf("expected to run without errors, got: %v:", err)
		return
	}
	expected := []string{
		"https://www.tesourotransparente.gov.br/ckan/dataset/abb968cb-3710-4f85-89cf-875c91b9c7f6/resource/eebb3bc6-9eea-4496-8bcf-304f33155282/download/TABMUN.CSV",
	}
	testutils.AssertArraysHaveSameItems(t, got, expected)
}
