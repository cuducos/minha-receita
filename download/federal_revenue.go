package download

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// FederalRevenueUpdatedAt is a file that contains the date the data was
	// extracted by the Federal Revenue
	FederalRevenueUpdatedAt = "updated_at.txt"

	federalRevenueURL        = "https://dados.gov.br/api/publico/conjuntos-dados/cadastro-nacional-da-pessoa-juridica-cnpj"
	federalRevenueFormat     = "zip+csv"
	federalRevenueDateFormat = "02/01/2006 15:04:05"
)

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
}

func federalRevenueGetURLsBase(client *http.Client, url, dir string, updatedAt bool) ([]string, error) {
	r, err := client.Get(url)
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
	var u []string
	var t time.Time
	for _, v := range data.Resources {
		if v.Format == federalRevenueFormat {
			u = append(u, v.URL)
		}
		if t.Before(v.MetadataModified.Time) {
			t = v.MetadataModified.Time
		}
	}
	if updatedAt {
		if err := saveUpdatedAt(dir, t); err != nil {
			return nil, fmt.Errorf("could not save the update at date: %w", err)
		}
	}
	return u, nil
}

func federalRevenueGetURLs(client *http.Client, url, dir string) ([]string, error) {
	return federalRevenueGetURLsBase(client, url, dir, true)
}

func federalRevenueGetURLsNoUpdatedAt(client *http.Client, url, dir string) ([]string, error) {
	return federalRevenueGetURLsBase(client, url, dir, false)
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
