package transformnext

import (
	"context"
	"testing"
)

func TestLoadCSVs(t *testing.T) {
	srcs := sources()

	for _, tc := range []struct {
		key string
		exp []string
	}{ // expected value is the first column of each row
		{"cna", []string{"6204000", "6201501", "6202300", "6203100", "6209100", "6311900"}},
		{"emp", []string{"33683111", "19131243"}},
		{"imu", []string{"2023"}},
		{"arb", []string{"2023"}},
		{"pre", []string{"2018"}},
		{"rea", []string{"2023"}},
		{"mot", []string{"00", "01"}},
		{"mun", []string{"9701"}},
		{"nat", []string{"2011"}},
		{"pai", []string{"105"}},
		{"qua", []string{"05", "10", "16"}},
		{"sim", []string{"33683111"}},
		{"soc", []string{"33683111", "33683111", "33683111", "33683111", "33683111", "33683111", "19131243"}},
		{"tab", []string{"9701"}},
	} {
		src := srcs[tc.key]
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
			for _, id := range tc.exp {
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
