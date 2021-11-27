package transform

import (
	"os"
	"path/filepath"
	"strings"
)

type sourceType string

const (
	venue         sourceType = "ESTABELE"
	motive                   = "MOTICSV"
	main                     = "EMPRECSV"
	city                     = "MUNICCSV"
	cnae                     = "CNAECSV"
	country                  = "PAISCSV"
	nature                   = "NATJUCSV"
	partner                  = "SOCIOCSV"
	qualification            = "QUALSCSV"
	simple                   = "SIMPLES"
)

func filesFor(t sourceType, dir string) ([]string, error) {
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
