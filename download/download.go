package download

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/schollz/progressbar/v3"
)

const federalRevenue = "https://www.gov.br/receitafederal/pt-br/assuntos/orientacao-tributaria/cadastros/consultas/dados-publicos-cnpj"
const listOfCNAE = "https://cnae.ibge.gov.br/images/concla/documentacao/CNAE_Subclasses_2_3_Estrutura_Detalhada.xlsx"
const retries = 10

type file struct {
	url  string
	path string
}

func getURLs(src string) ([]string, error) {
	d, err := goquery.NewDocument(src)
	if err != nil {
		return nil, err
	}

	var urls []string
	d.Find("a.external-link").Each(func(_ int, a *goquery.Selection) {
		h, exist := a.Attr("href")
		if !exist {
			return
		}
		if strings.HasSuffix(h, ".zip") {
			urls = append(urls, h)
		}
	})
	return urls, nil
}

func getFiles(dir string) ([]file, error) {
	fs := []file{{
		url:  listOfCNAE,
		path: filepath.Join(dir, "CNAE_Subclasses_2_3_Estrutura_Detalhada.xlsx"),
	}}

	urls, err := getURLs(federalRevenue)
	if err != nil {
		return fs, err
	}

	for _, u := range urls {
		fs = append(fs, file{url: u, path: filepath.Join(dir, u[strings.LastIndex(u, "/")+1:])})
	}
	return fs, nil
}

type Downloader struct {
	files     []file
	client    *http.Client
	totalSize int64
	bar       *progressbar.ProgressBar
}

func (d *Downloader) getSize(ch chan<- error, url string) {
	// IBGE does not respond to HTTP HEAD as expected
	if url == listOfCNAE {
		d.totalSize += 137216
		ch <- nil
		return
	}

	r, err := d.client.Head(url)
	if err != nil {
		ch <- err
		return
	}
	for _, k := range []string{"Content-Length", "content-length"} {
		var s int
		s, err = strconv.Atoi(r.Header.Get(k))
		if err == nil {
			d.totalSize += int64(s)
			ch <- nil
			return
		}
	}
	err = fmt.Errorf("Could not get size for %s", url)
	ch <- err
}

func (d *Downloader) getTotalSize() error {
	d.totalSize = 0
	q := make(chan error)
	for _, f := range d.files {
		go d.getSize(q, f.url)
	}
	for range d.files {
		err := <-q
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Downloader) setProgressBar() {
	d.bar = progressbar.DefaultBytes(d.totalSize, "Downloading")
}

func (d *Downloader) download(ch chan<- error, f *file, a int) {
	r, err := d.client.Get(f.url)
	if err != nil {
		log.Output(2, fmt.Sprintf("HTTP request to %s failed", f.url))
		ch <- err
		return
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		if a < retries {
			time.Sleep(time.Duration(int(math.Pow(float64(2), float64(a)))) * time.Second)
			d.download(ch, f, a+1)
			return
		} else {
			err = fmt.Errorf("After %d attempts, could not download %s (%s)", retries, f.url, r.Status)
			ch <- err
			return
		}
	}

	var h *os.File
	h, err = os.Create(f.path)
	if err != nil {
		log.Output(2, fmt.Sprintf("Failed to create %s", f.path))
		ch <- err
		return
	}
	defer h.Close()

	_, err = io.Copy(io.MultiWriter(h, d.bar), r.Body)
	if err != nil {
		log.Output(2, fmt.Sprintf("Error downloading %s", f.url))
		ch <- err
		return
	}
	ch <- nil
}

func (d *Downloader) downloadAll() error {
	q := make(chan error)
	for _, f := range d.files {
		go d.download(q, &f, 0)
	}
	for range d.files {
		err := <-q
		if err != nil {
			return err
		}
	}
	d.bar.Finish()
	return nil
}

func NewDownloader(c *http.Client, fs []file) (*Downloader, error) {
	d := Downloader{files: fs, client: c}
	err := d.getTotalSize()
	if err != nil {
		return nil, err
	}
	d.setProgressBar()
	return &d, nil
}

// Download all the files (might take several minutes).
func Download(dir string, timeout time.Duration, urlsOnly bool) error {
	if !urlsOnly {
		log.Output(2, "Preparing to download from the Federal Revenue official website…")
	}
	fs, err := getFiles(dir)
	if err != nil {
		return err
	}
	if urlsOnly {
		urls := make([]string, 0, len(fs))
		for _, f := range fs {
			urls = append(urls, f.url)
		}
		sort.Strings(urls)
		for _, u := range urls {
			fmt.Println(u)
		}
		return nil
	}
	c := &http.Client{Timeout: timeout}
	d, err := NewDownloader(c, fs)
	if err != nil {
		return err
	}
	return d.downloadAll()
}
