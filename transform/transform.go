package transform

import (
	"fmt"
)

const maxFilesOpened = 512 // TODO how to optimize this number?

// Transform the downloaded files for company venues creating a JSON file per CNPJ
func Transform(srcDir, outDir string) error {
	t, err := newTask(srcDir, outDir)
	if err != nil {
		return fmt.Errorf("error creating new task for venues in %s: %w", srcDir, err)
	}
	if err := t.run(maxFilesOpened); err != nil {
		return err
	}
	return addPartners(srcDir, outDir, t.lookups)
}
