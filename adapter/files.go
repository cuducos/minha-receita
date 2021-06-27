package adapter

import "strings"

func isFileOf(a *Adapter, f string) bool {
	s := strings.Split(f, ".")
	p := len(s) - 2
	if p < 0 {
		return false
	}

	return s[p] == string(a.kind)
}

func filesFor(a *Adapter, ls []string) []string {
	var r []string
	for _, f := range ls {
		if isFileOf(a, f) {
			r = append(r, f)
		}
	}
	return r
}
