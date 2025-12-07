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

	"github.com/cuducos/go-cnpj"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/sync/errgroup"
	"golang.org/x/text/encoding/charmap"
)

var multipleSpaces = regexp.MustCompile(`\s{2,}`)

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

type countReader struct {
	reader io.Reader
	read   int64
}

func (b *countReader) Read(p []byte) (int, error) {
	n, err := b.reader.Read(p)
	b.read += int64(n)
	return n, err
}

type reader struct {
	pth string
	src *source
}

func (c *reader) readFromReader(ctx context.Context, f io.Reader, bar *progressbar.ProgressBar, kv *kv) error {
	b := countReader{f, 0}
	r := csv.NewReader(charmap.ISO8859_15.NewDecoder().Reader(&b))
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
					return nil
				}
				return fmt.Errorf("error reading %s: %w", c.pth, err)
			}
			if len(row) < 2 {
				return fmt.Errorf("unexpected row with %d columns in %s", len(row), c.src.prefix)
			}
			for n := range row {
				row[n] = cleanupColumn(row[n])
			}
			key := row[0]
			val := row[1:]
			if c.src.key == "imu" || c.src.key == "arb" || c.src.key == "pre" || c.src.key == "rea" {
				key = cnpj.Base(row[1])
				val = append([]string{row[0]}, row[2:]...)
			}
			if err := kv.put(c.src, key, val); err != nil {
				return fmt.Errorf("could not save %s line %v to badger: %w", c.src.prefix, row, err)
			}
			s := b.read - prev
			if bar != nil && s > 0 {
				if err := bar.Add64(s); err != nil {
					slog.Warn("could not update the progress bar", "error", err)
				}
			}
			prev = b.read

		}
	}
}

func (c *reader) readArchivedCSV(ctx context.Context, bar *progressbar.ProgressBar, kv *kv) error {
	a, err := zip.OpenReader(c.pth)
	if err != nil {
		return fmt.Errorf("could not open archive %s: %w", c.pth, err)
	}
	defer func() {
		if err := a.Close(); err != nil {
			slog.Warn("could not close %s reader", "path", c.pth, "error", err)
		}
	}()
	var g errgroup.Group
	for _, z := range a.File {
		if bar != nil {
			bar.AddMax64(int64(z.UncompressedSize64))
		}
		st := z.FileInfo()
		if st.IsDir() {
			continue
		}
		f, err := z.Open()
		if err != nil {
			return fmt.Errorf("could not read %s from %s: %w", z.Name, c.pth, err)
		}
		r := f
		g.Go(func() error {
			defer func() {
				if err := r.Close(); err != nil {
					slog.Warn("Could not close csv reader", "path", c.pth, "name", z.Name, "error", err)
				}
			}()
			return c.readFromReader(ctx, r, bar, kv)
		})
	}
	return g.Wait()
}

func (c *reader) readCSV(ctx context.Context, bar *progressbar.ProgressBar, kv *kv) error {
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
		return fmt.Errorf("could not get stats for %s: %w", c.pth, err)
	}
	if bar != nil {
		bar.AddMax64(st.Size())
	}
	return c.readFromReader(ctx, f, bar, kv)
}

func loadCSVs(ctx context.Context, dir string, src *source, bar *progressbar.ProgressBar, kv *kv) error {
	if bar != nil {
		defer func() {
			bar.AddMax(-1) // compensate for the extra byte added when creating the bar
		}()
	}
	pths, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("could not read directory %s: %w", dir, err)
	}
	var g errgroup.Group
	for _, pth := range pths {
		if strings.HasPrefix(pth.Name(), src.prefix) {
			p := pth
			g.Go(func() error {
				pth := filepath.Join(dir, p.Name())
				r := reader{pth, src}
				switch filepath.Ext(p.Name()) {
				case ".zip":
					return r.readArchivedCSV(ctx, bar, kv)
				case ".csv":
					return r.readCSV(ctx, bar, kv)
				default:
					return fmt.Errorf("unexpected file extension for %s", pth)
				}
			})
		}
	}
	return g.Wait()
}
