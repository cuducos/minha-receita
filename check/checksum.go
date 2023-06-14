package check

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/schollz/progressbar/v3"
)

func checksumFor(path string) (string, error) {
	h := md5.New()

	r, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("error opening %s: %w", path, err)
	}
	defer r.Close()

	if _, err := io.Copy(h, r); err != nil {
		return "", fmt.Errorf("error reading %s content: %w", path, err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func checkChecksum(src, target string) (bool, error) {
	name := filepath.Base(src)
	sc, err := checksumFor(src)
	if err != nil {
		return false, fmt.Errorf("error getting %s checksum: %w", src, err)
	}

	pth := filepath.Join(target, name)
	tc, err := checksumFor(pth)
	if err != nil {
		return false, fmt.Errorf("error getting %s checksum: %w", pth, err)
	}

	return sc == tc, nil
}

// CheckChecksum compares md5 checksum files from source and target directories.
func CheckChecksum(src, target string) error {
	ls, err := filepath.Glob(filepath.Join(src, "*.md5"))
	if err != nil {
		return fmt.Errorf("error listing checksum files in %s directory: %w", src, err)
	}
	if len(ls) == 0 {
		return fmt.Errorf("target directory %s has no checksum files to compare with", target)
	}

	bar := progressbar.Default(int64(len(ls)))
	defer bar.Close()
	bar.Describe("Checking files checksum")
	if err := bar.RenderBlank(); err != nil {
		return fmt.Errorf("error rendering the progress bar: %w", err)
	}

	type qm struct {
		f     string
		equal bool
		err   error
	}
	q := make(chan qm, len(ls))
	var wg sync.WaitGroup

	for _, f := range ls {
		wg.Add(1)

		go func(f string, wg *sync.WaitGroup, q chan<- qm) {
			defer wg.Done()

			equal, err := checkChecksum(f, target)

			q <- qm{f, equal, err}
		}(f, &wg, q)
	}

	go func(wg *sync.WaitGroup) {
		wg.Wait()
		close(q)
	}(&wg)

	var notEqual []string
	for m := range q {
		if m.err != nil {
			return err
		}

		bar.Add(1)

		if !m.equal {
			notEqual = append(notEqual, m.f)
		}

		if bar.IsFinished() {
			if len(notEqual) > 0 {
				return fmt.Errorf("got different checksum for file(s): %v", notEqual)
			}

			log.Output(1, "OK!")
			return nil
		}
	}

	return nil
}

func createChecksum(src string) error {
	checksum, err := checksumFor(src)
	if err != nil {
		return fmt.Errorf("error getting checksum for %s: %w", src, err)
	}

	pth := fmt.Sprintf("%s.md5", src)
	if err := os.WriteFile(pth, []byte(checksum), 0755); err != nil {
		return fmt.Errorf("error writing %s checksum file: %w", pth, err)
	}

	return nil
}

// CreateChecksum creates an MD5 checksum file for each file in the source directory.
func CreateChecksum(src string) error {
	ls, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("error reading %s directory: %w", src, err)
	}

	bar := progressbar.Default(int64(len(ls)))
	defer bar.Close()
	bar.Describe("Creating checksum files")
	if err := bar.RenderBlank(); err != nil {
		return fmt.Errorf("error rendering the progress bar: %w", err)
	}

	q := make(chan error, len(ls))
	var wg sync.WaitGroup

	for _, f := range ls {
		if f.IsDir() || f.Name()[0] == '.' || strings.HasSuffix(f.Name(), ".md5") {
			continue
		}

		wg.Add(1)

		go func(f string, wg *sync.WaitGroup, q chan<- error) {
			defer wg.Done()

			q <- createChecksum(filepath.Join(src, f))
		}(f.Name(), &wg, q)
	}

	go func(wg *sync.WaitGroup) {
		wg.Wait()
		close(q)
	}(&wg)

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
