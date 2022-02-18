package transform

import "testing"

func TestCitiesLookup(t *testing.T) {
	l, err := citiesLookup(testdata)
	if err != nil {
		t.Errorf("expected no error creating the cities lookup, got %s", err)
	}
	got := l[9701]
	expected := "5300108"
	if got != expected {
		t.Errorf("expected ibge city code to be %s, got %s", expected, got)
	}
}
