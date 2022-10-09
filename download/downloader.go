package download

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
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

	// DefaultChunkSize sets the size of the chunks to be dowloaded using HTTP
	// requests by bytes range
	DefaultChunkSize = 4096

	// DefaultTimeout sets the timeout for each HTTP request
	DefaultTimeout = 3 * time.Minute
)

type chunk struct {
	url              string
	dest             *os.File
	idx              int
	retries          int
	start, end, size uint64
	err              error
	contents         []byte
}

func newChunk(url string, dest *os.File, retries, idx int, start, end uint64) chunk {
	c := chunk{url: url, dest: dest, idx: idx, retries: retries, start: start, end: end}
	c.size = end - start + 1
	return c
}

func retryChunk(chunk chunk) chunk {
	return newChunk(chunk.url, chunk.dest, chunk.idx, chunk.retries-1, chunk.start, chunk.end)
}

type downloadStatus struct{ chunks, done uint64 }

type chunckDownloader struct {
	files     []file
	handlers  []*os.File
	client    *http.Client
	chunkSize uint64
	retries   int
	queue     chan chunk
	results   chan chunk
	bar       *progressbar.ProgressBar
	status    map[string]*downloadStatus
	logger    *log.Logger
}

func (c *chunckDownloader) log(fmt string, v ...interface{}) {
	if c.logger == nil {
		return
	}
	c.logger.Printf(fmt, v...)
}

func (c *chunckDownloader) downloadChunk(chunk chunk) {
	c.log("%s chunk %d: starting download (remaning retries %d)", chunk.dest.Name(), chunk.idx+1, chunk.retries)
	defer func() { c.results <- chunk }()
	req, err := http.NewRequest("GET", chunk.url, nil)
	if err != nil {
		chunk.err = fmt.Errorf("could not create a request: %w", err)
		return
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", chunk.start, chunk.end))
	if c.logger != nil {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			started := time.Now()
			tick := time.NewTicker(5 * time.Second)
			elapsed := func() string {
				d := time.Since(started)
				return d.Round(time.Second).String()

			}
			for {
				select {
				case <-ctx.Done():
					c.log("%s chunk %d download stopped after %s", chunk.dest.Name(), chunk.idx, elapsed())
					return
				case <-tick.C:
					c.log("%s chunk %d downloading for %s", chunk.dest.Name(), chunk.idx, elapsed())
				}
			}
		}()
	}
	resp, err := c.client.Do(req)
	if err != nil {
		chunk.err = fmt.Errorf("error sending the http request: %w", err)
		return
	}
	defer resp.Body.Close()
	if resp.ContentLength != int64(chunk.size) {
		chunk.err = fmt.Errorf("got wrong content-length, expected %d, got %d", chunk.size, resp.ContentLength)
		return
	}
	chunk.contents, err = io.ReadAll(resp.Body)
	if err != nil {
		chunk.err = fmt.Errorf("could not read chunk response body: %w", err)
		return
	}
	if err != nil {
		chunk.err = fmt.Errorf("could not write chunk to file: %w", err)
	}
}

func (c *chunckDownloader) worker() {
	for chunk := range c.queue {
		c.downloadChunk(chunk)
	}
}

func (c *chunckDownloader) prepareDownload(file file, idx int, errs chan<- error) {
	var err error
	c.handlers[idx], err = os.Create(file.path)
	if err != nil {
		errs <- fmt.Errorf("could not create %s: %w", file.path, err)
		return
	}
	if err := c.handlers[idx].Truncate(int64(file.size)); err != nil {
		errs <- fmt.Errorf("could not truncate %s: %w", file.path, err)
		return
	}
	var i int
	var start, end uint64
	for {
		if start > file.size {
			break
		}
		end = (start + c.chunkSize) - 1
		if end > (file.size - 1) {
			end = file.size - 1
		}
		go func(i int, start, end uint64) {
			c.log("%s chunk %d size: %d", file.path, i+1, end-start+1)
			c.queue <- newChunk(file.url, c.handlers[idx], c.retries, i, start, end)
		}(i, start, end)
		start += c.chunkSize - 1
		i++
	}
}

