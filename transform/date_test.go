package transform

import "testing"

func TestDate(t *testing.T) {
	t.Run("successful date", func(t *testing.T) {
		d := date{}

		d.UnmarshalJSON([]byte("\"1967-06-30\""))
		if d.time.Year() != 1967 {
			t.Errorf("expected year to be 1967, got %d", d.time.Year())
		}
		if d.time.Month() != 6 {
			t.Errorf("expected year to be 6, got %d", d.time.Month())
		}
		if d.time.Day() != 30 {
			t.Errorf("expected year to be 30, got %d", d.time.Day())
		}

		b, err := d.MarshalJSON()
		got := string(b)
		if err != nil {
			t.Errorf("expected no error on marshal %v, got %s", d, err)
		}
		if got != "\"1967-06-30\"" {
			t.Errorf("expected result to be \"1967-06-30\", got %s", got)
		}
	})

	t.Run("missing date", func(t *testing.T) {
		d := date{nil}
		b, err := d.MarshalJSON()
		got := string(b)
		if err != nil {
			t.Errorf("expected no error on marshal %v, got %s", d, err)
		}
		if got != "null" {
			t.Errorf("expected result to be null, got %s", got)
		}

	})
}
