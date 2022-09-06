package download

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

// MaxRetries sets the maximum download attempt for each URL
const MaxRetries = 8

// MaxParallel setx the maximum parallel downloads
const MaxParallel = 8

type downloader struct {
	files          []file
	client         *http.Client
	totalSize      int64
	bar            *downloadProgressBar
	maxParallel    int
	maxRetries     int
	silent         bool
	isShuttingDown bool
	mutex          sync.Mutex
}

func (d *downloader) resetDownload(f file) error {
	h, err := os.Open(f.path)
	if err != nil {
		return fmt.Errorf("error cleaning up failed download %s: %w", f.path, err)
	}
	defer h.Close()

	i, err := h.Stat()
	if err != nil {
		return fmt.Errorf("could not get info for failed download %s: %v", f.path, err)
	}

	d.mutex.Lock()
	if !d.isShuttingDown {
		d.bar.updateBytes <- int64(-1) * i.Size()
	}
	d.mutex.Unlock()
	os.Remove(f.path)
	return nil
}

func (d *downloader) download(f file, a int) error {
	err := func(f file) error {
		r, err := d.client.Get(f.url)
		if err != nil {
			log.Output(2, fmt.Sprintf("HTTP request to %s failed: %v", f.url, err))
			return err
		}
		defer r.Body.Close()

		if r.StatusCode != http.StatusOK {
			return fmt.Errorf("http request to %s got %s", f.url, r.Status)
		}
		h, err := os.Create(f.path)
		if err != nil {
			return fmt.Errorf("failed to create %s: %v", f.path, err)
		}
		defer h.Close()

		_, err = io.Copy(io.MultiWriter(h, d.bar), r.Body)
		if err != nil {
			return fmt.Errorf("error downloading %s: %v", f.url, err)
		}
		return nil
	}(f)

	if err != nil {
		if err := d.resetDownload(f); err != nil {
			return fmt.Errorf("error resetting failed download %s: %w", f.path, err)
		}
		if a < d.maxRetries {
			time.Sleep(time.Duration(int(math.Pow(float64(2), float64(a)))) * time.Second)
			d.download(f, a+1)
			return nil
		}
		return fmt.Errorf("after %d attempts, could not download %s: %c", d.maxRetries, f.url, err)
	}
	return nil
}

func (d *downloader) downloadWorker(queue chan file, errors chan<- error) {
	for f := range queue {
		if err := d.download(f, 0); err != nil {
			func() {
				d.mutex.Lock()
				defer d.mutex.Unlock()

				if !d.isShuttingDown {
					d.isShuttingDown = true
					errors <- err
				}
			}()
			break
		}
		d.mutex.Lock()
		if !d.isShuttingDown {
			d.bar.updateTotal <- struct{}{}
		}
		d.mutex.Unlock()
	}
}

func (d *downloader) downloadAll() error {
	queue := make(chan file, len(d.files))
	errors := make(chan error)
	for _, f := range d.files {
		go func(f file) {
			d.mutex.Lock()
			if !d.isShuttingDown {
				queue <- f
			}
			d.mutex.Unlock()
		}(f)
	}
	for i := 0; i < d.maxParallel; i++ {
		go d.downloadWorker(queue, errors)
	}
	defer close(queue)

	for {
		select {
		case err := <-errors:
			return fmt.Errorf("error downloading files: %w", err)
		case n := <-d.bar.updateBytes:
			d.bar.addBytes(n)
			if d.bar.main.IsFinished() {
				return nil
			}
		case <-d.bar.updateTotal:
			d.bar.addFile()
		}
	}
}

func newDownloader(c *http.Client, fs []file, p, r int, s bool) (*downloader, error) {
	d := downloader{files: fs, client: c, maxParallel: p, maxRetries: r, silent: s}
	for _, f := range fs {
		d.totalSize += f.size
	}
	d.bar = &downloadProgressBar{
		total:       len(fs),
		updateBytes: make(chan int64),
		updateTotal: make(chan struct{}),
	}
	newBar := progressbar.DefaultBytes
	if s {
		newBar = progressbar.DefaultBytesSilent
	}
	d.bar.main = newBar(d.totalSize, d.bar.description())
	return &d, nil
}
