package transformnext

import (
	"fmt"
	"strings"
	"sync/atomic"
)

type source struct {
	prefix       string
	key          string
	sep          rune
	hasHeader    bool
	isCumulative bool
	counter      atomic.Uint32
}

func (s *source) keyFor(id string) []byte {
	if !s.isCumulative {
		return fmt.Appendf([]byte{}, "%s::%s", id, s.key)
	}
	c := s.counter.Add(1)
	return fmt.Appendf([]byte{}, "%s::%s::%d", id, s.key, c)
}

func (s *source) keyPrefixFor(id string) []byte {
	if !s.isCumulative {
		return s.keyFor(id)
	}
	return fmt.Appendf([]byte{}, "%s::%s", id, s.key)
}

func newSource(prefix string, sep rune, hasHeader, isCumulative bool) *source {
	key := strings.ToLower(strings.TrimPrefix(prefix, "Lucro ")[0:3])
	return &source{prefix: prefix, key: key, sep: sep, hasHeader: hasHeader, isCumulative: isCumulative}
}
