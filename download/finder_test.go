package download

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

var expectedURLs = []string{
	"http://200.152.38.155/CNPJ/F.K03200$W.SIMPLES.CSV.D10710.zip",
	"http://200.152.38.155/CNPJ/F.K03200$Z.D10710.CNAECSV.zip",
	"http://200.152.38.155/CNPJ/F.K03200$Z.D10710.MOTICSV.zip",
	"http://200.152.38.155/CNPJ/F.K03200$Z.D10710.MUNICCSV.zip",
	"http://200.152.38.155/CNPJ/F.K03200$Z.D10710.NATJUCSV.zip",
	"http://200.152.38.155/CNPJ/F.K03200$Z.D10710.PAISCSV.zip",
	"http://200.152.38.155/CNPJ/F.K03200$Z.D10710.QUALSCSV.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y0.D10710.EMPRECSV.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y0.D10710.ESTABELE.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y0.D10710.SOCIOCSV.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y1.D10710.EMPRECSV.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y1.D10710.ESTABELE.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y1.D10710.SOCIOCSV.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y2.D10710.EMPRECSV.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y2.D10710.ESTABELE.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y2.D10710.SOCIOCSV.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y3.D10710.EMPRECSV.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y3.D10710.ESTABELE.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y3.D10710.SOCIOCSV.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y4.D10710.EMPRECSV.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y4.D10710.ESTABELE.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y4.D10710.SOCIOCSV.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y5.D10710.EMPRECSV.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y5.D10710.ESTABELE.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y5.D10710.SOCIOCSV.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y6.D10710.EMPRECSV.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y6.D10710.ESTABELE.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y6.D10710.SOCIOCSV.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y7.D10710.EMPRECSV.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y7.D10710.ESTABELE.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y7.D10710.SOCIOCSV.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y8.D10710.EMPRECSV.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y8.D10710.ESTABELE.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y8.D10710.SOCIOCSV.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y9.D10710.EMPRECSV.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y9.D10710.ESTABELE.zip",
	"http://200.152.38.155/CNPJ/K3241.K03200Y9.D10710.SOCIOCSV.zip",
	"http://200.152.38.155/CNPJ/anual/Dados%20Abertos%20S%c3%adtio%20RFB%20Extracao%2020.10.2021.zip",
}

func TestGetUpdateDates(t *testing.T) {
	ts := httpTestServer(t)
	defer ts.Close()

	doc, err := getHTMLDocument(ts.Client(), ts.URL)
	if err != nil {
		t.Errorf("expected getHTMLDocument to run without errors, got: %v:", err)
		return
	}

	expected := []string{"2021-01-08", "2021-10-20"}
	got := getLastUpdate(doc)
	if len(expected) != len(got) {
		t.Errorf("expected getLastUpdate to return %d dates, got %d", len(expected), len(got))
	}

	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected getLastUpdate to return %v, got %v", expected, got)
	}
}

func TestGetURLs(t *testing.T) {
	ts := httpTestServer(t)
	defer ts.Close()

	doc, err := getHTMLDocument(ts.Client(), ts.URL)
	if err != nil {
		t.Errorf("expected getHTMLDocument to run without errors, got: %v:", err)
		return
	}

	got, err := getURLs(doc)
	if err != nil {
		t.Errorf("expected getURLs to run without errors, got: %v:", err)
		return
	}
	assertArraysHaveSameItems(t, got, expectedURLs)
}

func TestGetFiles(t *testing.T) {
	ts := httpTestServer(t)
	defer ts.Close()

	doc, err := getHTMLDocument(ts.Client(), ts.URL)
	if err != nil {
		t.Errorf("expected getHTMLDocument to run without errors, got: %v:", err)
		return
	}

	tmp := t.TempDir()
	expected := 38
	got, err := getFiles(doc, tmp, false)
	if err != nil {
		t.Errorf("expected getFiles to run withour errors, got: %v:", err)
		return
	}

	if expected != len(got) {
		t.Errorf("expected getFiles to return %d files, got %d", expected, len(got))
	}
	for _, f := range got {
		if g := filepath.Dir(f.path); g != tmp {
			t.Errorf("expected %s parent to be %s, got %s", f.path, tmp, g)
		}
		if g := filepath.Base(f.path); !strings.HasSuffix(f.url, g) {
			t.Errorf("unexpected file name for %s: %s", f.url, g)
		}
	}
}
func httpTestServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			f, s := loadFixture(t)
			defer f.Close()

			if r.Method == http.MethodHead {
				w.Header().Add("Content-Length", fmt.Sprint(s))
				return
			}
			io.Copy(w, f)
		}))
}
