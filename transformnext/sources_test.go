package transformnext

import (
	"strings"
	"testing"
)

func TestSourceKey(t *testing.T) {
	srcs := sources()
	for _, exp := range []string{
		"42::cna",
		"42::emp",
		"42::imu::1",
		"42::arb::1",
		"42::pre::1",
		"42::rea::1",
		"42::mot",
		"42::mun",
		"42::nat",
		"42::pai",
		"42::qua",
		"42::sim",
		"42::soc::1",
		"42::tab",
	} {
		key := strings.TrimSuffix(strings.TrimPrefix(exp, "42::"), "::1")
		src, ok := srcs[key]
		if !ok {
			t.Fatalf("expected source %s in %v, got nil", key, srcs)
		}
		t.Run(src.prefix, func(t *testing.T) {
			got := src.keyFor("42")
			if string(got) != exp {
				t.Errorf("expected key for %s to be %s, got %s", src.prefix, exp, string(got))
			}
		})
	}
}
