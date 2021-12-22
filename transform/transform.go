package transform

import (
	"fmt"
)

const maxFilesOpened = 512 // TODO how to optimize this number?

// Transform the downloaded files for company venues creating a JSON file per CNPJ
func Transform(dir string) error {
	t, err := newTask(dir, venues)
	if err != nil {
		return fmt.Errorf("error creating new task for %s in %s: %w", string(venues), dir, err)
	}
	if err := t.run(maxFilesOpened); err != nil {
		return err
	}
	if err := addPartners(dir, &t.lookups); err != nil {
		return err
	}
	if err := addBaseCPNJ(dir, &t.lookups); err != nil {
		return err
	}
	return nil
}
