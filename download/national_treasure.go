package download

import (
	"net/http"
)

const (
	nationalTreasureBaseURL = "https://www.tesourotransparente.gov.br"
	nationalTreasurePkgID   = "abb968cb-3710-4f85-89cf-875c91b9c7f6"
)

func nationalTreasureGetURLs(client *http.Client, baseURL string) ([]string, error) {
	return ckanGetURLS(client, baseURL, nationalTreasurePkgID)
}
