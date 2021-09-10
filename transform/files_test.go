package transform

import "testing"

func TestIsFileOf(t *testing.T) {
	types := []kind{company, partner, facility}
	testCases := []struct {
		name     string
		expected kind
	}{
		{"F.K03200$W.SIMPLES.CSV.D10612.zip", simple},
		{"F.K03200$Z.D10612.CNAECSV.zip", cnae},
		{"F.K03200$Z.D10612.MOTICSV.zip", motive},
		{"F.K03200$Z.D10612.MUNICCSV.zip", city},
		{"F.K03200$Z.D10612.NATJUCSV.zip", nature},
		{"F.K03200$Z.D10612.PAISCSV.zip", country},
		{"F.K03200$Z.D10612.QUALSCSV.zip", qualification},
		{"K3241.K03200Y0.D10612.EMPRECSV.zip", company},
		{"K3241.K03200Y0.D10612.ESTABELE.zip", facility},
		{"K3241.K03200Y0.D10612.SOCIOCSV.zip", partner},
		{"K3241.K03200Y1.D10612.SOCIOCSV.zip", partner},
		{"K3241.K03200Y2.D10612.ESTABELE.zip", facility},
		{"K3241.K03200Y2.D10612.SOCIOCSV.zip", partner},
		{"K3241.K03200Y5.D10612.EMPRECSV.zip", company},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			for _, k := range types {
				e := c.expected == k
				if r := isFileOf(newDataset(k, "testdata", ""), c.name); r != e {
					t.Errorf("Expected %s to be of kind %s", c.name, c.expected)
				}
			}
		})
	}
}

func TestArchivesFor(t *testing.T) {
	ls := []string{
		".gitkeep",
		"F.K03200$W.SIMPLES.CSV.D10612.zip",
		"F.K03200$Z.D10612.CNAECSV.zip",
		"F.K03200$Z.D10612.MOTICSV.zip",
		"F.K03200$Z.D10612.MUNICCSV.zip",
		"F.K03200$Z.D10612.NATJUCSV.zip",
		"F.K03200$Z.D10612.PAISCSV.zip",
		"F.K03200$Z.D10612.QUALSCSV.zip",
		"K3241.K03200Y0.D10612.EMPRECSV.zip",
		"K3241.K03200Y0.D10612.ESTABELE.zip",
		"K3241.K03200Y0.D10612.SOCIOCSV.zip",
		"K3241.K03200Y1.D10612.SOCIOCSV.zip",
		"K3241.K03200Y2.D10612.ESTABELE.zip",
		"K3241.K03200Y2.D10612.SOCIOCSV.zip",
		"K3241.K03200Y5.D10612.EMPRECSV.zip",
	}
	testCases := []struct {
		kind     kind
		expected []string
	}{

		{city, []string{"F.K03200$Z.D10612.MUNICCSV.zip"}},
		{cnae, []string{"F.K03200$Z.D10612.CNAECSV.zip"}},
		{company, []string{
			"K3241.K03200Y0.D10612.EMPRECSV.zip",
			"K3241.K03200Y1.D10612.EMPRECSV.zip",
		}},
		{country, []string{"F.K03200$Z.D10612.PAISCSV.zip"}},
		{facility, []string{
			"K3241.K03200Y1.D10612.ESTABELE.zip",
			"K3241.K03200Y0.D10612.ESTABELE.zip",
		}},
		{motive, []string{"F.K03200$Z.D10612.MOTICSV.zip"}},
		{nature, []string{"F.K03200$Z.D10612.NATJUCSV.zip"}},
		{partner, []string{
			"K3241.K03200Y0.D10612.SOCIOCSV.zip",
			"K3241.K03200Y1.D10612.SOCIOCSV.zip",
		}},
		{qualification, []string{"F.K03200$Z.D10612.QUALSCSV.zip"}},
		{simple, []string{"F.K03200$W.SIMPLES.CSV.D10612.zip"}},
	}

	for _, c := range testCases {
		r := filesFor(newDataset(c.kind, "testdata", ""), ls)
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
