package transform

import (
	"strings"
	"time"
)

const dateInputFormat = "20060102"
const dateOutputFormat = "2006-01-02"

type date time.Time

func (d *date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" {
		return nil
	}
	t, err := time.Parse(dateOutputFormat, s)
	if err != nil {
		return err
	}

	*d = date(t)
	return nil
}

func (d *date) MarshalJSON() ([]byte, error) {
	t := time.Time(*d)
	return []byte(`"` + t.Format(dateOutputFormat) + `"`), nil
}
