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
	counter      atomic.Uint32
}

func (s *source) keyFor(id string) []byte {
	k := strings.ToLower(strings.TrimPrefix(s.prefix, "Lucro ")[0:3])
	if !s.isCumulative {
		return fmt.Appendf([]byte{}, "%s::%s", id, k)
	}
	c := s.counter.Add(1)
	return fmt.Appendf([]byte{}, "%s::%s::%d", id, k, c)
}

func (s *source) keyPrefixFor(id string) []byte {
	if !s.isCumulative {
		return s.keyFor(id)
	}
	k := strings.ToLower(strings.TrimPrefix(s.prefix, "Lucro ")[0:3])
	return fmt.Appendf([]byte{}, "%s::%s", id, k)
}

func newSource(prefix string, sep rune, hasHeader, isCumulative bool) *source {
	return &source{prefix: prefix, sep: sep, hasHeader: hasHeader, isCumulative: isCumulative}
}
