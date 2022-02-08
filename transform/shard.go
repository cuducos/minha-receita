package transform

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
)

const numOfShards = 64

// shards the base of a CNPJ (first 8 digits) to a number between 0 (included)
// and 16 (not included). It takes the first two digits of the hex digest of
// the base CNPJ MD5 hash. It multiplies the first of these digits by 1, 2, 3
// or 4 depending on the second of thsse digits (1 if the second digit is below
// 4 and so on).
func shard(n string) (int, error) {
	c := md5.Sum([]byte(n[0:8]))
	h := hex.EncodeToString(c[:])
	h1 := h[:1]
	h2 := h[1:2]
	d1, err := strconv.ParseInt(h1, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing the shard from hex %s for %s: %w", h1, n, err)
	}
	d2, err := strconv.ParseInt(h2, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing the shard from hex %s for %s: %w", h2, n, err)
	}
	var m int
	if d2 < 4 {
		m = 1
	} else if d2 < 8 {
		m = 2
	} else if m < 12 {
		m = 3
	} else {
		m = 4
	}
	return int(d1) * m, nil
}
