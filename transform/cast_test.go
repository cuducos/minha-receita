package transform

import (
	"encoding/json"
	"testing"
	"time"
)

func TestToInt(t *testing.T) {
	t.Run("successful *int casting", func(t *testing.T) {
		n := 42
		tc := []struct {
			value    string
			expected *int
		}{
			{"42", &n},
			{"", nil},
		}
		for _, c := range tc {
			got, err := toInt(c.value)
			if err != nil {
				t.Errorf("expected no errors when converting %s to *int, got %s", c.value, err)
			}
			if c.expected != nil {
				if *got != *c.expected {
					t.Errorf("got %d, expected %d", *got, *c.expected)
				}
			} else {
				if got != c.expected {
					t.Errorf("got %d, expected nil", *got)
				}
			}
		}
	})

	t.Run("unsuccessful *int casting", func(t *testing.T) {
		tc := []string{"4.2", "foobar"}
		for _, v := range tc {
			_, err := toInt(v)
			if err == nil {
				t.Errorf("expected a error when converting %s to *int, got ni", v)
			}
		}
	})
}

func TestToFloat(t *testing.T) {
	t.Run("successful *float32 casting", func(t *testing.T) {
		n1 := float32(42)
		n2 := float32(0.42)
		tc := []struct {
			value    string
			expected *float32
		}{
			{"42", &n1},
			{"0.42", &n2},
			{"", nil},
		}
		for _, c := range tc {
			got, err := toFloat(c.value)
			if err != nil {
				t.Errorf("expected no errors when converting %s to *float32, got %s", c.value, err)
			}
			if c.expected != nil {
				if *got != *c.expected {
					t.Errorf("got %f, expected %f", *got, *c.expected)
				}
			} else {
				if got != c.expected {
					t.Errorf("got %f, expected nil", *got)
				}
			}
		}
	})

	t.Run("unsuccessful *float32 casting", func(t *testing.T) {
		_, err := toFloat("foobar")
		if err == nil {
			t.Errorf("expected a error when converting foobar to *float32, got nil")
		}
	})
}

func TestToBool(t *testing.T) {
	expectedTrue := true
	expectedFalse := false
	tc := []struct {
		value    string
		expected *bool
	}{
		{"S", &expectedTrue},
		{"s", &expectedTrue},
		{"N", &expectedFalse},
		{"n", &expectedFalse},
		{"", nil},
		{" ", nil},
		{"42", nil},
	}
	for _, c := range tc {
		got := toBool(c.value)
		if got == nil && c.expected != nil {
			t.Errorf("expected %s to be nil, got %t", c.value, *got)
		}
		if got != nil && *got != *c.expected {
			t.Errorf("expected %s to be %t, got %t", c.value, *c.expected, *got)
		}
	}
}

func TestToDate(t *testing.T) {
	t.Run("successful date casting", func(t *testing.T) {
		v := "19940717"
		d, err := time.Parse(dateInputFormat, v)
		if err != nil {
			t.Errorf("could not create a date for the test")
		}
		expected := date(d)

		tc := []struct {
			value    string
			expected *date
		}{
			{v, &expected},
			{"", nil},
			{"00000000", nil},
		}
		for _, c := range tc {
			got, err := toDate(c.value)
			if err != nil {
				t.Errorf("expected no errors when converting %s to date, got %s", c.value, err)
			}
			if c.expected != nil {
				if *got != *c.expected {
					t.Errorf("got %q, expected %q", time.Time(*got), time.Time(*c.expected))
				}
			} else {
				if got != c.expected {
					t.Errorf("got %q, expected nil", time.Time(*got))
				}
			}
		}
	})

	t.Run("unsuccessful date casting", func(t *testing.T) {
		got, err := toDate("foobar")
		if err == nil {
			t.Errorf("expected a error when converting foobar to date, got nil")
		}
		if got != nil {
			t.Errorf("expected nil, got %s", time.Time(*got))
		}
		if err == nil {
			t.Error("expected an error, got nil")
		}
	})
}

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
