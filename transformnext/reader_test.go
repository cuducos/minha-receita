package transformnext

import (
	"context"
	"testing"

	"github.com/cuducos/minha-receita/testutils"
)

func TestReadCSVs(t *testing.T) {
	for _, tc := range []struct {
		name     string
		prefix   string
		sep      rune
		expected []string // first column of each row
	}{
		{"csv", "tabmun", ';', []string{"9701"}},
		{"zip", "Empresas", ';', []string{"33683111", "19131243"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			ch := make(chan []string)
			ok := make(chan struct{})
			var rows [][]string
			go func() {
				defer close(ok)
				for row := range ch {
					rows = append(rows, row)
				}
			}()
			err := readCSVs(ctx, "../testdata", tc.prefix, tc.sep, false, ch)
			if err != nil {
				t.Errorf("expected no error reading csvs, got %s", err)
			}
			close(ch)
			<-ok
			if len(rows) != len(tc.expected) {
				t.Errorf("expected %d rows, got %d", len(tc.expected), len(rows))
			}
			var got []string
			for _, r := range rows {
				got = append(got, r[0])
			}
			testutils.AssertArraysHaveSameItems(t, got, tc.expected)
		})
	}
}
