package download

import (
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestNewDownloader(t *testing.T) {
	ts := httpTestServer(t)
	defer ts.Close()

	tmp := t.TempDir()
	fs := []file{
		{ts.URL + "/file1.html", filepath.Join(tmp, "file1.html")},
		{ts.URL + "/file2.html", filepath.Join(tmp, "file2.html")},
	}
	d, err := newDownloader(ts.Client(), fs, 4, 4)
	if err != nil {
		t.Errorf("expected newDownloader to return a downloader, got: %s", err)
	}

	f, s := loadFixture(t)
	defer f.Close()

	expectedTotalSize := int64(len(fs)) * s
	if d.totalSize != expectedTotalSize {
		t.Errorf("expected totalSize to be %d, got %d", expectedTotalSize, d.totalSize)
	}
	if d.bar == nil {
		t.Errorf("expected downloader to have a progess bar")
	}
}

func TestDownloadAll(t *testing.T) {
	ts := httpTestServer(t)
	defer ts.Close()

	f, s := loadFixture(t)
	defer f.Close()

	tmp := t.TempDir()
	fs := []file{
		{ts.URL + "/file1.html", filepath.Join(tmp, "file1.html")},
		{ts.URL + "/file2.html", filepath.Join(tmp, "file2.html")},
	}
	d, err := newDownloader(ts.Client(), fs, 4, 4)
	if err != nil {
		t.Errorf("expected newDownloader to return a downloader, got: %s", err)
	}

	err = d.downloadAll()
	if err != nil {
		t.Errorf("expected downloadAll to run without errors, got: %s", err)
	}

	for _, f := range fs {
		i, err := os.Stat(f.path)
		if err != nil {
			if os.IsNotExist(err) {
				t.Errorf("expected %s to exist", f.path)
			} else {
				t.Errorf("error getting info about %s: %s", f.path, err)
			}
			continue
		}
		if i.Size() != s {
			t.Errorf("expected %s to have length %d, got %d", f.path, s, i.Size())
		}
	}
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

func loadFixture(t *testing.T) (*os.File, int64) {
	p := path.Join("..", "testdata", "dados-publicos-cnpj.html")
	f, err := os.Open(p)
	if err != nil {
		t.Errorf("could not open %s: %s", p, err)
		return nil, 0
	}
	i, err := f.Stat()
	if err != nil {
		t.Errorf("could not get info for %s: %s", p, err)
		return nil, 0
	}
	return f, i.Size()
}
