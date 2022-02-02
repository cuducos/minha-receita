package transform

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cuducos/minha-receita/db"
)

var (
	testdata = filepath.Join("..", "testdata")
)

func companyFromString(j string) (company, error) {
	var c company
	if err := json.Unmarshal([]byte(j), &c); err != nil {
		return company{}, fmt.Errorf("error unmarshaling: %w", err)
	}
	return c, nil
}

func newTestDB(t *testing.T) *db.PostgreSQL {
	u := os.Getenv("TEST_POSTGRES_URI")
	if u == "" {
		t.Errorf("expected a posgres uri at TEST_POSTGRES_URI, found nothing")
		return nil
	}
	r, err := db.NewPostgreSQL(u, "public")
	if err != nil {
		t.Errorf("expected no error creating a test database, got %s", err)
		return nil
	}
	if err := r.DropTable(); err != nil {
		t.Errorf("expected no error droping the table in the test database, got %s", err)
		return nil
	}
	if err := r.CreateTable(); err != nil {
		t.Errorf("expected no error creating the table in the test database, got %s", err)
		return nil
	}
	return &r
}
