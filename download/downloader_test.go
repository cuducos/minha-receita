package download

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDownloader(t *testing.T) {
	ts := httpTestServer(t, []string{"Empresas1.zip"})
	defer ts.Close()

	f, s := loadFixture(t, "Empresas1.zip")
	defer func() {
		if err := f.Close(); err != nil {
			t.Errorf("expected no error closing fixture, got %s", err)
		}
	}()

	tmp := t.TempDir()
	urls := []string{ts.URL + "/file1.html", ts.URL + "/file2.html"}
	if err := download(tmp, urls, DefaultMaxParallel, DefaultMaxRetries, DefaultChunkSize, 10*time.Second, true); err != nil {
		t.Errorf("Expected downloadAll to run without errors, got: %v", err)
	}
	for _, u := range urls {
		p := filepath.Join(tmp, filepath.Base(u))
		i, err := os.Stat(p)
		if err != nil {
			if os.IsNotExist(err) {
				t.Errorf("Expected %s to exist", p)
			} else {
				t.Errorf("Error getting info about %s: %v", p, err)
			}
			continue
		}
		if i.Size() != s {
			t.Errorf("Expected %s to have length %d, got %d", p, s, i.Size())
		}
	}
}
