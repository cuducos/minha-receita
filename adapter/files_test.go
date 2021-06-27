package adapter

import "testing"

func TestIsFileOf(t *testing.T) {
	types := []kind{company, partner, facility}
	cases := []struct {
		name     string
		expected kind
	}{
		{"K3241.K03200Y0.D10612.EMPRECSV.zip", company},
		{"K3241.K03200Y0.D10612.ESTABELE.zip", facility},
		{"K3241.K03200Y0.D10612.SOCIOCSV.zip", partner},
	}

	for _, c := range cases {
		for _, k := range types {
			a := Adapter{k, "testdata"}
			e := c.expected == k
			if r := isFileOf(&a, c.name); r != e {
				t.Errorf("Expected %s to be of kind %s", c.name, c.expected)
			}
		}
	}
}

func TestArchivesFor(t *testing.T) {
	ls := []string{
		".gitkeep",
		"F.K03200$W.SIMPLES.CSV.D10612.zip",
		"K3241.K03200Y0.D10612.EMPRECSV.zip",
		"K3241.K03200Y0.D10612.ESTABELE.zip",
		"K3241.K03200Y0.D10612.SOCIOCSV.zip",
		"K3241.K03200Y1.D10612.EMPRECSV.zip",
		"K3241.K03200Y1.D10612.ESTABELE.zip",
		"K3241.K03200Y1.D10612.SOCIOCSV.zip",
	}
	cases := []struct {
		kind     kind
		expected []string
	}{
		{company, []string{
			"K3241.K03200Y0.D10612.EMPRECSV.zip",
			"K3241.K03200Y1.D10612.EMPRECSV.zip",
		},
		},
		{facility, []string{
			"K3241.K03200Y0.D10612.ESTABELE.zip",
			"K3241.K03200Y1.D10612.ESTABELE.zip",
		}},
		{partner, []string{
			"K3241.K03200Y0.D10612.SOCIOCSV.zip",
			"K3241.K03200Y1.D10612.SOCIOCSV.zip",
		}},
	}

	for _, c := range cases {
		a := Adapter{c.kind, "testdata"}
		r := filesFor(&a, ls)

		if len(r) != len(c.expected) {
			t.Errorf("Expected %d files for %s, got %d", len(c.expected), c.kind, len(r))
		}

		for i, _ := range r {
			if r[i] != c.expected[i] {
				t.Errorf("Expected item %d of %s to be %s, got %s", i, c.kind, c.expected[i], r[i])
			}
		}
	}

}
