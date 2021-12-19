package transform

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
)

const numOfShards = 256

// shards the base of a CNPJ (first 8 digits) to a number between 0 (included)
// and 256 (not included).
func shard(c string) (int, error) {
	h := md5.Sum([]byte(c[0:8]))
	s := hex.EncodeToString(h[:])[0:2]
	i, err := strconv.ParseInt(s, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing the shard from hex %s for %s: %w", s, c, err)
	}
	return int(i), nil
}
