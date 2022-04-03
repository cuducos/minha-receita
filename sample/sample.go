package sample

import (
	"archive/zip"
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cuducos/minha-receita/transform"
	"github.com/schollz/progressbar/v3"
)

// MaxLines to use when creating sample data
const MaxLines = 10000

// TargetDir to use when creating sample data
const TargetDir = "sample"

func sampleLines(r io.Reader, w io.Writer, m int) error {
	var c int
	s := bufio.NewScanner(r)
	for s.Scan() {
		c++
		if c > m {
			break
		}
		t := s.Text() + "\n"
		_, err := w.Write([]byte(t))
		if err != nil {
			return fmt.Errorf("error writing sample: %w", err)
		}
	}
	if err := s.Err(); err != nil {
		return fmt.Errorf("error reading lines: %w", err)
	}
	return nil
}

func makeSampleFromCSV(src, outDir string, m int) error {
	name := filepath.Base(src)
	out := filepath.Join(outDir, name)

	r, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening %s: %w", src, err)
	}
	defer r.Close()

	w, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("error creating %s: %w", out, err)
	}
	defer w.Close()

	if err := sampleLines(r, w, m); err != nil {
		return fmt.Errorf("error creating sample %s from %s: %w", out, src, err)
	}

	return nil
}

func makeSampleFromZIP(src, outDir string, m int) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("error opening %s: %w", src, err)
	}

	name := filepath.Base(src)
	base := strings.TrimSuffix(name, filepath.Ext(src))
	out := filepath.Join(outDir, name)
	for _, z := range r.File {
		if z.Name != base && filepath.Ext(z.Name) == "csv" {
			continue
		}
		if z.FileInfo().IsDir() {
			continue
		}
		fSrc, err := z.Open()
		if err != nil {
			return fmt.Errorf("error reading file %s in %s: %w", z.Name, src, err)
		}
		defer fSrc.Close()

		o, err := os.Create(out)
		if err != nil {
			return fmt.Errorf("error creating %s: %w", out, err)
		}
		defer o.Close()

		buf := bufio.NewWriter(o)
		w := zip.NewWriter(buf)
		defer w.Close()

		fOut, err := w.Create(base)
		if err != nil {
			return fmt.Errorf("error creating %s in %s: %w", name, out, err)
		}
		if err := sampleLines(fSrc, fOut, m); err != nil {
			return fmt.Errorf(
				"error creating sample %s from %s in %s: %w",
				out,
				z.Name,
				src,
				err,
			)
		}
		break
	}
	return nil
}

func makeSample(src, outDir string, m int) error {
	ext := strings.ToLower(filepath.Ext(src))
	switch ext {
	case ".zip":
		return makeSampleFromZIP(src, outDir, m)
	case ".csv":
		return makeSampleFromCSV(src, outDir, m)
	}
	return fmt.Errorf("no make sample handler for %s", ext)
}

// Sample generates sample data on the target directory, coping the first `m`
// lines of each file from the source directory.
func Sample(src, target string, m int) error {
	if src == target {
		return fmt.Errorf("data directory and target directory cannot be the same")
	}
	if err := os.MkdirAll(target, 0755); err != nil {
		return fmt.Errorf("error creating directory %s: %w", target, err)
	}
	ls, err := filepath.Glob(filepath.Join(src, "*.zip"))
	if err != nil {
		return fmt.Errorf("error looking for zip files in %s: %w", target, err)
	}
	if len(ls) == 0 {
		return errors.New("source directory %s has no zip files")
	}
	ls = append(ls, filepath.Join(src, transform.NationalTreasureFileName))

	bar := progressbar.Default(int64(len(ls)))
	bar.Describe("Creating sample files")
	if err := bar.RenderBlank(); err != nil {
		return fmt.Errorf("error rendering the progress bar: %w", err)
	}

	q := make(chan error)
	defer close(q)
	for _, pth := range ls {
		go func(pth string) {
			q <- makeSample(pth, target, m)
		}(pth)
	}
	for err := range q {
		if err != nil {
			return err
		}
		bar.Add(1)
		if bar.IsFinished() {
			return nil
		}
	}
	return nil
}
