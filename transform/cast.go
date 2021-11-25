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

func toTime(v string) (*time.Time, error) {
	if v == "" {
		return nil, nil
	}

	t, err := time.Parse("20060102", v)
	if err != nil {
		return nil, fmt.Errorf("error converting %s to Time: %w", v, err)
	}

	return &t, nil
}
