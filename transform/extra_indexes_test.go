package transform

import "testing"

func TestIndexValidator_validate(t *testing.T) {
	v := NewIndexValidator()
	err := v.Validate("qsa.nome_socio")
	if err != nil {
		t.Errorf("expected no error for valid index, got %s", err)
	}
	err = v.Validate("index1")
	if err == nil {
		t.Errorf("expected error for index1 index, got nil")
	}
}
