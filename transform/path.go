package transform

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cuducos/go-cnpj"
)

var InvalidCNPJError = errors.New("Invalid CNPJ")
var InvalidPathError = errors.New("Invalid path for a CNPJ JSON file")

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

func PathForCNPJ(c string) (string, error) {
	if !cnpj.IsValid(c) {
		return "", fmt.Errorf("Error finding file path for %s: %w", c, InvalidCNPJError)
	}

	c = cnpj.Unmask(c)
	return c[:8] + string(os.PathSeparator) + c[8:] + ".json", nil
}

func CNPJForPath(p string) (string, error) {
	c := strings.Split(p, string(os.PathSeparator))
	if len(c) < 2 {
		return "", fmt.Errorf("Error finding the CNPJ for %s: %w", p, InvalidPathError)
	}

	r := c[len(c)-2] + strings.TrimSuffix(c[len(c)-1], filepath.Ext(p))
	if !cnpj.IsValid(r) {
		return "", fmt.Errorf("Error finding the CNPJ for %s: %w", p, InvalidPathError)
	}

	return r, nil
}
