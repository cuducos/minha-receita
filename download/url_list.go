package download

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

func simpleURLList(fs []file) string {
	urls := make([]string, 0, len(fs))
	for _, f := range fs {
		urls = append(urls, f.url)
	}
	sort.Strings(urls)
	return strings.Join(urls, "\n")
}

func tsvURLList(fs []file) (string, error) {
	var ls []string
	buf := bytes.NewBufferString("TsvHttpData-1.0\n")
	for _, f := range fs {
		ls = append(ls, fmt.Sprintf("%s\t%d", f.url, f.size))
	}
	sort.Strings(ls)
	buf.WriteString(strings.Join(ls, "\n"))
	return buf.String(), nil
}

func listURLs(fs []file, tsv bool) error {
	var err error
	var s string
	if tsv {
		s, err = tsvURLList(fs)
		if err != nil {
			return fmt.Errorf("error creating url list: %w", err)
		}
	} else {
		s = simpleURLList(fs)
	}
	fmt.Print(s)
	return nil
}
