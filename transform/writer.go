package transform

import (
	"compress/gzip"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const maxCSVBufferSize = 10_000

type resourceWriter struct {
	header []string
	path   string
	file   *os.File
	gz     *gzip.Writer
	csv    *csv.Writer
	buffer int
}

func (r *resourceWriter) write(ls [][]string) (int, error) {
	if r.csv == nil {
		return 0, errors.New("cannot write to a non-initialized CSV writer")
	}

	for i, l := range ls {
		if len(l) != len(r.header) {
			return 0, fmt.Errorf(
				"cannot write to CSV; the CSV has %d columns, but the line %d has %d columns",
				len(r.header),
				i+1,
				len(l),
			)
		}
	}

	var c int
	for _, l := range ls {
		err := r.csv.Write(l)
		if err != nil {
			return 0, err
		}
		c++
		r.buffer++
	}
	if r.buffer >= maxCSVBufferSize {
		r.csv.Flush()
		if err := r.csv.Error(); err != nil {
			return 0, err
		}
		r.buffer = 0
	}
	return len(ls), nil
}

func (r *resourceWriter) Close() error {
	r.csv.Flush()
	err := r.gz.Close()
	if err != nil {
		return err
	}
	err = r.file.Close()
	if err != nil {
		return err
	}
	return r.csv.Error()
}

func newResourceWriter(p string, h []string) (*resourceWriter, error) {
	f, err := os.Create(p)
	if err != nil {
		return nil, err
	}

	r := resourceWriter{header: h, path: p, file: f, gz: gzip.NewWriter(f)}
	r.csv = csv.NewWriter(r.gz)
	_, err = r.write([][]string{h})
	if err != nil {
		return nil, err
	}
	return &r, nil
}

type writers struct {
	company *resourceWriter
	partner *resourceWriter
	cnae    *resourceWriter
}

func (w *writers) Close() error {
	if err := w.company.Close(); err != nil {
		return err
	}
	if err := w.partner.Close(); err != nil {
		return err
	}
	if err := w.cnae.Close(); err != nil {
		return err
	}
	return nil
}

func newWriters(dir string) (*writers, error) {
	company, err := newResourceWriter(filepath.Join(dir, "empresa.csv.gz"), CompanySchema.Headers())
	if err != nil {
		return nil, err
	}
	partner, err := newResourceWriter(filepath.Join(dir, "socio.csv.gz"), PartnerSchema.Headers())
	if err != nil {
		return nil, err
	}
	cnae, err := newResourceWriter(filepath.Join(dir, "cnae_secundaria.csv.gz"), CNAESchema.Headers())
	if err != nil {
		return nil, err
	}
	return &writers{company, partner, cnae}, nil
}
