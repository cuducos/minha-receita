package download

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/PuerkitoBio/goquery"
)

type lastUpdate struct {
	Companies string `json:"companies"`
	Taxes     string `json:"taxes"`
}

const lastUpdateOutputFile = "last_update.json"

func createLastUpdateJSONFile(dir string, doc *goquery.Document) error {
	dates, err := getLastUpdate(doc)
	if err != nil {
		return fmt.Errorf("error getting last update dates from HTML document: %w", err)
	}
	// Create last_update.txt
	pth := filepath.Join(dir, lastUpdateOutputFile)
	f, err := os.Create(pth)
	if err != nil {
		return fmt.Errorf("error creating last update file at %s: %w", pth, err)
	}
	defer f.Close()

	body, err := json.MarshalIndent(&dates, "", "    ")
	if err != nil {
		return fmt.Errorf("error encoding last update struct to json: %w", err)
	}
	_, err = f.Write(body)
	if err != nil {
		return fmt.Errorf("error writing last update file at %s: %w", pth, err)
	}
	return nil
}
