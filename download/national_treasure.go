package download

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	ckanPkgPath             = "/ckan/api/3/action/package_show?id="
	nationalTreasureBaseURL = "https://www.tesourotransparente.gov.br"
	nationalTreasurePkgID   = "abb968cb-3710-4f85-89cf-875c91b9c7f6"
)

type ckanResource struct {
	URL string
}

type ckanResult struct {
	Resources []ckanResource
}

type ckanPkg struct {
	Success bool
	Result  ckanResult
}

func ckanGetURLS(baseURL, pkgID string) ([]string, error) {
	url := strings.Join([]string{baseURL, ckanPkgPath, pkgID}, "")
	r, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting %s: %w", url, err)
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s responded with %s", url, r.Status)
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response from %s: %w", url, err)
	}
	var pkg ckanPkg
	if err = json.Unmarshal(b, &pkg); err != nil {
		return nil, fmt.Errorf("error unmarshilling response from %s: %w", url, err)
	}
	if !pkg.Success {
		return nil, fmt.Errorf("error in ckan api response:\n%s", string(b))
	}
	urls := make([]string, len(pkg.Result.Resources))
	for i, s := range pkg.Result.Resources {
		urls[i] = s.URL
	}
	return urls, nil
}

func nationalTreasureGetURLs(baseURL, dir string) ([]string, error) {
	return ckanGetURLS(baseURL, nationalTreasurePkgID)
}
