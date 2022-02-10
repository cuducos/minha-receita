package download

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const federalRevenue = "https://www.gov.br/receitafederal/pt-br/assuntos/orientacao-tributaria/cadastros/consultas/dados-publicos-cnpj"

var reUpateDate = regexp.MustCompile(`(?i)data da última extração:\s*(\d{2}/\d{2}/\d{4})`)

type file struct {
	url  string
	path string
}

func toRFC3339Date(brDt string) (string, error) {
	dt, err := time.Parse("02/01/2006", brDt)
	if err != nil {
		return "", fmt.Errorf("error parsing date %s: %w", brDt, err)
	}
	return dt.Format("2006-01-02"), nil
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

func getLastUpdate(doc *goquery.Document) (lastUpdate, error) {
	var dates []string
	doc.Find("#parent-fieldname-text").Each(func(_ int, p *goquery.Selection) {
		m := reUpateDate.FindAllStringSubmatch(p.Text(), -1)
		if len(m) > 0 {
			for _, d := range m {
				dates = append(dates, d[1])
			}
		}
	})
	if len(dates) == 0 {
		return lastUpdate{}, fmt.Errorf("no update dates found")
	}
	if len(dates) != 2 {
		return lastUpdate{}, fmt.Errorf("could not find the two expected update dates, found %d instead", len(dates))
	}

	dtC, err := toRFC3339Date(dates[0])
	if err != nil {
		return lastUpdate{}, fmt.Errorf("could not get companies update date: %w", err)
	}
	dtT, err := toRFC3339Date(dates[1])
	if err != nil {
		return lastUpdate{}, fmt.Errorf("could not get companies update date: %w", err)
	}

	return lastUpdate{dtC, dtT}, nil
}

func getURLs(d *goquery.Document) ([]string, error) {

	urls := make(map[string]struct{})
	d.Find("a.external-link").Each(func(_ int, a *goquery.Selection) {
		h, exist := a.Attr("href")
		if !exist {
			return
		}
		if strings.HasSuffix(h, ".zip") {
			h = strings.ReplaceAll(h, "http//", "")
			h = strings.ReplaceAll(h, "http://http://", "http://")
			urls[h] = struct{}{}
		}
	})

	var u []string
	for k := range urls {
		u = append(u, k)
	}
	return u, nil
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
