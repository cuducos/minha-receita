package download

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
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
}

func TestGetURLs(t *testing.T) {
	ts := httpTestServer(t)
	defer ts.Close()

	got, err := getURLs(ts.Client(), ts.URL)
	if err != nil {
		t.Errorf("Expected getURLs to run withour errors, got: %v:", err)
		return
	}
	assertArraysHaveSameItems(t, got, expectedURLs)
}

func TestGetFiles(t *testing.T) {
	ts := httpTestServer(t)
	defer ts.Close()

	tmp, err := ioutil.TempDir("", "minha-receita")
	if err != nil {
		t.Errorf("Could not create a temporary directory for tests: %v", err)
		return
	}
	defer os.RemoveAll(tmp)

	expected := 38
	got, err := getFiles(ts.Client(), ts.URL, tmp)
	if err != nil {
		t.Errorf("Expected getFiles to run withour errors, got: %v:", err)
		return
	}

	if expected != len(got) {
		t.Errorf("Expected getFiles to return %d files, got %d", expected, len(got))
	}
	for _, f := range got {
		if g := filepath.Dir(f.path); g != tmp {
			t.Errorf("Expected %s parent to be %s, got %s", f.path, tmp, g)
		}
		if g := filepath.Base(f.path); !strings.HasSuffix(f.url, g) {
			t.Errorf("Unexpected file name for %s: %s", f.url, g)
		}
		if !arrayContains(expectedURLs, f.url) && f.url != listOfCNAE {
			t.Errorf("Unexpected URL in getFiles result: %s", f.url)
		}
	}
}

func TestNewDownloader(t *testing.T) {
	ts := httpTestServer(t)
	defer ts.Close()

	tmp, err := ioutil.TempDir("", "minha-receita")
	if err != nil {
		t.Errorf("Could not create a temporary directory for tests: %v", err)
		return
	}
	defer os.RemoveAll(tmp)

	fs := []file{
		{ts.URL + "/file1.html", filepath.Join(tmp, "file1.html")},
		{ts.URL + "/file2.html", filepath.Join(tmp, "file2.html")},
	}
	d, err := newDownloader(ts.Client(), fs)
	if err != nil {
		t.Errorf("Expected newDownloader to return a downloader, got: %v", err)
	}

	f, s := loadFixture(t)
	defer f.Close()

	expectedTotalSize := int64(len(fs)) * s
	if d.totalSize != expectedTotalSize {
		t.Errorf("Expected totalSize to be %d, got %d", expectedTotalSize, d.totalSize)
	}
	if d.bar == nil {
		t.Errorf("Expected downloader to have a progess bar")
	}
}

func TestDownloadAll(t *testing.T) {
	ts := httpTestServer(t)
	defer ts.Close()

	f, s := loadFixture(t)
	defer f.Close()

	tmp, err := ioutil.TempDir("", "minha-receita")
	if err != nil {
		t.Errorf("Could not create a temporary directory for tests: %v", err)
		return
	}
	defer os.RemoveAll(tmp)

	fs := []file{
		{ts.URL + "/file1.html", filepath.Join(tmp, "file1.html")},
		{ts.URL + "/file2.html", filepath.Join(tmp, "file2.html")},
	}
	d, err := newDownloader(ts.Client(), fs)
	if err != nil {
		t.Errorf("Expected newDownloader to return a downloader, got: %v", err)
	}

	err = d.downloadAll()
	if err != nil {
		t.Errorf("Expected downloadAll to run without errors, got: %v", err)
	}

	for _, f := range fs {
		i, err := os.Stat(f.path)
		if err != nil {
			if os.IsNotExist(err) {
				t.Errorf("Expected %s to exist", f.path)
			} else {
				t.Errorf("Error getting info about %s: %v", f.path, err)
			}
			continue
		}
		if i.Size() != s {
			t.Errorf("Expected %s to have length %d, got %d", f.path, s, i.Size())
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

func assertArraysHaveSameItems(t *testing.T, a1, a2 []string) {
	if len(a1) != len(a2) {
		t.Errorf("Arrays lengths are different: %d != %d", len(a1), len(a2))
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

func arrayContains(a []string, v string) bool {
	for _, s := range a {
		if s == v {
			return true
		}
	}
	return false
}

func loadFixture(t *testing.T) (*os.File, int64) {
	p := path.Join("..", "testdata", "dados-publicos-cnpj.html")
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
