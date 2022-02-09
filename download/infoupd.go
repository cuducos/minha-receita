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
	dates := getLastUpdate(doc)
	// Create last_update.txt
	pth := filepath.Join(dir, lastUpdateOutputFile)
	f, err := os.Create(pth)
	if err != nil {
		return fmt.Errorf("error creating last update file at %s: %w", pth, err)
	}
	defer f.Close()

	st := lastUpdate{
		Companies: dates[0],
	}
	if len(dates) > 1 {
		st.Taxes = dates[1]
	}

	body, err := json.MarshalIndent(&st, "", "    ")
	if err != nil {
		return fmt.Errorf("error encoding last update struct to json: %w", err)
	}
	_, err = f.Write(body)
	if err != nil {
		return fmt.Errorf("error writing last updte file at %s: %w", pth, err)
	}
	return nil
}
