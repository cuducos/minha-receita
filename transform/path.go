package transform

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cuducos/go-cnpj"
)

var nonDigits = regexp.MustCompile(`\D`)

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
		return "", fmt.Errorf("error finding file path for %s: invalid cnpj", c)
	}

	c = cnpj.Mask(c)
	p := nonDigits.Split(cnpj.Mask(c), 5)
	n := p[3] + p[4] + ".json"
	return strings.Join(append(p[:3], n), string(os.PathSeparator)), nil
}

func pathForBaseCNPJ(s string) (string, error) {
	if len(s) != 8 {
		return "", fmt.Errorf("invalid base cnpj: %s", s)
	}
	return filepath.Join(s[:2], s[2:5], s[5:]), nil
}

// CNPJForPath creates a CNPJ from a path of a JSON file related to a CNPJ.
func CNPJForPath(f string) (string, error) {
	p := strings.Split(f, string(os.PathSeparator))
	if len(p) < 4 {
		return "", fmt.Errorf("error finding the cnpj for %s: invalid path", f)
	}
	c := strings.TrimSuffix(strings.Join(p[len(p)-4:], ""), filepath.Ext(f))
	if !cnpj.IsValid(c) {
		return "", fmt.Errorf("error finding the cnpj for %s: invalid resulting cnpj", f)
	}
	return c, nil
}
