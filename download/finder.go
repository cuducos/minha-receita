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

const federalRevenue = "https://www.gov.br/receitafederal/pt-br/assuntos/orientacao-tributaria/cadastros/consultas/dados-publicos-cnpj"

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

func getFiles(client *http.Client, src, dir string, skip bool) ([]file, error) {
	var fs []file
	urls, err := getURLs(client, src)
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
