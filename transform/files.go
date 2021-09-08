package transform

import (
	"archive/zip"
	"strings"
)

func isFileOf(a *dataset, f string) bool {
	k := string(a.kind)
	for _, part := range strings.Split(f, ".") {
		if part == k {
			return true
		}
	}

	return false
}

func filesFor(a *dataset, ls []string) []string {
	var r []string
	for _, f := range ls {
		if isFileOf(a, f) {
			r = append(r, f)
		}
	}
	return r
}

func (a *dataset) unzip(e chan<- error, l chan<- []string, s string) {
	z, err := zip.OpenReader(s)
	if err != nil {
		e <- err
		return
	}
	defer z.Close()

	for _, f := range z.File {
		if err := a.lineProducer(l, f); err != nil {
			e <- err
			return
		}
	}

	e <- nil
}
