package download

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/melbahja/got"
)

const (
	// MaxRetries sets the maximum download attempt for each URL
	MaxRetries = 8

	// MaxParallel sets the maximum parallels downloads
	MaxParallel = 8
)

type downloader struct {
	files          []file
	got            *got.Got
	bar            *downloadProgressBar
	maxRetries     uint
	maxParallel    uint
	isShuttingDown bool
	mutex          sync.Mutex
}

func (d *downloader) download(f file, a uint) error {
	err := func(f file) error {
		var done uint32
		g := &got.Download{URL: f.url, Dest: f.path, Client: d.got.Client}
		go func(g *got.Download, done *uint32) {
			for {
				d.bar.updateBytes <- bytesProgress{g.Dest, g.Size()}
				if atomic.LoadUint32(done) == 1 {
					break
				}
				time.Sleep(1 * time.Second)
			}
		}(g, &done)
		err := d.got.Do(g)
		atomic.StoreUint32(&done, 1)
		if err != nil {
			return fmt.Errorf("error downloading %s with got package: %v", f.url, err)
		}
		return nil
	}(f)

	if err != nil {
		d.bar.updateBytes <- bytesProgress{f.path, 0}
		if err := os.Remove(f.path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("error cleaning up %s: %w", f.path, err)
		}
		if a < d.maxRetries {
			return d.download(f, a+1)
		}
		return fmt.Errorf("after %d attempts, could not download %s: %c", d.maxRetries, f.url, err)
	}
	return nil
}

func (d *downloader) worker(queue chan file, errors chan<- error) {
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
			d.bar.updateFiles <- struct{}{}
		}
		d.mutex.Unlock()
	}
}

func (d *downloader) downloadAll() error {
	queue := make(chan file, len(d.files))
	errors := make(chan error)
	done := make(chan struct{})
	go func() {
		done <- d.bar.run()
	}()
	for _, f := range d.files {
		go func(f file) {
			d.mutex.Lock()
			if !d.isShuttingDown {
				queue <- f
			}
			d.mutex.Unlock()
		}(f)
	}
	for i := 0; i < int(d.maxParallel); i++ {
		go d.worker(queue, errors)
	}
	defer func() {
		close(done)
		close(queue)
		close(errors)
	}()
	for {
		select {
		case err := <-errors:
			return fmt.Errorf("error downloading files: %w", err)
		case <-done:
			return nil
		}
	}
}

func newDownloader(c *http.Client, fs []file, p, r uint, s bool) (*downloader, error) {
	d := downloader{
		files:       fs,
		got:         got.New(),
		maxRetries:  r,
		maxParallel: p,
	}
	d.got.Client = c
	var t uint64
	for _, f := range fs {
		t += f.size
	}
	d.bar = newBar(uint(len(fs)), t, s)
	return &d, nil
}
