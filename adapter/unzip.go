package adapter

import (
	"archive/zip"
	"encoding/csv"
	"io"
	"log"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/ulikunitz/xz"
	"golang.org/x/text/encoding/charmap"
)

func cleanLine(l []string) ([]string, error) {
	var c []string
	var err error

	for _, v := range l {
		v := strings.TrimSpace(v)
		if !utf8.ValidString(v) {
			v, err = charmap.ISO8859_1.NewDecoder().String(v)
			if err != nil {
				return nil, err
			}
		}
		c = append(c, strings.TrimSpace(v))
	}
	return c, nil
}

func lineProducer(a *Adapter, l chan<- []string, f *zip.File) error {
	z, err := f.Open()
	if err != nil {
		return err
	}
	defer z.Close()

	r := csv.NewReader(z)
	r.Comma = separator
	for {
		s, err := r.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		l <- s
	}
}

func lineConsumer(a *Adapter, l chan []string) {
	f, err := os.OpenFile(csvPathFor(a), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	x, err := xz.NewWriter(f)
	if err != nil {
		log.Fatal(err)
	}
	defer x.Close()

	w := csv.NewWriter(x)
	for {
		s, ok := <-l
		if !ok {
			w.Flush()
			return
		}

		if err := w.Write(s); err != nil {
			log.Fatal(err)
		}
	}
}

func unzip(a *Adapter, e chan<- error, l chan<- []string, s string) {
	z, err := zip.OpenReader(s)
	if err != nil {
		e <- err
		return
	}
	defer z.Close()

	for _, f := range z.File {
		if err := lineProducer(a, l, f); err != nil {
			e <- err
			return
		}
	}

	e <- nil
}

func writeCsvsFor(a *Adapter, q chan<- error) {
	ls, err := a.files()
	if err != nil {
		q <- err
		return
	}

	e := make(chan error)
	l := make(chan []string)
	for _, f := range ls {
		go unzip(a, e, l, f)
	}

	go lineConsumer(a, l)
	for range ls {
		err := <-e
		if err != nil {
			q <- err
			return
		}
	}
	close(l)

	q <- nil
	return
}
