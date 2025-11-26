package transformnext

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/text/encoding/charmap"
)

func writeJSONs(ctx context.Context, srcs map[string]*source, kv *kv, db database, dir string, privacy bool) error { // TODO: test
	bar, err := newProgressBar("[Step 2 of 2] Writing JSONs", 1)
	if err != nil {
		return fmt.Errorf("could not create a progress bar: %w", err)
	}
	defer func() {
		bar.AddMax(-1) // compensate for the extra byte added when creating the bar
	}()
	pths, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("could not read directory %s: %w", dir, err)
	}
	src := newSource("Estabelecimentos", ';', false, false)
	buf := &sync.Pool{
		New: func() any {
			return &bytes.Buffer{}
		},
	}
	var g errgroup.Group
	for _, pth := range pths {
		if !strings.HasPrefix(pth.Name(), src.prefix) {
			continue
		}
		p := pth
		g.Go(func() error {
			pth := filepath.Join(dir, p.Name())
			a, err := zip.OpenReader(pth)
			if err != nil {
				return fmt.Errorf("could not open archive %s: %w", pth, err)
			}
			defer func() {
				if err := a.Close(); err != nil {
					slog.Warn("could not close %s reader", "path", pth, "error", err)
				}
			}()
			var sg errgroup.Group
			for _, z := range a.File {
				sg.Go(func() error {
					bar.AddMax64(int64(z.UncompressedSize64))
					st := z.FileInfo()
					if st.IsDir() {
						return nil
					}
					f, err := z.Open()
					if err != nil {
						return fmt.Errorf("could not read %s from %s: %w", z.Name, pth, err)
					}
					defer func() {
						if err := f.Close(); err != nil {
							slog.Warn("Could not close csv reader", "path", pth, "name", z.Name, "error", err)
						}
					}()
					b := countReader{f, 0}
					r := csv.NewReader(charmap.ISO8859_15.NewDecoder().Reader(&b))
					r.Comma = src.sep
					var batch [][]string
					var prev int64
					for {
						select {
						case <-ctx.Done():
							return ctx.Err()
						default:
							row, err := r.Read()
							if err != nil {
								if errors.Is(err, io.EOF) {
									if len(batch) > 0 {
										if err := db.CreateCompanies(batch); err != nil {
											return fmt.Errorf("could not save batch: %w", err)
										}
									}
								}
								return fmt.Errorf("error reading %s: %w", pth, err)
							}
							if len(row) < 2 {
								return fmt.Errorf("unexpected row with %d columns in %s", len(row), src.prefix)
							}
							for n := range row {
								row[n] = cleanupColumn(row[n])
							}
							c, err := newCompany(srcs, kv, row)
							if err != nil {
								return fmt.Errorf("could not create company %v: %w", row[:3], err)
							}
							if privacy {
								c.withPrivacy()
							}
							j, err := c.JSON(buf)
							if err != nil {
								return err
							}
							batch = append(batch, []string{c.CNPJ, j})
							if len(batch) > 1<<13 {
								if err := db.CreateCompanies(batch); err != nil {
									return fmt.Errorf("could not save batch: %w", err)
								}
								batch = [][]string{}
							}
							s := b.read - prev
							if s > 0 {
								if err := bar.Add64(s); err != nil {
									slog.Warn("could not update the progress bar", "error", err)
								}
							}
							prev = b.read
						}
					}
				})
			}
			return sg.Wait()
		})
	}
	return g.Wait()
}
