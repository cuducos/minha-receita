package transform

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

// NationalTreasureFileName is the name of the CSV containing city names and
// their codes.
const NationalTreasureFileName = "TABMUN.CSV"

func citiesLookup(dir string) (lookup, error) {
	pth := filepath.Join(dir, NationalTreasureFileName)
	f, err := os.Open(pth)
	if err != nil {
		return nil, fmt.Errorf("error opening %s: %w", pth, err)
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
