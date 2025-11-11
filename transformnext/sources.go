package transformnext

import (
	"fmt"
	"strings"
	"sync/atomic"
)

type source struct {
	prefix       string
	sep          rune
	hasHeader    bool
	isCumulative bool

	// for accumulative keys
	counter atomic.Uint32

	// progress tracker
	total atomic.Int64
	done  atomic.Int64
}

func (s *source) keyFor(id string) []byte {
	k := strings.ToLower(strings.TrimPrefix(s.prefix, "Lucro ")[0:3])
	if !s.isCumulative {
		return []byte(fmt.Sprintf("%s::%s", id, k))
	}
	c := s.counter.Add(1)
	return []byte(fmt.Sprintf("%s::%s::%d", id, k, c))
}

func newSource(prefix string, sep rune, hasHeader, isCumulative bool) *source {
	return &source{prefix: prefix, sep: sep, hasHeader: hasHeader, isCumulative: isCumulative}
}
