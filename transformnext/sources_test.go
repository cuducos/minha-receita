package transformnext

import "testing"

func TestSourceKey(t *testing.T) {
	srcs := sources()
	for idx, exp := range []string{
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
		src := srcs[idx]
		t.Run(src.prefix, func(t *testing.T) {
			got := src.keyFor("42")
			if string(got) != exp {
				t.Errorf("expected key for %s to be %s, got %s", src.prefix, exp, string(got))
			}
		})
	}
}
