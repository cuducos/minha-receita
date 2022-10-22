package download

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

const (
	// DefaultMaxRetries sets the maximum download attempt for each URL
	DefaultMaxRetries = 32

	// DefaultMaxParallel sets the maximum parallels downloads
	DefaultMaxParallel = 16

	// DefaultTimeout sets the timeout for each HTTP request
	DefaultTimeout = 3 * time.Minute
)

type downloadStatus struct{ chunks, done int }

func (s *downloadStatus) IsFinished() bool { return s.done == s.chunks }

type chunkDownloader struct {
	files              []file
	handlers           []*os.File
	client             *http.Client
	chunkSize, retries int
	queue, results     chan chunk
	errors             chan error
	bar                *progressbar.ProgressBar
	status             map[string]*downloadStatus
	recover            *recover
}

func (c *chunkDownloader) worker() {
	for k := range c.queue {
		b, err := k.download(c.client, c.retries)
		if err != nil {
			c.errors <- err
		}
		k.contents = b
		c.results <- k
	}
}

func (c *chunkDownloader) prepareDownload(file file, idx int) {
	var err error
	c.handlers[idx], err = os.Create(file.path)
	if err != nil {
		c.errors <- fmt.Errorf("could not create %s: %w", file.path, err)
		return
	}
	if err := c.handlers[idx].Truncate(file.size); err != nil {
		c.errors <- fmt.Errorf("could not truncate %s: %w", file.path, err)
		return
	}
	count := int(totalChunksFor(file.size, c.chunkSize))
	c.recover.addFile(file.path, count)

	var start, end int64
	s := int64(c.chunkSize)
	for i := 0; i < count; i++ {
		end = (start + s) - 1
		if end > (file.size - 1) {
			end = file.size - 1
		}
		k := newChunk(file.url, c.handlers[idx], c.retries, i, start, end)
		if c.recover.shouldDownload(file.path, i) {
			c.queue <- k
		} else {
			c.results <- k
		}
		start += s - 1
	}
}

func (c *chunkDownloader) progressBarDescription() {
	var t int
	for _, s := range c.status {
		if s.IsFinished() {
			t += 1
		}
	}
	c.bar.Describe(fmt.Sprintf("Downloading (%d of %d files done)", t, len(c.files)))
}

func (c *chunkDownloader) updateProgressBar(path string, size int64) {
	c.bar.Add64(size)
	c.status[path].done += 1
	c.progressBarDescription()
}

func (c *chunkDownloader) handleResult(k chunk) error {
	if err := k.save(); err != nil {
		return fmt.Errorf("could not write chunk %d to %s: %w", k.idx+1, k.dest.Name(), err)
	}
	c.recover.chunkDone(k.dest.Name(), k.idx)
	if err := c.recover.save(); err != nil {
		return fmt.Errorf("error updating recover file: %w", err)
	}
	c.updateProgressBar(k.dest.Name(), k.size)
	return nil
}

func download(client *http.Client, files []file, r *recover, parallel, retries, chunkSize int) error {
	c := chunkDownloader{
		files:     files,
		handlers:  make([]*os.File, len(files)),
		client:    client,
		chunkSize: chunkSize,
		recover:   r,
		retries:   retries,
		queue:     make(chan chunk),
		results:   make(chan chunk),
		errors:    make(chan error),
		status:    make(map[string]*downloadStatus, len(files)),
	}
	defer func() {
		for _, h := range c.handlers {
			h.Close()
		}
		close(c.queue)
		close(c.results)
		close(c.errors)
	}()
	var t int64
	for _, f := range files {
		t += f.size
	}
	c.bar = progressbar.DefaultBytes(t, "Downloading")
	defer c.bar.Close()
	c.progressBarDescription()
	if err := c.bar.RenderBlank(); err != nil {
		return fmt.Errorf("could not render the progress bar: %w", err)
	}
	for i := 0; i < parallel; i++ {
		go c.worker()
	}
	for idx, f := range c.files {
		n := totalChunksFor(f.size, c.chunkSize)
		c.bar.ChangeMax64(c.bar.GetMax64() + int64(n-1)) // adjust for extra EOF bytes
		c.status[f.path] = &downloadStatus{chunks: n}
		go c.prepareDownload(f, idx)
	}
	for {
		select {
		case err := <-c.errors:
			return err
		case k := <-c.results:
			if err := c.handleResult(k); err != nil {
				return err
			}
		}
		if c.bar.IsFinished() {
			return nil
		}
	}
}
