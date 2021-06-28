package adapter

import (
	"os"
	"path/filepath"
)

type kind string

const (
	company  kind = "EMPRECSV"
	facility      = "ESTABELE"
	partner       = "SOCIOCSV"
)

type Adapter struct {
	kind kind
	dir  string
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

// Transform unzips the downloaded files and merge them into CSV files.
func Transform(dir string) {
	var as []*Adapter
	for _, k := range []kind{company, facility, partner} {
		a := Adapter{k, dir}
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
			panic(err)
		}
	}
}
