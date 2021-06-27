package adapter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
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

func (a *Adapter) status() string {
	p := csvPathFor(a)
	s, err := os.Stat(p)
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("%s (%s)", p, humanize.Bytes(uint64(s.Size())))
}

func status(as []*Adapter) {
	p := 0
	s := []string{"⠋", "⠙", "⠸", "⠴", "⠦", "⠇"}

	for {
		l := []string{}
		for _, a := range as {
			l = append(l, a.status())
		}

		fmt.Printf(fmt.Sprintf("\r%s %s", s[p], strings.Join(l, " | ")))
		time.Sleep(1 * time.Second)

		p++
		if p >= len(s) {
			p = 0
		}
	}
}

// Transform unzips the downloaded files and merge them into CSV files.
func Transform(dir string) {
	as := []*Adapter{
		&Adapter{company, dir},
		&Adapter{facility, dir},
		&Adapter{partner, dir},
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