func (c *chunckDownloader) progressBarDescription() {
	var t int
	for _, s := range c.status {
		if s.chunks == s.done {
			t += 1
		}
	}
	c.bar.Describe(fmt.Sprintf("Downloading (%d of %d files done)", t, len(c.files)))
}

func (c *chunckDownloader) updateProgressBar(path string, size int) {
	c.bar.Add(size)
	c.status[path].done += 1
	c.progressBarDescription()
}

func (c *chunckDownloader) cleanUp() error {
	for path, s := range c.status {
		if s.chunks != s.done {
			c.log("%s expected to have %d chunks downloaded, got %d: deleting %s", path, s.chunks, s.done, path)
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("error deleting %s: %w", path, err)
			}
		}

	}
	return nil
}

func download(client *http.Client, files []file, parallel, retries int, chunkSize uint64) error {
	c := chunckDownloader{
		files:     files,
		handlers:  make([]*os.File, len(files)),
		client:    client,
		chunkSize: chunkSize,
		retries:   retries,
		queue:     make(chan chunk),
		results:   make(chan chunk),

		status: make(map[string]*downloadStatus, len(files)),
	}
	defer func() {
		for _, h := range c.handlers {
			h.Close()
		}
		close(c.queue)
		close(c.results)
		c.cleanUp()
	}()
	if os.Getenv("DEBUG") != "" {
		tmp, err := os.CreateTemp("", fmt.Sprintf("minha-receita-download-%s-", time.Now().Format("2006-01-02-150405")))
		if err != nil {
			return fmt.Errorf("could not create log file: %w", err)
		}
		log.Output(2, fmt.Sprintf("Creating detailed logs at %s", tmp.Name()))
		defer tmp.Close()
		c.logger = log.New(tmp, "", log.LstdFlags)
	}
	var total uint64
	for _, f := range files {
		c.log("%s total size: %d", f.path, f.size)
		total += f.size
	}
	c.log("total size: %d", total)
	c.bar = progressbar.NewOptions64(
		int64(total),
		progressbar.OptionShowBytes(true),
		progressbar.OptionThrottle(500*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
	)
	c.progressBarDescription()
	if err := c.bar.RenderBlank(); err != nil {
		return fmt.Errorf("could not render the progress bar: %w", err)
	}
	for i := 0; i < parallel; i++ {
		go c.worker()
	}
	errs := make(chan error)
	for idx, f := range c.files {
		go c.prepareDownload(f, idx, errs)
		count := uint64(math.Ceil(float64(f.size) / float64(c.chunkSize)))
		if count == 0 {
			count = 1
		}
		c.log("%s will need %d chunks", f.path, count)
		if count != 1 {
			newTotal := c.bar.GetMax64() + int64(count-1)
			c.log("%s adjusting the progress bar from %d to %d", f.path, c.bar.GetMax64(), newTotal)
			c.bar.ChangeMax64(newTotal) // adjust for extra EOF bytes
		}
		c.status[f.path] = &downloadStatus{chunks: count}
	}
	for {
		select {
		case err := <-errs:
			return err
		case chunk := <-c.results:
			path := chunk.dest.Name()
			idx := chunk.idx + 1
			if chunk.err != nil {
				c.log("%s chunk %d errored with: %s", path, idx, chunk.err.Error())
				if chunk.retries != 0 {
					c.log("%s chunk %d: retrying", path, idx)
					c.queue <- retryChunk(chunk)
				} else {
					return fmt.Errorf("could not download %s: %w", path, chunk.err)
				}
			}
			size, err := chunk.dest.WriteAt(chunk.contents, int64(chunk.start))
			if err != nil {
				return fmt.Errorf("could not write to target file: %w", err)
			}
			c.updateProgressBar(path, size)
			c.log("%s chunk %d: progress is %d out of %d", path, idx, int(c.bar.State().CurrentBytes), c.bar.GetMax64())
		}
		if c.bar.IsFinished() {
			return nil
		}
	}
}
