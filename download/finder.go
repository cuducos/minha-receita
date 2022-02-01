package download

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const federalRevenue = "https://www.gov.br/receitafederal/pt-br/assuntos/orientacao-tributaria/cadastros/consultas/dados-publicos-cnpj"

var reUpateDate = regexp.MustCompile(`(?i)data da última extração:.*([0-9]{2}/[0-9]{2}/[0-9]{4})`)

type file struct {
	url  string
	path string
}

func getHTMLDocument(client *http.Client, src string) (*goquery.Document, error) {
	r, err := client.Get(src)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s responded with %s", src, r.Status)
	}

	return goquery.NewDocumentFromReader(r.Body)
}

func getLastUpdate(doc *goquery.Document) []string {
	var updateDates []string

	doc.Find("#parent-fieldname-text").Each(func(_ int, p *goquery.Selection) {
		m := reUpateDate.FindAllStringSubmatch(p.Text(), -1)
		if len(m) > 0 {
			for _, md := range m {
				// Transform BR date to intl date

				updateDates = append(updateDates, md[1])
			}
		}
	})
	return updateDates
}

func getURLs(d *goquery.Document) ([]string, error) {

	var urls []string
	d.Find("a.external-link").Each(func(_ int, a *goquery.Selection) {
		h, exist := a.Attr("href")
		if !exist {
			return
		}
		if strings.HasSuffix(h, ".zip") {
			h = strings.ReplaceAll(h, "http//", "")
			h = strings.ReplaceAll(h, "http://http://", "http://")
			urls = append(urls, h)
		}
	})
	return urls, nil
}

func getFiles(d *goquery.Document, dir string, skip bool) ([]file, error) {
	var fs []file
	urls, err := getURLs(d)
	if err != nil {
		return fs, err
	}
	for _, u := range urls {
		f := file{url: u, path: filepath.Join(dir, u[strings.LastIndex(u, "/")+1:])}
		h, err := os.Open(f.path)
		if !skip || errors.Is(err, os.ErrNotExist) {
			fs = append(fs, f)
			continue
		}
		if err == nil {
			h.Close()
		}
	}
	return fs, nil
}
