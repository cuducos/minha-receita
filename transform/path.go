package transform

import (
	"os"
	"path/filepath"
	"strings"
)

// PathsForSource lists files for a given `sourceType` in a directory `dir`.
func PathsForSource(t sourceType, dir string) ([]string, error) {
	r, err := os.ReadDir(dir)
	if err != nil {
		return []string{}, err
	}
	var ls []string
	for _, f := range r {
		if f.IsDir() {
			continue
		}
		if strings.Contains(strings.ToLower(f.Name()), strings.ToLower(string(t))) {
			ls = append(ls, filepath.Join(dir, f.Name()))
		}
	}
	return ls, nil
}
