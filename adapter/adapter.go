package adapter

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz"
)

type kind string

const (
	company  kind = "EMPRECSV"
	facility      = "ESTABELE"
	partner       = "SOCIOCSV"
)

const CompressionAlgorithms = "xz, gz"

type Adapter struct {
	kind        kind
	dir         string
	compression string
}

func (a *Adapter) files() ([]string, error) {
	var o []string

	ls, err := os.ReadDir(a.dir)
	if err != nil {
		return []string{}, err
	}

	for _, f := range ls {
		if !f.IsDir() {
			o = append(o, filepath.Join(a.dir, f.Name()))
		}
	}
	return filesFor(a, o), nil
}

func (a *Adapter) Writer(i io.WriteCloser) (io.WriteCloser, error) {
	switch a.compression {
	case "xz":
		return xz.NewWriter(i)
	case "gz":
		return gzip.NewWriter(i), nil
	default:
		return i, nil
	}
}

func validateCompressionAlgorithm(c string) bool {
	if c == "" {
		return true
	}

	for _, o := range strings.Split(CompressionAlgorithms, ",") {
		if c == strings.TrimSpace(o) {
			return true
		}
	}

	return false
}

// Transform unzips the downloaded files and merge them into CSV files.
func Transform(dir string, compression string) error {
	if !validateCompressionAlgorithm(compression) {
		return fmt.Errorf(
			"Unknown compression algorithm %s, options are: %s",
			compression,
			CompressionAlgorithms,
		)
	}

	var as []*Adapter
	for _, k := range []kind{company, facility, partner} {
		a := Adapter{k, dir, compression}
		as = append(as, &a)
	}

	c := make(chan error)
	for _, a := range as {
		createCsvFor(a)
		go writeCsvsFor(a, c)
	}

	go status(as)

	for range as {
		err := <-c
		if err != nil {
			return err
		}
	}
	return nil
}
