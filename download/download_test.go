package download

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"sync/atomic"
	"testing"
)

func TestGetURLs(t *testing.T) {
	for _, tc := range []struct {
		name     string
		fixture  []string
		handler  getURLsHandler
		expected int
	}{
		{"federal revenue", []string{"dados_abertos_cnpj.html", "2024-08.html", "regime_tributario.html"}, federalRevenueGetURLs, 41},

		{"national treasure", []string{"national-treasure.json"}, nationalTreasureGetURLs, 1},
	} {
		ts := httpTestServer(t, tc.fixture)
		defer ts.Close()
		got, err := getURLs(ts.URL, tc.handler, t.TempDir(), true)
		if err != nil {
			t.Errorf("expected to run without errors, got: %v:", err)
			return
		}
		if len(got) != tc.expected {
			t.Errorf("expected %d url(s) from the %s, got %d", tc.expected, tc.name, len(got))
		}
	}
}

func loadFixture(t *testing.T, n string) (*os.File, int64) {
	p := path.Join("..", "testdata", n)
	f, err := os.Open(p)
	if err != nil {
		t.Errorf("Could not open %s: %v", p, err)
		return nil, 0
	}
	i, err := f.Stat()
	if err != nil {
		t.Errorf("Could not get info for %s: %v", p, err)
		return nil, 0
	}
	return f, i.Size()
}

func httpTestServer(t *testing.T, cs []string) *httptest.Server {
	if len(cs) == 0 {
		panic("no content provided to the test server")
	}
	var c uint32
	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			idx := int(atomic.LoadUint32(&c)) % len(cs)
			atomic.AddUint32(&c, 1)
			if r.Method == http.MethodHead {
				f, s := loadFixture(t, cs[idx])
				defer f.Close()
				w.Header().Add("Content-Length", fmt.Sprint(s))
				return
			}
			http.ServeFile(w, r, path.Join("..", "testdata", cs[idx]))
		}))
}

func assertArraysHaveSameItems(t *testing.T, a1, a2 []string) {
	if len(a1) != len(a2) {
		t.Errorf("arrays lengths are different: %d != %d", len(a1), len(a2))
		return
	}

	c1 := make(map[string]int)
	c2 := make(map[string]int)
	for _, v := range a1 {
		c1[v]++
	}
	for _, v := range a2 {
		c2[v]++
	}

	diff := make(map[string]struct{})
	for k := range c1 {
		if c1[k] != c2[k] {
			diff[k] = struct{}{}
		}
	}
	for k := range c2 {
		if c1[k] != c2[k] {
			diff[k] = struct{}{}
		}
	}

	for k := range diff {
		t.Errorf("%q appears %d in the first array, but %d in the second array", k, c1[k], c2[k])
	}
}
