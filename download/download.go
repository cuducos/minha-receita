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
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/schollz/progressbar/v3"
)

const federalRevenue = "https://www.gov.br/receitafederal/pt-br/assuntos/orientacao-tributaria/cadastros/consultas/dados-publicos-cnpj"
const retries = 10

type file struct {
	url  string
	path string
}

func getURLs(client *http.Client, src string) ([]string, error) {
	r, err := client.Get(src)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s responded with %s", src, r.Status)
	}

	d, err := goquery.NewDocumentFromReader(r.Body)
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
			urls = append(urls, strings.ReplaceAll(h, "http//", ""))
		}
	})
	return urls, nil
}

func getFiles(client *http.Client, src, dir string) ([]file, error) {
	var fs []file
	urls, err := getURLs(client, src)
	if err != nil {
		return fs, err
	}
	for _, u := range urls {
		fs = append(fs, file{url: u, path: filepath.Join(dir, u[strings.LastIndex(u, "/")+1:])})
	}
	return fs, nil
}

type downloader struct {
	files     []file
	client    *http.Client
	totalSize int64
	bar       *progressbar.ProgressBar
}

type size struct {
	size int64
	err  error
}

func (d *downloader) getSize(ch chan<- size, url string) {
	r, err := d.client.Head(url)
	if err != nil {
		ch <- size{err: fmt.Errorf("Error sending a HTTP HEAD request to %s: %s", url, err)}
		return
	}
	defer r.Body.Close()

	if r.ContentLength == 0 {
		ch <- size{err: fmt.Errorf("Could not get size for %s", url)}
		return
	}

	ch <- size{size: r.ContentLength}
}

func (d *downloader) getTotalSize() error {
	d.totalSize = 0
	q := make(chan size)
	for _, f := range d.files {
		go d.getSize(q, f.url)
	}
	for range d.files {
		s := <-q
		if s.err != nil {
			return s.err
		}
		d.totalSize += s.size
	}
	return nil
}

func (d *downloader) setProgressBar() {
	d.bar = progressbar.DefaultBytes(d.totalSize, "Downloading")
}

func (d *downloader) download(ch chan<- error, f file, a int) {
	err := func(f file) error {
		r, err := d.client.Get(f.url)
		if err != nil {
			log.Output(2, fmt.Sprintf("HTTP request to %s failed: %v", f.url, err))
			return err
		}
		defer r.Body.Close()

		if r.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP request to %s got %s", f.url, r.Status)
		}

		var h *os.File
		h, err = os.Create(f.path)
		if err != nil {
			return fmt.Errorf("Failed to create %s: %v", f.path, err)
		}
		defer h.Close()

		_, err = io.Copy(io.MultiWriter(h, d.bar), r.Body)
		if err != nil {
			return fmt.Errorf("Error downloading %s: %v", f.url, err)
		}
		return nil
	}(f)

	if err != nil {
		if a < retries {
			time.Sleep(time.Duration(int(math.Pow(float64(2), float64(a)))) * time.Second)
			d.download(ch, f, a+1)
			return
		}
		err = fmt.Errorf("After %d attempts, could not download %s: %c", retries, f.url, err)
		ch <- err
		return
	}
	ch <- nil
}

func (d *downloader) downloadAll() error {
	q := make(chan error)
	for _, f := range d.files {
		go d.download(q, f, 0)
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

func newDownloader(c *http.Client, fs []file) (*downloader, error) {
	d := downloader{files: fs, client: c}
	err := d.getTotalSize()
	if err != nil {
		return nil, err
	}
	d.setProgressBar()
	return &d, nil
}

// Download all the files (might take several minutes).
func Download(dir string, timeout time.Duration, urlsOnly bool) error {
	c := &http.Client{Timeout: timeout}
	if !urlsOnly {
		log.Output(2, "Preparing to download from the Federal Revenue official website…")
	}

	fs, err := getFiles(c, federalRevenue, dir)
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

	d, err := newDownloader(c, fs)
	if err != nil {
		return err
	}
	return d.downloadAll()
}
