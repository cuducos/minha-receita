package download

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	federalRevenue          = "https://www.gov.br/receitafederal/pt-br/assuntos/orientacao-tributaria/cadastros/consultas/dados-publicos-cnpj"
	federalRevenueSelector  = "a.external-link"
	federalRevenueExtension = ".zip"

	nationalTreasure          = "https://www.tesourotransparente.gov.br/ckan/dataset/lista-de-municipios-do-siafi/resource/eebb3bc6-9eea-4496-8bcf-304f33155282"
	nationalTreasureSelector  = "a.btn"
	nationalTreasureExtension = ".CSV"
)

type search struct {
	name      string
	url       string
	selector  string
	extension string
}

type file struct {
	url  string
	path string
}

func getURLs(client *http.Client, s search) ([]string, error) {
	r, err := client.Get(s.url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s responded with %s", s.url, r.Status)
	}
	d, err := goquery.NewDocumentFromReader(r.Body)
	if err != nil {
		return nil, err
	}
	urls := make(map[string]struct{})
	d.Find(s.selector).Each(func(_ int, a *goquery.Selection) {
		h, exist := a.Attr("href")
		if !exist {
			return
		}
		if strings.HasSuffix(h, s.extension) {
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

func appendFiles(fs *[]file, client *http.Client, s search, dir string, skip bool) error {
	urls, err := getURLs(client, s)
	if err != nil {
		return fmt.Errorf("error listing files from the %s: %w", s.url, err)
	}
	for _, u := range urls {
		f := file{url: u, path: filepath.Join(dir, u[strings.LastIndex(u, "/")+1:])}
		h, err := os.Open(f.path)
		if !skip || errors.Is(err, os.ErrNotExist) {
			*fs = append(*fs, f)
			continue
		}
		if err == nil {
			h.Close()
		}
	}
	return nil
}

func getFiles(client *http.Client, searches []search, dir string, skip bool) ([]file, error) {
	var fs []file
	for _, s := range searches {
		if err := appendFiles(&fs, client, s, dir, skip); err != nil {
			return []file{}, fmt.Errorf("error getting files from the %s: %w", s.name, err)
		}
	}
	return fs, nil
}
