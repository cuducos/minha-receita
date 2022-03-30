package sample

import (
    "io"
	"archive/zip"
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
)

// MaxLines to use when creating sample data
const MaxLines = 10000

// TargetDir to use when creating sample data
const TargetDir = "sample_data"

func makeSample(src, outDir string, m int) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("error opening %s: %w", src, err)
	}

	name := filepath.Base(src)
	base := strings.TrimSuffix(name, filepath.Ext(src))
	out := filepath.Join(outDir, name)
	for _, z := range r.File {
		if z.Name != base {
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
		var c int
		s := bufio.NewScanner(fSrc)
		for s.Scan() {
			c++
			if c > m {
				break
			}
			t := s.Text() + "\n"
			_, err := fOut.Write([]byte(t))
			if err != nil {
				return fmt.Errorf("error writing to %s in %s: %w", name, out, err)
			}
		}
		if err := s.Err(); err != nil {
			return fmt.Errorf("error reading lines from %s in %s: %w", z.Name, src, err)
		}
		break
	}
	return nil
}

func copy(src, dst, filename string) (int64, error) {
        src_path := filepath.Join(src, filename)
        dst_path := filepath.Join(src, filename)
        sourceFileStat, err := os.Stat(src)
        if err != nil {
                return 0, err
        }

        if !sourceFileStat.Mode().IsRegular() {
                return 0, fmt.Errorf("%s is not a regular file", src)
        }

        source, err := os.Open(src_path)
        if err != nil {
                return 0, err
        }
        defer source.Close()

        destination, err := os.Create(dst_path)
        if err != nil {
                return 0, err
        }
        defer destination.Close()
        nBytes, err := io.Copy(destination, source)
        return nBytes, err
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
	bar := progressbar.Default(int64(len(ls) + 1))
	bar.Describe("Creating sample files")
	if err := bar.RenderBlank(); err != nil {
		return fmt.Errorf("error rendering the progress bar: %w", err)
	}

    tabmun := "TABMUN.CSV"
    dados_abertos := "Dados%20Abertos%20S%c3%adtio%20RFB%20Extracao%2020.10.2021.zip"

    copy(src, target, tabmun)
    bar.Add(1)

    copy(src, target, dados_abertos)
    bar.Add(1)

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
