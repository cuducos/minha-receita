package check

import (
	"archive/zip"
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func checkZipFile(pth string) error {
	r, err := zip.OpenReader(pth)
	if err != nil {
		return fmt.Errorf("error opening %s: %w", pth, err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			slog.Error("could not close", "path", pth, "error", err)
		}
	}()
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		r, err := f.Open()
		if err != nil {
			return fmt.Errorf("error opening %s in %s: %w", f.Name, pth, err)
		}
		s := bufio.NewScanner(r)
		for s.Scan() {
			continue
		}
		if err := s.Err(); err != nil {
			return fmt.Errorf("error reading %s in %s: %w", f.Name, pth, err)
		}
	}
	return nil
}

type check struct {
	path string
	err  error
}

func checkZipFiles(dir string) (map[string]error, error) {
	r := make(map[string]error)
	ls, err := filepath.Glob(filepath.Join(dir, "*.zip"))
	if err != nil {
		return r, fmt.Errorf("error listing zip files: %w", err)
	}
	if len(ls) == 0 {
		return r, fmt.Errorf("no zip files found")
	}
	slog.Info(fmt.Sprintf("Checking %d files…\n", len(ls)))
	checks := make(chan check)
	for _, pth := range ls {
		go func(pth string) {
			err := checkZipFile(pth)
			if err != nil {
				slog.Error("Failed checking", "path", pth, "error", err)
			}
			checks <- check{pth, err}
		}(pth)
	}
	for range ls {
		c := <-checks
		if c.err != nil {
			r[c.path] = c.err
		}
	}
	return r, nil
}

func Check(dir string, del bool) error {
	fails, err := checkZipFiles(dir)
	if err != nil {
		return fmt.Errorf("error checking zip files in %s: %w", dir, err)
	}
	if len(fails) != 0 {
		if del {
			for f := range fails {
				slog.Info("Deleting", "file", f)
				if err := os.Remove(f); err != nil {
					return fmt.Errorf("error deleting %s: %w", f, err)
				}
			}
			return nil
		}
		return errors.New("error checking the zip files above")
	}
	return nil
}
