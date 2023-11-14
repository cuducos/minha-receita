package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cuducos/chunk"
	"github.com/schollz/progressbar/v3"
)

const (
	// DefaultChunkSize sets the size of the chunks to be downloaded using HTTP
	// requests by bytes range
	DefaultChunkSize = 1_048_576

	// DefaultMaxRetries sets the maximum download attempt for each chunk
	DefaultMaxRetries = uint(32)

	// DefaultMaxParallel sets the maximum parallels downloads per server
	DefaultMaxParallel = 16

	// DefaultTimeout sets the timeout for each HTTP request
	DefaultTimeout = 3 * time.Minute
)

type bar struct {
	main       *progressbar.ProgressBar
	urls       map[string]int64
	totalFiles int
	filesDone  int
	totalBytes int64
}

func (b *bar) label() string {
	if len(b.urls) < b.totalFiles {
		return fmt.Sprintf("Gathering file sizes (%d of %d)", len(b.urls), b.totalFiles)
	}
	return fmt.Sprintf("Downloading (%d of %d files done)", b.filesDone, b.totalFiles)
}

func (b *bar) downloadedBytes() int64 {
	var t int64
	for _, b := range b.urls {
		t += b
	}
	return t
}

func (b *bar) update(s chunk.DownloadStatus) {
	_, exists := b.urls[s.URL]
	if !exists {
		b.urls[s.URL] = 0
		b.totalBytes += s.FileSizeBytes
	}
	b.urls[s.URL] = s.DownloadedFileBytes
	if s.IsFinished() {
		b.filesDone += 1
	}
	if b.main == nil {
		b.main = progressbar.DefaultBytes(b.totalBytes, b.label())
	} else {
		b.main.ChangeMax64(b.totalBytes)
		b.main.Describe(b.label())
	}
	b.main.Set64(b.downloadedBytes())
}

func download(dir string, urls []string, parallel int, retries uint, chunkSize int64, timeout time.Duration, restart bool) error {
	d := chunk.DefaultDownloader()
	d.OutputDir = dir
	d.ConcurrencyPerServer = parallel
	d.Timeout = timeout
	d.MaxRetries = retries
	d.ChunkSize = chunkSize
	d.RestartDownloads = restart
	b := bar{urls: make(map[string]int64), totalFiles: len(urls)}
	for s := range d.Download(urls...) {
		if s.Error != nil {
			return s.Error
		}
		b.update(s)
	}
	return nil
}

func simpleDownload(url, dir string) error {
	pth := filepath.Join(dir, filepath.Base(url))
	h, err := os.Create(pth)
	if err != nil {
		return fmt.Errorf("could not create %s: %w", pth, err)
	}
	defer h.Close()
	c := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request %s: %w", url, err)
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("error requesting %s: %w", url, err)
	}
	defer resp.Body.Close()
	_, err = io.Copy(h, resp.Body)
	if err != nil {
		return fmt.Errorf("error writing to %s: %w", pth, err)
	}
	return nil
}
