// Package transform includes these cast functions to allow the project to
// accomplish two things:
//
//   - convert strings from the CSV files to other formats (e.g. int, float32,
//     time.Time);
//   - differentiate empty values (such as 0 for int) from missing values.
//
// This is achieved using pointers, so we have nil as a marker for missing
// value.
//
// Since our use case involves serving this data in JSON format, this is crucial
// so we can use `null` when there is no value, and "0" when the value of an
// integer, for example, is 0.
package transform

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

const (
	dateInputFormat  = "20060102"
	dateOutputFormat = "2006-01-02"
)

func toInt(v string) (*int, error) {
	if v == "" {
		return nil, nil
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return nil, fmt.Errorf("error converting %s to int: %w", v, err)
	}
	return &i, nil
}

func toFloat(v string) (*float32, error) {
	if v == "" {
		return nil, nil
	}
	f, err := strconv.ParseFloat(strings.ReplaceAll(v, ",", "."), 32)
	if err != nil {
		return nil, fmt.Errorf("error converting %s to float32: %w", v, err)
	}
	f32 := float32(f)
	return &f32, nil
}

func toBool(v string) *bool {
	v = strings.ToUpper(v)
	var b bool
	switch v {
	case "S":
		b = true
	case "N":
		b = false
	default:
		return nil
	}
	return &b
}

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

func (d date) MarshalBSONValue() (bsontype.Type, []byte, error) {
	t := time.Time(d)
	return bson.TypeString, bsoncore.AppendString(nil, t.Format(dateOutputFormat)), nil
}

func (d *date) UnmarshalBSONValue(t bsontype.Type, v []byte) error {
	switch t {
	case bson.TypeString:
		s, _, ok := bsoncore.ReadString(v)
		if !ok {
			return fmt.Errorf("invalid bson string")
		}
		if s == "" {
			return nil
		}
		p, err := time.Parse(dateOutputFormat, s)
		if err != nil {
			return fmt.Errorf("invalid date parse: %s", err)
		}
		*d = date(p)
		return nil
	case bson.TypeDateTime:
		i, _, ok := bsoncore.ReadDateTime(v)
		if !ok {
			return fmt.Errorf("invalid bson datetime")
		}
		*d = date(time.UnixMilli(i))
		return nil
	default:
		return fmt.Errorf("unsupported bson type for date: %v", t)
	}
}

// toDate expects a date as string in the format YYYYMMDD (that is the format
// used by the Federal Revenue in their CSV files).
func toDate(v string) (*date, error) {
	onlyZeros := func(s string) bool {
		v, err := strconv.Atoi(s)
		if err != nil {
			return false
		}
		return v == 0
	}
	if v == "" || onlyZeros(v) {
		return nil, nil
	}
	t, err := time.Parse(dateInputFormat, v)
	if err != nil {
		return nil, fmt.Errorf("error converting %s to Time: %w", v, err)
	}
	d := date(t)
	return &d, nil
}
