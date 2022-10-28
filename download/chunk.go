package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/avast/retry-go/v4"
)

// DefaultChunkSize sets the size of the chunks to be dowloaded using HTTP
// requests by bytes range
const DefaultChunkSize = 4096

func totalChunksFor(f int64, c int) int {
	var pos int64
	var t int
	s := int64(c)
	for {
		if pos > f {
			break
		}
		pos += s - 1
		t++
	}
	return t
}

type chunk struct {
	url              string
	dest             *os.File
	idx              int
	start, end, size int64
	contents         []byte
}

func (c *chunk) save() error {
	if len(c.contents) == 0 {
		return nil
	}
	size, err := c.dest.WriteAt(c.contents, c.start)
	if err != nil {
		return fmt.Errorf("could not write to target file: %w", err)
	}
	if size != int(c.size) {
		return fmt.Errorf("should have written %d bytes, wrote %d", c.size, size)
	}
	return nil
}

func (c *chunk) downloadWithContext(ctx context.Context, h *http.Client) ([]byte, error) {
	req, err := http.NewRequest("GET", c.url, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create a request: %w", err)
	}
	req = req.WithContext(ctx)
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", c.start, c.end))
	resp, err := h.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending the http request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusPartialContent {
		return nil, fmt.Errorf("got %s from %s", resp.Status, c.url)
	}
	if resp.ContentLength != c.size {
		return nil, fmt.Errorf("got wrong content-length, expected %d, got %d", c.size, resp.ContentLength)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read chunk response body: %w", err)
	}
	return b, nil
}

func (c *chunk) downloadWithTimeout(h *http.Client) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), h.Timeout)
	ch := make(chan []byte)
	e := make(chan error)
	defer func() {
		defer cancel()
		defer close(ch)
		defer close(e)
	}()
	b, err := c.downloadWithContext(ctx, h)
	if err != nil {
		return nil, fmt.Errorf("%s chunk %d: errored with: %s", c.dest.Name(), c.idx+1, err)
	}
	return b, nil
}

func (c *chunk) download(h *http.Client, r int) ([]byte, error) {
	var b []byte
	err := retry.Do(
		func() error {
			d, err := c.downloadWithTimeout(h)
			if err != nil {
				return fmt.Errorf("%s chunk %d: failed after %d retries", c.dest.Name(), c.idx+1, r)
			}
			b = d
			return nil
		},
		retry.Attempts(uint(r)),
	)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func newChunk(url string, dest *os.File, retries, idx int, start, end int64) chunk {
	c := chunk{url: url, dest: dest, idx: idx, start: start, end: end}
	c.size = end - start + 1
	return c
}
