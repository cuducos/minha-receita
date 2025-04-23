package testutils

import "testing"

func AssertArraysHaveSameItems(t *testing.T, a1, a2 []string) {
	if len(a1) != len(a2) {
		t.Errorf("got %v", a1) // TODO: remove
		t.Errorf("arrays lengths are different: %d != %d", len(a1), len(a2))
		return
	}

	c1 := make(map[string]int)
	c2 := make(map[string]int)
	for _, v := range a1 {
		c1[v]++
	}
	for _, v := range a2 {
		c2[v]++
	}

	diff := make(map[string]struct{})
	for k := range c1 {
		if c1[k] != c2[k] {
			diff[k] = struct{}{}
		}
	}
	for k := range c2 {
		if c1[k] != c2[k] {
			diff[k] = struct{}{}
		}
	}

	for k := range diff {
		t.Errorf("%q appears %d in the first array, but %d in the second array", k, c1[k], c2[k])
	}
}
