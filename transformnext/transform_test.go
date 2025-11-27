package transformnext

import (
	"testing"
)

func TestTransform(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Test with empty directory - should fail gracefully
	tmpDir := "/nonexistent/directory"
	err := Transform(tmpDir)
	if err == nil {
		t.Error("expected Transform to fail with nonexistent directory")
	}
}

func TestSources(t *testing.T) {
	srcs := sources()
	expectedSources := []string{
		"Cnaes", "Empresas", "Imunes e Isentas", "Lucro Arbitrado",
		"Lucro Presumido", "Lucro Real", "Motivos", "Municipios",
		"Naturezas", "Paises", "Qualificacoes", "Simples", "Socios", "tabmun",
	}

	if len(srcs) != len(expectedSources) {
		t.Errorf("expected %d sources, got %d", len(expectedSources), len(srcs))
	}

	for i, src := range srcs {
		if src.prefix != expectedSources[i] {
			t.Errorf("expected source %s at index %d, got %s", expectedSources[i], i, src.prefix)
		}
	}
}

func TestNewSource(t *testing.T) {
	src := newSource("Test", ',', true, false)

	if src.prefix != "Test" {
		t.Errorf("expected prefix Test, got %s", src.prefix)
	}

	if src.sep != ',' {
		t.Errorf("expected separator ',', got %c", src.sep)
	}

	if !src.hasHeader {
		t.Error("expected hasHeader to be true")
	}

	if src.isCumulative {
		t.Error("expected isCumulative to be false")
	}
}

func TestSourceKeyFor(t *testing.T) {
	tests := []struct {
		name       string
		prefix     string
		id         string
		cumulative bool
		expected   string
	}{
		{
			name:       "non-cumulative source",
			prefix:     "Cnaes",
			id:         "12345",
			cumulative: false,
			expected:   "12345::cna",
		},
		{
			name:       "cumulative source",
			prefix:     "Socios",
			id:         "12345",
			cumulative: true,
			expected:   "12345::soc::1",
		},
		{
			name:       "lucro source",
			prefix:     "Lucro Presumido",
			id:         "12345",
			cumulative: false,
			expected:   "12345::pre",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := newSource(tt.prefix, ';', false, tt.cumulative)
			key := src.keyFor(tt.id)

			if string(key) != tt.expected {
				t.Errorf("expected key %s, got %s", tt.expected, string(key))
			}
		})
	}
}
