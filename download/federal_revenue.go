package download

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	federalRevenueURL       = "https://www.gov.br/receitafederal/pt-br/assuntos/orientacao-tributaria/cadastros/consultas/dados-publicos-cnpj"
	federalRevenueSelector  = "a.external-link"
	federalRevenueExtension = ".zip"
	federalRevenueUpdatedAt = "updated_at.txt"
)

var updatedAtRegex = regexp.MustCompile(`(?i)data da última extração:\s*(\d{2}/\d{2}/\d{4})`)

func federalRevenueGetURLs(client *http.Client, url, dir string) ([]string, error) {
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
	if err := saveUpdatedAt(dir, d); err != nil {
		return nil, fmt.Errorf("could not save the update at date: %w", err)
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

func saveUpdatedAt(dir string, dom *goquery.Document) error {
	b := dom.Find("body").First()
	m := updatedAtRegex.FindAllStringSubmatch(b.Text(), -1)
	if len(m) == 0 {
		return fmt.Errorf("cannot find date in %s", dom.Url.RequestURI())
	}
	t, err := time.Parse("02/01/2006", m[0][1])
	if err != nil {
		return fmt.Errorf("cannot parse date %s: %w", m[0][1], err)
	}
	pth := filepath.Join(dir, federalRevenueUpdatedAt)
	f, err := os.Create(pth)
	if err != nil {
		return fmt.Errorf("could not create %s: %w", pth, err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	_, err = w.WriteString(t.Format("2006-01-02"))
	if err != nil {
		return fmt.Errorf("could not write to %s: %w", pth, err)
	}
	w.Flush()
	return nil
}
