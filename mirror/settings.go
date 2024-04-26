package mirror

import (
	"fmt"
	"os"
	"strings"
)

type settings struct {
	accessKey       string
	secretAccessKey string
	region          string
	endpointURL     string
	bucket          string
	publicDomain    string
}

func newSettings() (settings, error) {
	var m []string
	load := func(key string) string {
		v := os.Getenv(key)
		if v == "" {
			m = append(m, key)
		}
		return v
	}

	s := settings{
		accessKey:       load("AWS_ACCESS_KEY_ID"),
		secretAccessKey: load("AWS_SECRET_ACCESS_KEY"),
		region:          load("AWS_DEFAULT_REGION"),
		endpointURL:     load("ENDPOINT_URL"),
		bucket:          load("BUCKET"),
		publicDomain:    load("PUBLIC_DOMAIN"),
	}

	if len(m) > 0 {
		return settings{}, fmt.Errorf("missing environment variable(s): %s", strings.Join(m, ", "))
	}
	return s, nil

}
