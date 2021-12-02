// These functions allow the project to accomplish two things:
// * convert strings from the CSV files to other formats (e.g. int, float32,
//   time.Time);
// * differentiate empty values (such as 0 for int) from missing values.
//
// This is achieved using pointers, so we have nil as a marker for missing
// value.
// Since our use case involves serving this data in JSON file, this is crucial
// so we can use `null` when there is no value, and "0" when the value of an
// integer, for example, is 0.
package transform

import (
	"fmt"
	"strconv"
	"strings"
	"time"
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

// toDate expects a date as string in the format YYYYMMDD (that is the format
// used by the Federal Revenue in their CSV files).
func toDate(v string) (date, error) {
	if v == "" {
		return date{nil}, nil
	}

	t, err := time.Parse(dateInputFormat, v)
	if err != nil {
		return date{nil}, fmt.Errorf("error converting %s to Time: %w", v, err)
	}

	return date{&t}, nil
}
