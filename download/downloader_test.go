package download

import (
	"os"
	"testing"
)

func TestDownloader(t *testing.T) {
	ts := httpTestServer(t, "dados-publicos-cnpj.html")
	defer ts.Close()

	f, s := loadFixture(t, "dados-publicos-cnpj.html")
	defer f.Close()

	tmp := t.TempDir()
	fs := []file{
		newFile(ts.URL+"/file1.html", tmp),
		newFile(ts.URL+"/file2.html", tmp),
	}
	for i := range fs {
		fs[i].size = 203867
	}

	err := download(ts.Client(), fs, DefaultMaxParallel, DefaultMaxRetries, DefaultChunkSize)
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
