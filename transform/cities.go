package transform

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// NationalTreasureFileName is the name of the CSV containing city names and
// their codes.
const NationalTreasureFileName = "tabmun.csv"

func citiesLookup(dir string) (lookup, error) {
	var f io.ReadCloser
	var err error
	var pth string
	// backward compatibility: until May/2025 the file was named TABMUN.CSV
	for _, n := range []string{NationalTreasureFileName, strings.ToUpper(NationalTreasureFileName)} {
		pth := filepath.Join(dir, n)
		f, err = os.Open(pth)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("error opening %s: %w", pth, err)
		}
		break
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.Comma = ';'
	l := make(map[int]string)
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading %s: %w", pth, err)
		}
		code, err := strconv.Atoi(row[0])
		if err != nil {
			return nil, fmt.Errorf("error converting %s to int: %w", row[4], err)
		}
		l[code] = row[4]
	}
	return l, nil
}
