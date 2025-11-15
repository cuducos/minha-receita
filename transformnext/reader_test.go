package transformnext

import (
	"context"
	"testing"
)

func TestLoadCSVs(t *testing.T) {
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
			kv, err := newBadger(t.TempDir(), false)
			defer func() {
				if err := kv.db.Close(); err != nil {
					t.Errorf("expected no error closing badger, got %s", err)
				}
			}()
			if err != nil {
				t.Errorf("expected no error creating badger, got %s", err)
			}
			if err := loadCSVs(ctx, "../testdata", src, nil, kv); err != nil {
				t.Errorf("expected no error loading csvs, got %s", err)
			}
			for _, id := range exp {
				key := src.keyPrefixFor(id)
				got, err := kv.getPrefix(key)
				if err != nil {
					t.Errorf("expect no error getting %s, got %s", string(key), err)
				}
				if got == nil {
					t.Errorf("expected to find key %s, got nil", string(key))
				}
			}
		})
	}
}
