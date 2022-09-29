package download

import (
	"os"
	"testing"
)

func TestNewDownloader(t *testing.T) {
	tt := []struct {
		name   string
		silent bool
	}{
		{"verbose", false},
		{"silent", true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ts := httpTestServer(t, "dados-publicos-cnpj.html")
			defer ts.Close()

			tmp := t.TempDir()
			fs := []file{
				newFile(ts.URL+"/file1.html", tmp),
				newFile(ts.URL+"/file2.html", tmp),
			}
			fs, err := getSizes(ts.Client(), fs, tc.silent)
			if err != nil {
				t.Errorf("Expected getSizes to run without errors, got: %s", err)
			}
			d, err := newDownloader(ts.Client(), fs, 4, 4, tc.silent)
			if err != nil {
				t.Errorf("Expected newDownloader to return a downloader, got: %v", err)
			}

			f, s := loadFixture(t, "dados-publicos-cnpj.html")
			defer f.Close()

			expectedTotalSize := int64(len(fs)) * s
			if d.bar.main.GetMax64() != expectedTotalSize {
				t.Errorf("Expected totalSize to be %d, got %d", expectedTotalSize, d.bar.main.GetMax64())
			}
			if d.bar == nil {
				t.Errorf("Expected downloader to have a progess bar")
			}
		})
	}
}

func TestDownloadAll(t *testing.T) {
	tt := []struct {
		name   string
		silent bool
	}{
		{"verbose", false},
		{"silent", true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ts := httpTestServer(t, "dados-publicos-cnpj.html")
			defer ts.Close()

			f, s := loadFixture(t, "dados-publicos-cnpj.html")
			defer f.Close()

			tmp := t.TempDir()
			fs := []file{
				newFile(ts.URL+"/file1.html", tmp),
				newFile(ts.URL+"/file2.html", tmp),
			}
			fs, err := getSizes(ts.Client(), fs, tc.silent)
			if err != nil {
				t.Errorf("Expected getSizes to run without errors, got: %s", err)
			}
			d, err := newDownloader(ts.Client(), fs, 4, 4, tc.silent)
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
		})
	}
}
