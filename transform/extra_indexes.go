package transform

import "fmt"

type IndexValidator struct {
	indexes map[string]struct{}
}

func (v *IndexValidator) Validate(i string) error {
	_, ok := v.indexes[i]
	if !ok {
		return fmt.Errorf("invalid index %s", i)
	}
	return nil
}

func NewIndexValidator() IndexValidator {
	m := make(map[string]struct{})
	for _, i := range CompanyJSONFields() {
		m[i] = struct{}{}
	}
	return IndexValidator{m}
}
