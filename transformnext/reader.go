package transformnext

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/sync/errgroup"
	"golang.org/x/text/encoding/charmap"
)

var multipleSpaces = regexp.MustCompile(`\s{2,}`)

type byteCountingReader struct {
	reader io.Reader
	count  int64
}

func (b *byteCountingReader) Read(p []byte) (int, error) {
	n, err := b.reader.Read(p)
	b.count += int64(n)
	return n, err
}

func removeNulChar(r rune) rune {
	if r == '\x00' {
		return -1
	}
	return r
}

func cleanupColumn(s string) string {
	s = strings.Map(removeNulChar, s)
	s = multipleSpaces.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

type reader struct {
	pth string
	src *source
	ch  chan<- []string
}

func (c *reader) readFromReader(ctx context.Context, f io.Reader) error {
	b := &byteCountingReader{reader: f}
	r := csv.NewReader(charmap.ISO8859_15.NewDecoder().Reader(b))
	r.Comma = c.src.sep
	if c.src.hasHeader {
		if _, err := r.Read(); err != nil {
			return fmt.Errorf("could not skip %s header: %w", c.pth, err)
		}
	}
	var prev int64
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			row, err := r.Read()
			if err != nil {
				if errors.Is(err, io.EOF) {
					c.src.done.Add(b.count - prev)
					return nil
				}
				return fmt.Errorf("error reading %s: %w", c.pth, err)
			}
			for n := range row {
				row[n] = cleanupColumn(row[n])
			}
			c.ch <- row
			c.src.done.Add(b.count - prev)
			prev = b.count
		}
	}
}

func (c *reader) readArchivedCSV(ctx context.Context) error {
	a, err := zip.OpenReader(c.pth)
	if err != nil {
		return fmt.Errorf("could not open archive %s: %w", c.pth, err)
	}
	defer func() {
		if err := a.Close(); err != nil {
			slog.Warn("could not close %s reader", "path", c.pth, "error", err)
		}
	}()
	for _, z := range a.File {
		st := z.FileInfo()
		if st.IsDir() {
			continue
		}
		f, err := z.Open()
		if err != nil {
			return fmt.Errorf("could not read %s from %s: %w", z.Name, c.pth, err)
		}
		err = func() error {
			defer func() {
				if err := f.Close(); err != nil {
					slog.Warn("Could not close csv reader", "path", c.pth, "name", z.Name, "error", err)
				}
			}()
			c.src.total.Add(int64(z.UncompressedSize64))
			return c.readFromReader(ctx, f)
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *reader) readCSV(ctx context.Context) error {
	f, err := os.Open(c.pth)
	if err != nil {
		return fmt.Errorf("could not open csv %s: %w", c.pth, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			slog.Warn("could not close csv reader", "path", c.pth, "error", err)
		}
	}()
	st, err := f.Stat()
	if err != nil {
		return fmt.Errorf("could not get %s info: %w", c.pth, err)
	}
	c.src.total.Add(st.Size())
	return c.readFromReader(ctx, f)
}

func readCSVs(ctx context.Context, dir string, src *source, ch chan<- []string) error {
	ps, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("could not read directory %s: %w", dir, err)
	}
	var g errgroup.Group
	for _, p := range ps {
		if strings.HasPrefix(p.Name(), src.prefix) {
			g.Go(func() error {
				pth := filepath.Join(dir, p.Name())
				r := reader{pth, src, ch}
				switch filepath.Ext(p.Name()) {
				case ".zip":
					return r.readArchivedCSV(ctx)
				case ".csv":
					return r.readCSV(ctx)
				default:
					return fmt.Errorf("unexpected file extension for %s", pth)
				}
			})
		}
	}
	return g.Wait()
}
