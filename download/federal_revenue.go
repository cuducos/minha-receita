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
	"strings"
	"time"
)

const (
	// FederalRevenueUpdatedAt is a file that contains the date the data was
	// extracted by the Federal Revenue
	FederalRevenueUpdatedAt = "updated_at.txt"

	federalRevenueURL             = "https://dados.gov.br/api/publico/conjuntos-dados/cadastro-nacional-da-pessoa-juridica---cnpj"
	federalRevenueFormat          = "zip+csv"
	federalRevenueDateFormat      = "02/01/2006 15:04:05"
	federalRevenueDateFormatNotes = "02/01/2006"

	userAgent = "Minha Receita/0.0.1 (minhareceita.org)"
)

var datePattern = regexp.MustCompile(`Data da última extração:? +(?P<updatedAt>\d{2}/\d{2}/\d{4})`)

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

type federalRevenueResource struct {
	Format           string             `json:"format"`
	URL              string             `json:"url"`
	MetadataModified federalRevenueTime `json:"metadata_modified"`
}

type federalRevenueResponse struct {
	Resources []federalRevenueResource `json:"resources"`
	Notes     string                   `json:"notes"`
}

func (r *federalRevenueResponse) updatedAt() (t time.Time) {
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

func newFederalRevenueResponse(url string) (*federalRevenueResponse, error) {
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
	var data federalRevenueResponse
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, fmt.Errorf("could not unmarshal %s json response: %w", url, err)
	}
	return &data, nil
}

func federalRevenueGetURLsBase(url, dir string, updatedAt bool) ([]string, error) {
	data, err := newFederalRevenueResponse(url)
	if err != nil {
		return nil, fmt.Errorf("error getting federal revenue data: %w", err)
	}
	var u []string
	for _, v := range data.Resources {
		if v.Format == federalRevenueFormat {
			u = append(u, v.URL)
		}
	}
	if updatedAt {
		if err := saveUpdatedAt(dir, data.updatedAt()); err != nil {
			return nil, fmt.Errorf("could not save the update at date: %w", err)
		}
	}
	return u, nil
}

func federalRevenueGetURLs(url, dir string) ([]string, error) {
	return federalRevenueGetURLsBase(url, dir, true)
}

func federalRevenueGetURLsNoUpdatedAt(url, dir string) ([]string, error) {
	return federalRevenueGetURLsBase(url, dir, false)
}

func fetchUpdatedAt(url string) (string, error) {
	data, err := newFederalRevenueResponse(url)
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
