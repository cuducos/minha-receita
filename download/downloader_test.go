package download

import (
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestNewDownloader(t *testing.T) {
	ts := httpTestServer(t, "dados-publicos-cnpj.html")
	defer ts.Close()

	tmp := t.TempDir()
	fs := []file{
		{ts.URL + "/file1.html", filepath.Join(tmp, "file1.html")},
		{ts.URL + "/file2.html", filepath.Join(tmp, "file2.html")},
	}
	d, err := newDownloader(ts.Client(), fs, 4, 4)
	if err != nil {
		t.Errorf("Expected newDownloader to return a downloader, got: %v", err)
	}

	f, s := loadFixture(t, "dados-publicos-cnpj.html")
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
	ts := httpTestServer(t, "dados-publicos-cnpj.html")
	defer ts.Close()

	f, s := loadFixture(t, "dados-publicos-cnpj.html")
	defer f.Close()

	tmp := t.TempDir()
	fs := []file{
		{ts.URL + "/file1.html", filepath.Join(tmp, "file1.html")},
		{ts.URL + "/file2.html", filepath.Join(tmp, "file2.html")},
	}
	d, err := newDownloader(ts.Client(), fs, 4, 4)
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
