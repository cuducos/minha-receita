package transform

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/cuducos/minha-receita/download"
)

func saveUpdatedAt(db database, dir string) error {
	log.Output(2, "Saving the updated at date to the database…")
	p := filepath.Join(dir, download.FederalRevenueUpdatedAt)
	v, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", p, err)

	}
	return db.MetaSave("updated-at", string(v))
}
