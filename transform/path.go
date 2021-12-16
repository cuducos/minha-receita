package transform

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cuducos/go-cnpj"
)

// ErrInvalidCNPJ is raised when a string value is not a valid CNPJ number.
var ErrInvalidCNPJ = errors.New("invalid CNPJ")

// ErrInvalidPath is raised when a path does not correspond to the path of a
// valid CNPJ.
var ErrInvalidPath = errors.New("invalid path for a CNPJ JSON file")

// PathsForSource lists files for a given `sourceType` in a directory `dir`.
func PathsForSource(t sourceType, dir string) ([]string, error) {
	var ls []string

	r, err := os.ReadDir(dir)
	if err != nil {
		return ls, err
	}

	s := strings.ToLower(string(t) + ".zip")
	for _, f := range r {
		if f.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToLower(f.Name()), s) {
			ls = append(ls, filepath.Join(dir, f.Name()))
		}
	}
	return ls, nil
}

// PathForCNPJ creates the file path for a JSON file related to a CNPJ.
func PathForCNPJ(c string) (string, error) {
	if !cnpj.IsValid(c) {
		return "", fmt.Errorf("Error finding file path for %s: %w", c, ErrInvalidCNPJ)
	}

	c = cnpj.Unmask(c)
	return c[:8] + string(os.PathSeparator) + c[8:] + ".json", nil
}

// CNPJForPath creates a CNPJ from a path of a JSON file related to a CNPJ.
func CNPJForPath(p string) (string, error) {
	c := strings.Split(p, string(os.PathSeparator))
	if len(c) < 2 {
		return "", fmt.Errorf("Error finding the CNPJ for %s: %w", p, ErrInvalidPath)
	}

	r := c[len(c)-2] + strings.TrimSuffix(c[len(c)-1], filepath.Ext(p))
	if !cnpj.IsValid(r) {
		return "", fmt.Errorf("Error finding the CNPJ for %s: %w", p, ErrInvalidPath)
	}

	return r, nil
}
