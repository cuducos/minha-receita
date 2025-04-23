package transform

import "testing"

func TestIndexValidator(t *testing.T) {
	err := ValidateIndexes([]string{"qsa.nome_socio"})
	if err != nil {
		t.Errorf("expected no error for valid index, got %s", err)
	}
	err = ValidateIndexes([]string{"index1"})
	if err == nil {
		t.Errorf("expected error for index1 index, got nil")
	}
}
