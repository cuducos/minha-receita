package download

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateLastUpdateFile(t *testing.T) {
	ts := httpTestServer(t)
	defer ts.Close()

	doc, err := getHTMLDocument(ts.Client(), ts.URL)
	if err != nil {
		t.Errorf("expected getHTMLDocument to run without errors, got: %v:", err)
		return
	}

	tmp := t.TempDir()

	err = createLastUpdateJSONFile(tmp, doc)
	if err != nil {
		t.Errorf("expected no errors on create last update file, got: %v", err)
	}

	var lus lastUpdate
	pth := filepath.Join(tmp, lastUpdateOutputFile)
	f, err := os.Open(pth)
	if err != nil {
		t.Errorf("could not open %s, got: %v", pth, err)
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&lus)
	if err != nil {
		t.Errorf("could not decode %s as JSON, got: %v", pth, err)
	}

	expected := []string{"2021-01-08", "2021-10-20"}
	if lus.Companies != expected[0] {
		t.Errorf("expected %s for companies, got %s", expected[0], lus.Companies)
	}
	if lus.Taxes != expected[1] {
		t.Errorf("expected %s for taxes, got %s", expected[1], lus.Taxes)
	}
}
