package download

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
)

const (
	userAgent = "Minha Receita/0.0.1 (minhareceita.org)"

	// FederalRevenueUpdatedAt is a file that contains the date the data was
	// extracted by the Federal Revenue
	FederalRevenueUpdatedAt = "updated_at.txt"

	// Zipped CSV source
	federalRevenueURL        = "https://arquivos.receitafederal.gov.br/dados/cnpj/"
	federalRevenueSourcePath = "dados_abertos_cnpj"
	federalRevenueTaxesPath  = "regime_tributario"
	federalRevenueDateFormat = "02/01/2006 15:04"
)

var fileTimestampPattern = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
var yearMonthPattern = regexp.MustCompile(`href="(\d{4}-\d{2}/)"`)
var filePattern = regexp.MustCompile(`href="(\w+\d?\.zip)"`)
var taxFilePattern = regexp.MustCompile(`href="((Imune|Lucro).+\.zip)"`)

func get(url string) (string, error) {
	c := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request %s: %w", url, err)
	}
	req.Header.Set("User-Agent", userAgent)
	r, err := c.Do(req)
	if err != nil {
		return "", fmt.Errorf("error getting %s: %w", url, err)
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s responded with %s", url, r.Status)
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return "", fmt.Errorf("could not read %s response body: %w", url, err)
	}
	return string(b), nil
}

func federalRevenueGetMostRecentURL(url string) (string, error) {
	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}
	b, err := get(url)
	if err != nil {
		return "", fmt.Errorf("error getting %s: %w", url, err)
	}
	var bs []string
	for _, m := range yearMonthPattern.FindAllStringSubmatch(b, -1) {
		bs = append(bs, m[1])
	}
	slices.Sort(bs)
	if len(bs) == 0 {
		return "", fmt.Errorf("no batches found in %s", url)
	}
	return url + bs[len(bs)-1], nil
}

func taxRegimeGetURLs(url string) ([]string, error) {
	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}
	b, err := get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting %s: %w", url, err)
	}
	var urls []string
	for _, m := range taxFilePattern.FindAllStringSubmatch(b, -1) {
		urls = append(urls, url+m[1])
	}
	return urls, nil
}

func federalRevenueGetURLs(url string) ([]string, error) {
	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}
	u, err := federalRevenueGetMostRecentURL(url + federalRevenueSourcePath)
	if err != nil {
		return nil, fmt.Errorf("could not read %s response body: %w", url, err)
	}
	b, err := get(u)
	if err != nil {
		return nil, fmt.Errorf("error getting %s: %w", url, err)
	}
	var urls []string
	for _, m := range filePattern.FindAllStringSubmatch(b, -1) {
		urls = append(urls, u+m[1])
	}
	ts, err := taxRegimeGetURLs(url + federalRevenueTaxesPath)
	if err != nil {
		return nil, fmt.Errorf("error getting taxe regime urls: %w", err)
	}
	urls = append(urls, ts...)
	return urls, nil
}

func saveUpdatedAt(dir string) error {
	u := federalRevenueURL + federalRevenueSourcePath
	m, err := federalRevenueGetMostRecentURL(u)
	if err != nil {
		return fmt.Errorf("error getting most recent source url: %w", err)
	}
	b, err := get(m)
	if err != nil {
		return fmt.Errorf("error getting contents of the most recent source: %w", err)
	}
	ds := fileTimestampPattern.FindAllString(b, -1)
	if len(ds) < 1 {
		return fmt.Errorf("could not find updated at date in %s", u)
	}
	sort.Strings(ds)
	d := ds[len(ds)-1]
	pth := filepath.Join(dir, FederalRevenueUpdatedAt)
	f, err := os.Create(pth)
	if err != nil {
		return fmt.Errorf("error creating %s: %w", pth, err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	_, err = w.WriteString(d)
	if err != nil {
		return fmt.Errorf("error writing %s: %w", pth, err)
	}
	w.Flush()
	return nil
}
