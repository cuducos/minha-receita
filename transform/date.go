package transform

import (
	"strings"
	"time"
)

const dateInputFormat = "20060102"
const dateOutputFormat = "2006-01-02"

type date struct {
	time *time.Time
}

func (d *date) UnmarshalJSON(b []byte) error {
	if string(b) == "" {
		return nil
	}

	t, err := time.Parse(dateOutputFormat, strings.Trim(string(b), `"`))
	if err != nil {
		return err
	}

	d.time = &t
	return nil
}

func (d *date) MarshalJSON() ([]byte, error) {
	if d.time == nil {
		return []byte("null"), nil
	}

	return []byte(`"` + d.time.Format(dateOutputFormat) + `"`), nil
}
