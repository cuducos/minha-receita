package adapter

import "strings"

func isFileOf(a *Adapter, f string) bool {
	k := string(a.kind)
	for _, part := range strings.Split(f, ".") {
		if part == k {
			return true
		}
	}

	return false
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
