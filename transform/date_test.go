package transform

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDate(t *testing.T) {
	t.Run("successful unmarshal and marshal", func(t *testing.T) {
		var d date
		err := json.Unmarshal([]byte(`"1967-06-30"`), &d)
		if err != nil {
			t.Errorf("expected no error on date Unmarshal, got %s", err)
		}

		got := time.Time(d)
		if got.Year() != 1967 {
			t.Errorf("expected year to be 1967, got %d", got.Year())
		}
		if got.Month() != 6 {
			t.Errorf("expected year to be 6, got %d", got.Month())
		}
		if got.Day() != 30 {
			t.Errorf("expected year to be 30, got %d", got.Day())
		}

		b, err := d.MarshalJSON()
		s := string(b)
		if err != nil {
			t.Errorf("expected no error on marshal %v, s %s", d, err)
		}
		if s != "\"1967-06-30\"" {
			t.Errorf("expected result to be \"1967-06-30\", s %s", s)
		}
	})
}
