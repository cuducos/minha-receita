package transform

import (
	"fmt"
)

// MaxFilesOpened is the maximum number of files opened at the same time.
const MaxFilesOpened = 512 // TODO how to optimize this number?

// Transform the downloaded files for company venues creating a JSON file per CNPJ
func Transform(srcDir, outDir string) error {
	t, err := newTask(srcDir, outDir)
	if err != nil {
		return fmt.Errorf("error creating new task for venues in %s: %w", srcDir, err)
	}
	if err := t.run(MaxFilesOpened); err != nil {
		return err
	}
	if err := addBases(srcDir, outDir, t.lookups); err != nil {
		return err
	}
	if err := addBases(srcDir, outDir, t.lookups); err != nil {
		return err
	}
	if err := addPartners(srcDir, outDir, t.lookups); err != nil {
		return err
	}
	return addSimplesToCompanies(outDir)
}
