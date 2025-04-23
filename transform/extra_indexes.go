package transform

import (
	"fmt"
	"strings"
)

func ValidateIndexes(idxs []string) error {
	m := make(map[string]struct{})
	var errs []string
	for _, i := range CompanyJSONFields() {
		m[i] = struct{}{}
	}
	for _, i := range idxs {
		_, ok := m[i]
		if !ok {
			errs = append(errs, i)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("invalid index(es): %s", strings.Join(errs, ", "))
	}
	return nil
}
