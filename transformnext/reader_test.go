package transformnext

import (
	"context"
	"testing"

	"github.com/cuducos/minha-receita/testutils"
)

func TestReadCSVs(t *testing.T) {
	srcs := sources()
	for idx, exp := range [][]string{ // expected value is the first column of each row
		{"6204000", "6201501", "6202300", "6203100", "6209100", "6311900"},
		{"33683111", "19131243"},
		{"2023"},
		{"2023"},
		{"2018"},
		{"2023"},
		{"00", "01"},
		{"9701"},
		{"2011"},
		{"105"},
		{"05", "10", "16"},
		{"33683111"},
		{"33683111", "33683111", "33683111", "33683111", "33683111", "33683111", "19131243"},
		{"9701"},
	} {
		src := srcs[idx]
		t.Run(src.prefix, func(t *testing.T) {
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
			err := readCSVs(ctx, "../testdata", src, ch)
			if err != nil {
				t.Errorf("expected no error reading csvs, got %s", err)
			}
			close(ch)
			<-ok
			if len(rows) != len(exp) {
				t.Errorf("expected %d rows, got %d", len(exp), len(rows))
			}
			var got []string
			for _, r := range rows {
				got = append(got, r[0])
			}
			testutils.AssertArraysHaveSameItems(t, got, exp)
		})
	}
}
