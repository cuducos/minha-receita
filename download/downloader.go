package download

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

// MaxRetries sets the maximum download attempt for each URL
const MaxRetries = 8

// MaxParallel setx the maximum parallel downloads
const MaxParallel = 8

type downloader struct {
	files       []file
	client      *http.Client
	totalSize   int64
	bar         *downloadProgressBar
	maxParallel int
	maxRetries  int
	done        int
}

func (d *downloader) getSize(url string) (int64, error) {
	r, err := d.client.Head(url)
	if err != nil {
		return 0, fmt.Errorf("error sending a http head request to %s: %s", url, err)
	}
	defer r.Body.Close()

	if r.ContentLength == 0 {
		return 0, fmt.Errorf("could not get size for %s", url)
	}
	return r.ContentLength, nil
}

func (d *downloader) getTotalSizeWorker(queue chan string, sizes chan int64, errors chan error) {
	for u := range queue {
		s, err := d.getSize(u)
		if err != nil {
			errors <- fmt.Errorf("error getting size of %s: %w", u, err)
			break
		}
		sizes <- s
	}
}

func (d *downloader) getTotalSize() error {
	d.totalSize = 0
	queue := make(chan string)
	sizes := make(chan int64)
	errors := make(chan error)
	for _, f := range d.files {
		go func(u string) { queue <- u }(f.url)
	}
	for i := 0; i < d.maxParallel; i++ {
		go d.getTotalSizeWorker(queue, sizes, errors)
	}
	defer func() {
		close(queue)
		close(errors)
		close(sizes)
	}()

	var c int
	for {
		select {
		case err := <-errors:
			return fmt.Errorf("error getting total size: %w", err)
		case s := <-sizes:
			d.totalSize += s
			c++
		}
		if c == len(d.files) {
			break
		}
	}
	return nil
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

	d.bar.updateBytes <- int64(-1) * i.Size()
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

		var h *os.File
		h, err = os.Create(f.path)
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
		err := d.download(f, 0)
		if err != nil {
			errors <- err
			break
		}
		d.bar.updateTotal <- struct{}{}
	}
}

func (d *downloader) downloadAll() error {
	queue := make(chan file)
	errors := make(chan error)
	for _, f := range d.files {
		go func(f file) { queue <- f }(f)
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
			d.bar.main.Add64(n)
		case <-d.bar.updateTotal:
			d.bar.done++
			d.bar.main.Describe(d.bar.description())
			if d.bar.isFinished() {
				return nil
			}
		}
	}
}

func newDownloader(c *http.Client, fs []file, p, r int) (*downloader, error) {
	d := downloader{files: fs, client: c, maxParallel: p, maxRetries: r}
	if err := d.getTotalSize(); err != nil {
		return nil, err
	}
	d.bar = &downloadProgressBar{
		total:       len(fs),
		updateBytes: make(chan int64),
		updateTotal: make(chan struct{}),
	}
	d.bar.main = progressbar.DefaultBytes(d.totalSize, d.bar.description())
	return &d, nil
}
