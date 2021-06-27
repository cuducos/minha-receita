package adapter

import (
	"archive/zip"
	"bufio"
	"log"
	"os"
)

func lineProducer(a *Adapter, l chan<- string, f *zip.File) error {
	r, err := f.Open()
	if err != nil {
		return err
	}
	defer r.Close()

	s := bufio.NewScanner(r)
	for s.Scan() {
		l <- s.Text() + "\n"
	}

	return nil
}

func lineConsumer(a *Adapter, l chan string) {
	f, err := os.OpenFile(csvPathFor(a), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	for {
		s, ok := <-l
		if !ok {
			return
		}

		if _, err := f.WriteString(s); err != nil {
			log.Fatal(err)
		}
	}
}

func unzip(a *Adapter, e chan<- error, l chan<- string, s string) {
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
	l := make(chan string)
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
