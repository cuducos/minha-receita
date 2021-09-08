package transform

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gosuri/uilive"
)

const (
	updateSpinner   = 236 * time.Millisecond
	updateFileStats = 5 * time.Second
)

type spinner struct {
	pos         int
	chars       []string
	adapter     *dataset
	exists      bool
	error       string
	currentSize uint64
}

func (s *spinner) char() string {
	if s.adapter.done {
		return "✓"
	}

	c := s.chars[s.pos]
	s.pos++
	if s.pos >= len(s.chars) {
		s.pos = 0
	}

	return c
}

func (s *spinner) update() {
	f, err := os.Open(s.adapter.csvPath())
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
	s.currentSize = uint64(i.Size())
}

func (s *spinner) status() string {
	if s.error != "" {
		return fmt.Sprintf("(%s)", s.error)
	}

	if !s.exists {
		return ""
	}

	return fmt.Sprintf("(%s)", humanize.Bytes(s.currentSize))
}

func (s *spinner) read() string {
	var c []string
	for _, p := range []string{s.char(), s.adapter.csvPath(), s.status()} {
		if p != "" {
			c = append(c, p)
		}
	}

	return strings.Join(c, " ")
}

func newSpinner(a *dataset) spinner {
	s := spinner{adapter: a, chars: []string{"⠋", "⠙", "⠸", "⠴", "⠦", "⠇"}}
	s.update()
	return s
}

func output(fs []*spinner, ui *uilive.Writer) {
	var l []string
	for _, s := range fs {
		l = append(l, s.read())
	}

	fmt.Fprint(ui, "\n"+strings.Join(l, "\n"))
	time.Sleep(updateSpinner)
}

func startSpinners(c chan struct{}, as []*dataset) {
	var fs []*spinner
	for _, a := range as {
		s := newSpinner(a)
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

	ui := uilive.New()
	ui.Start()

	for {
		select {
		case <-c:
			for _, s := range fs {
				s.update()
			}
			output(fs, ui)
			c <- struct{}{}
			return
		default:
			output(fs, ui)
		}
	}
}
