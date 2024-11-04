package download

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

const (
	userAgent = "Minha Receita/0.0.1 (minhareceita.org)"

	// FederalRevenueUpdatedAt is a file that contains the date the data was
	// extracted by the Federal Revenue
	FederalRevenueUpdatedAt = "updated_at.txt"

	// Metadata source
	federalRevenueMetadataURL     = "https://dados.gov.br/api/publico/conjuntos-dados/cadastro-nacional-da-pessoa-juridica---cnpj"
	federalRevenueDateFormat      = "02/01/2006 15:04:05"
	federalRevenueDateFormatNotes = "02/01/2006"

	// Zipped CSV source
	federalRevenueURL = "https://arquivos.receitafederal.gov.br/cnpj/dados_abertos_cnpj"
)

var datePattern = regexp.MustCompile(`Data da última extração:? +(?P<updatedAt>\d{2}/\d{2}/\d{4})`)
var yearMonthPattern = regexp.MustCompile(`href="(\d{4}-\d{2}/)"`)
var filePattern = regexp.MustCompile(`href="(\w+\d?\.zip)"`)

func httpGet(url string) (string, error) {
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
	b, err := httpGet(url)
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
	return url + "/" + bs[len(bs)-1], nil
}

func federalRevenueGetURLs(url string) ([]string, error) {
	u, err := federalRevenueGetMostRecentURL(url)
	if err != nil {
		return nil, fmt.Errorf("could not read %s response body: %w", url, err)
	}
	b, err := httpGet(u)
	if err != nil {
		return nil, fmt.Errorf("error getting %s: %w", url, err)
	}
	var urls []string
	for _, m := range filePattern.FindAllStringSubmatch(b, -1) {
		urls = append(urls, u+m[1])
	}
	return urls, nil
}

type federalRevenueTime struct{ Time time.Time }

func (t *federalRevenueTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		t.Time = time.Time{}
		return nil
	}
	var err error
	t.Time, err = time.Parse(federalRevenueDateFormat, s)
	if err != nil {
		return fmt.Errorf("could not parse date/time %s as %s: %w", s, federalRevenueDateFormat, err)
	}
	return nil
}

type federalRevenueMetadataResource struct {
	Format           string             `json:"format"`
	URL              string             `json:"url"`
	MetadataModified federalRevenueTime `json:"metadata_modified"`
}

type federalRevenueMetadata struct {
	Resources []federalRevenueMetadataResource `json:"resources"`
	Notes     string                           `json:"notes"`
}

func (r *federalRevenueMetadata) updatedAt() (t time.Time) {
	m := datePattern.FindStringSubmatch(r.Notes)
	if len(m) == 2 {
		t, err := time.Parse(federalRevenueDateFormatNotes, m[1])
		if err == nil {
			return t
		}
	}

	for _, v := range r.Resources {
		if t.Before(v.MetadataModified.Time) {
			t = v.MetadataModified.Time
		}
	}
	return t
}

func newFederalRevenueMetadata(url string) (*federalRevenueMetadata, error) {
	c := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request %s: %w", url, err)
	}
	req.Header.Set("User-Agent", userAgent)
	r, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error getting %s: %w", url, err)
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s responded with %s", url, r.Status)
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read %s response body: %w", url, err)
	}
	var data federalRevenueMetadata
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, fmt.Errorf("could not unmarshal %s json response: %w", url, err)
	}
	return &data, nil
}

func federalRevenueGetMetadata(url, dir string) error {
	data, err := newFederalRevenueMetadata(url)
	if err != nil {
		return fmt.Errorf("error getting federal revenue data: %w", err)
	}
	if err := saveUpdatedAt(dir, data.updatedAt()); err != nil {
		return fmt.Errorf("could not save the update at date: %w", err)

	}
	return nil
}

func fetchUpdatedAt(url string) (string, error) {
	data, err := newFederalRevenueMetadata(url)
	if err != nil {
		return "", fmt.Errorf("error getting federal revenue data: %w", err)
	}
	return data.updatedAt().Format("2006-01-02"), nil
}

func hasUpdate(url, dir string) (bool, error) {
	dt, err := fetchUpdatedAt(url)
	if err != nil {
		return false, fmt.Errorf("error getting federal revenue updated at: %w", err)
	}
	pth := filepath.Join(dir, FederalRevenueUpdatedAt)
	f, err := os.Open(pth)
	if err != nil {
		return false, fmt.Errorf("error opening %s: %w", pth, err)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return false, fmt.Errorf("error reading %s: %w", pth, err)
	}
	fmt.Printf("Local files\t%s\n", string(b))
	fmt.Printf("Remote files\t%s\n", dt)
	return string(b) != dt, nil
}

func saveUpdatedAt(dir string, u time.Time) error {
	pth := filepath.Join(dir, FederalRevenueUpdatedAt)
	f, err := os.Create(pth)
	if err != nil {
		return fmt.Errorf("could not create %s: %w", pth, err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	_, err = w.WriteString(u.Format("2006-01-02"))
	if err != nil {
		return fmt.Errorf("could not write to %s: %w", pth, err)
	}
	w.Flush()
	return nil
}
