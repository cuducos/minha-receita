package transform

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/cuducos/go-cnpj"
)

const numOfShards = 256

// shards the base of a CNPJ (first 8 digits) to a number between 0 (included)
// and 256 (not included). It takes the first two digits of the hex digest of
// the base CNPJ MD5 hash and converts them to an integer number.
func shard(n string) (int, error) {
	c := md5.Sum([]byte(cnpj.Base(n)))
	h := hex.EncodeToString(c[:])
	d := h[:2]
	i, err := strconv.ParseInt(d, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing the shard from hex %s for %s: %w", d, n, err)
	}
	return int(i), nil
}
