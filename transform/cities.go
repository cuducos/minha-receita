package transform

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const nationalTreasureFileName = "tabmun.csv"

// NationalTreasureFilePath deals with backward compatibility: until May/2025
// the file was named TABMUN.CSV. It returns the right file path testing if the
// file exist in lower case or in upper case.
func NationalTreasureFile(dir string) (string, *os.File, error) {
	for _, n := range []string{nationalTreasureFileName, strings.ToUpper(nationalTreasureFileName)} {
		pth := filepath.Join(dir, n)
		f, err := os.Open(pth)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", nil, fmt.Errorf("error opening %s: %w", pth, err)
		}
		return pth, f, nil
	}
	return "", nil, fmt.Errorf("could not find national treasure file in %s", dir)
}

func citiesLookup(dir string) (lookup, error) {
	pth, f, err := NationalTreasureFile(dir)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			slog.Warn("could not close", "path", pth, "error", err)
		}
	}()
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
	// manually add Boa Esperança do Norte (MT): created in 2025 but still absent in tabmun.csv
	l[1182] = "5101837"
	return l, nil
}
