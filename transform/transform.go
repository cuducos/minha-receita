package transform

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
)

func parseZipFile(wg *sync.WaitGroup, c chan<- parsedLine, z *zippedFile) {
	wg.Add(1)
	defer wg.Done()
	defer z.Close()

	s := bufio.NewScanner(z.firstFile)
	for s.Scan() {
		l := parseLine(s.Text())
		if l.valid {
			c <- l
		}
	}
}

func status(w *writers, f, c int) error {
	company, err := os.Stat(w.company.path)
	if err != nil {
		return err
	}
	partner, err := os.Stat(w.partner.path)
	if err != nil {
		return err
	}
	cnae, err := os.Stat(w.cnae.path)
	if err != nil {
		return err
	}

	fmt.Printf(
		"\rFixed-width lines read: %s | CSV lines written: %s | %s: %s | %s: %s | %s: %s ",
		humanize.Comma(int64(f)),
		humanize.Comma(int64(c)),
		w.company.path,
		humanize.Bytes(uint64(company.Size())),
		w.partner.path,
		humanize.Bytes(uint64(partner.Size())),
		w.cnae.path,
		humanize.Bytes(uint64(cnae.Size())),
	)
	return nil
}

// Parse the downloaded files and saves a compressed CSV version of them.
func Parse(dir string) {
	w, err := newWriters(dir)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	var wg sync.WaitGroup
	c := make(chan parsedLine)
	for i := 1; i >= 1; i++ { // infinite loop: breaks when file does not exist
		z, err := newZippedFile(dir, i)
		if os.IsNotExist(err) {
			break // no more files to read
		}
		if err != nil {
			panic(err)
		}
		go parseZipFile(&wg, c, z)
	}

	go func(wg *sync.WaitGroup) {
		wg.Wait()
		close(c)
	}(&wg)

	// show the status (progress)
	var r, s int
	go func() {
		for {
			status(w, r, s)
			time.Sleep(3 * time.Second)
		}
	}()

	for l := range c {
		switch l.kind {
		case "empresa":
			w.company.write(l.contents)
		case "socio":
			w.partner.write(l.contents)
		case "cnae":
			w.cnae.write(l.contents)
		}
		r++
		s += len(l.contents)
	}
}
