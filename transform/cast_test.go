package transform

import (
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

func TestToDate(t *testing.T) {
	t.Run("successful date casting", func(t *testing.T) {
		v := "19940717"
		d, err := time.Parse(dateInputFormat, v)
		if err != nil {
			t.Errorf("could not create a date for the test")
		}
		expected := date{&d}

		tc := []struct {
			value    string
			expected date
		}{
			{v, expected},
			{"", date{nil}},
		}
		for _, c := range tc {
			got, err := toDate(c.value)
			if err != nil {
				t.Errorf("expected no errors when converting %s to date, got %s", c.value, err)
			}
			if c.expected.time != nil {
				if *got.time != *c.expected.time {
					t.Errorf("got %q, expected %q", *got.time, *c.expected.time)
				}
			} else {
				if got != c.expected {
					t.Errorf("got %q, expected nil", *got.time)
				}
			}
		}
	})

	t.Run("unsuccessful date casting", func(t *testing.T) {
		_, err := toDate("foobar")
		if err == nil {
			t.Errorf("expected a error when converting foobar to date, got nil")
		}
	})
}
