// package mirror is  mirror of data from the Federal Revenue CNPJ, in addition
// to the executables. This repository provides a web interface for the bucket
// of these files. To sync, use AWS CLI:
// $ aws s3 --endpoint $ENDPOINT_URL sync ~/data s3://minhareceita --acl public-read

package mirror

import (
	"fmt"
	"log/slog"
	"net/http"
)

func startServer(c *Cache, p string) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if c.isExpired() {
			if err := c.refresh(); err != nil {
				slog.Error("Error loading files", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}
		var b []byte
		if r.Header.Get("Accept") == "application/json" {
			b = c.JSON
		} else {
			b = c.HTML
		}
		if _, err := w.Write(b); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error("could not write response", "error", err)
		}
	})
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead && r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	p = fmt.Sprintf(":%s", p)
	slog.Info(fmt.Sprintf("Server listening on http://0.0.0.0%s", p))
	return http.ListenAndServe(p, nil)
}

func Mirror(p string) error {
	s, err := newSettings()
	if err != nil {
		return fmt.Errorf("error loading mirror settings: %w", err)
	}
	c, err := newCache(s)
	if err != nil {
		return fmt.Errorf("error loading mirror cache: %w", err)
	}
	return startServer(c, p)
}
