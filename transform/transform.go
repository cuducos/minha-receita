package transform

import (
	"fmt"
)

const maxFilesOpened = 512 // TODO how to optimize this number?

// Transform the downloaded files for company venues creating a JSON file per CNPJ
func Transform(dir string) error {
	t, err := newTask(dir, venue)
	if err != nil {
		return fmt.Errorf("error creating new task for %s in %s: %w", string(venue), dir, err)
	}
	return t.run(maxFilesOpened)
}
