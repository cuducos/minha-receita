package download

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	federalRevenueURL       = "https://www.gov.br/receitafederal/pt-br/assuntos/orientacao-tributaria/cadastros/consultas/dados-publicos-cnpj"
	federalRevenueSelector  = "a.external-link"
	federalRevenueExtension = ".zip"
)

func federalRevenueGetURLs(client *http.Client, url string) ([]string, error) {
	r, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting %s: %w", federalRevenueURL, err)
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s responded with %s", federalRevenueURL, r.Status)
	}
	d, err := goquery.NewDocumentFromReader(r.Body)
	if err != nil {
		return nil, err
	}
	urls := make(map[string]struct{})
	d.Find(federalRevenueSelector).Each(func(_ int, a *goquery.Selection) {
		h, exist := a.Attr("href")
		if !exist {
			return
		}
		if strings.HasSuffix(h, federalRevenueExtension) {
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
