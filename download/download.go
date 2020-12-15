package download

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/schollz/progressbar/v3"
)

const FilePattern = "DADOS_ABERTOS_CNPJ_%02d.zip"
const federalRevenue = "http://200.152.38.155/CNPJ/"
const brasilIO = "https://data.brasil.io/mirror/socios-brasil/"
const files = 20

type file struct {
	url   string
	path  string
	extra bool // extra file (not from the Federal Revenue)
}

type size struct {
	size int
	err  error
}

func getFiles(m bool, dir string) []file {
	fs := []file{{
		url:   "https://cnae.ibge.gov.br/images/concla/documentacao/CNAE_Subclasses_2_3_Estrutura_Detalhada.xlsx",
		path:  filepath.Join(dir, "CNAE_Subclasses_2_3_Estrutura_Detalhada.xlsx"),
		extra: true,
	}}

	var s string
	if m {
		s = brasilIO
	} else {
		s = federalRevenue
	}
	for i := 1; i <= files; i++ {
		n := fmt.Sprintf(FilePattern, i)
		fs = append(fs, file{url: fmt.Sprintf("%s%s", s, n), path: filepath.Join(dir, n)})
	}
	return fs
}

func getSize(c chan<- size, url string) {
	var size size
	var r *http.Response
	r, size.err = http.Head(url)
	if size.err != nil {
		c <- size
		return
	}

	for _, k := range []string{"Content-Length", "content-length"} {
		size.size, size.err = strconv.Atoi(r.Header.Get(k))
		if size.err == nil {
			c <- size
			return
		}
	}

	size.err = fmt.Errorf("Could not get size for %s", url)
	c <- size
	return
}

func download(c chan<- error, b *progressbar.ProgressBar, f file) {
	r, err := http.Get(f.url)
	if err != nil {
		c <- err
		return
	}
	defer r.Body.Close()

	h, err := os.Create(f.path)
	if err != nil {
		c <- err
		return
	}
	defer h.Close()

	if b != nil {
		_, err = io.Copy(io.MultiWriter(h, b), r.Body)
	} else {
		_, err = io.Copy(h, r.Body)
	}
	if err != nil {
		c <- err
	}
	c <- nil
}

func getTotalSize(fs []file) (int64, error) {
	var t int64
	q := make(chan size)
	for _, f := range fs {
		if f.extra {
			continue
		}
		go getSize(q, f.url)
	}
	for _, f := range fs {
		if f.extra {
			continue
		}
		r := <-q
		if r.err != nil {
			return 0, r.err
		}
		t += int64(r.size)
	}
	return t, nil
}

// Download all the files (might take several minutes).
func Download(m bool, dir string) {
	var msg string
	if m {
		msg = "Preparing to downlaod from Brasil.IO mirror…"
	} else {
		msg = "Preparing to downlaod from the Federal Revenue official website…"
	}
	log.Output(2, msg)

	fs := getFiles(m, dir)
	t, err := getTotalSize(fs)
	if err != nil {
		panic(err)
	}

	q := make(chan error)
	bar := progressbar.DefaultBytes(t, "Downloading")
	for _, f := range fs {
		if f.extra {
			go download(q, nil, f)
		} else {
			go download(q, bar, f)
		}
	}
	for range fs {
		err := <-q
		if err != nil {
			panic(err)
		}
	}
	bar.Finish()
}
