package download

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/schollz/progressbar/v3"
)

const federalRevenue = "https://www.gov.br/receitafederal/pt-br/assuntos/orientacao-tributaria/cadastros/consultas/dados-publicos-cnpj"

type file struct {
	url   string
	path  string
	extra bool // extra file (not from the Federal Revenue)
}

type size struct {
	size int
	err  error
}

func getUrls() ([]string, error) {
	u := []string{}
	d, err := goquery.NewDocument(federalRevenue)
	if err != nil {
		return u, err
	}

	d.Find("a.external-link").Each(func(_ int, a *goquery.Selection) {
		h, exist := a.Attr("href")
		if !exist {
			return
		}

		if strings.HasSuffix(h, ".zip") {
			h = strings.Replace(h, "http//", "", -1) // fix error in HTML
			u = append(u, h)
		}
	})
	return u, nil
}

func getFiles(dir string) ([]file, error) {
	fs := []file{{
		url:   "https://cnae.ibge.gov.br/images/concla/documentacao/CNAE_Subclasses_2_3_Estrutura_Detalhada.xlsx",
		path:  filepath.Join(dir, "CNAE_Subclasses_2_3_Estrutura_Detalhada.xlsx"),
		extra: true,
	}}

	us, err := getUrls()
	if err != nil {
		return fs, err
	}

	for _, u := range us {
		fs = append(fs, file{url: u, path: filepath.Join(dir, u[strings.LastIndex(u, "/")+1:])})
	}
	return fs, nil
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

	if r.StatusCode != http.StatusOK {
		c <- fmt.Errorf("%s returned HTTP %s", f.url, r.Status)
		return
	}

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
func Download(dir string) {
	log.Output(2, "Preparing to downlaod from the Federal Revenue official website…")

	fs, err := getFiles(dir)
	if err != nil {
		panic(err)
	}

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
