package transformnext

import (
	"fmt"
)

type database interface {
	PreLoad() error
	CreateCompanies([][]string) error
	PostLoad() error
	CreateExtraIndexes([]string) error
	MetaSave(string, string) error
}

func Transform(db database) error {
	fmt.Println("TODO")
	return nil
}
