package adapter

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

const (
	updateSpinner   = 236 * time.Millisecond
	updateFileStats = 5 * time.Second
)

type spinner struct {
	pos    int
	chars  []string
	path   string
	exists bool
	size   uint64
	error  string
}

func (s *spinner) char() string {
	c := s.chars[s.pos]
	s.pos++
	if s.pos >= len(s.chars) {
		s.pos = 0
	}

	return c
}

func (s *spinner) update() {
	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			s.exists = false
		} else {
			s.error = err.Error()
		}
		return
	}
	defer f.Close()

	i, err := f.Stat()
	if err != nil {
		s.error = err.Error()
		return
	}

	s.error = ""
	s.exists = true
	s.size = uint64(i.Size())
}

func (s *spinner) status() string {
	if s.error != "" {
		return fmt.Sprintf("(%s)", s.error)
	}

	if !s.exists {
		return ""
	}

	return fmt.Sprintf("(%s)", humanize.Bytes(s.size))
}

func (s *spinner) read() string {
	var c []string
	for _, p := range []string{s.char(), s.path, s.status()} {
		if p != "" {
			c = append(c, p)
		}
	}

	return strings.Join(c, " ")
}

func newSpinner(p string) spinner {
	s := spinner{path: p, chars: []string{"⠋", "⠙", "⠸", "⠴", "⠦", "⠇"}}
	s.update()
	return s
}

func status(as []*Adapter) {
	var fs []*spinner
	for _, a := range as {
		s := newSpinner(csvPathFor(a))
		fs = append(fs, &s)
	}

	go func() {
		for {
			time.Sleep(updateFileStats)
			for _, s := range fs {
				s.update()
			}
		}
	}()

	for {
		var l []string
		for _, s := range fs {
			l = append(l, s.read())
		}

		fmt.Printf(fmt.Sprintf("\r %s", strings.Join(l, " | ")))
		time.Sleep(updateSpinner)
	}
}
