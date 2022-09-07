package download

import "testing"

var fs = []file{
	{"http://0.0.0.0:8000/one.zip", "/tmp/one.zip", 1},
	{"http://0.0.0.0:8000/two.zip", "/tmp/two.zip", 2},
}

func TestSimpleURLList(t *testing.T) {
	expected := "http://0.0.0.0:8000/one.zip\nhttp://0.0.0.0:8000/two.zip"
	if got := simpleURLList(fs); got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestTSVURLList(t *testing.T) {
	expected := "TsvHttpData-1.0\nhttp://0.0.0.0:8000/one.zip\t1\nhttp://0.0.0.0:8000/two.zip\t2"
	got, err := tsvURLList(fs)
	if err != nil {
		t.Errorf("expected no errors generating tsv ur list, got %s", err)
	}
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}
