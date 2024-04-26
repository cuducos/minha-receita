// package mirror is  mirror of data from the Federal Revenue CNPJ, in addition
// to the executables. This repository provides a web interface for the bucket
// of these files. To sync, use AWS CLI:
// $ aws s3 --endpoint $ENDPOINT_URL sync ~/data s3://minhareceita --acl public-read

package mirror

import (
	"fmt"
	"log"
	"net/http"
)

func startServer(c *Cache, p string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if c.isExpired() {
			if err := c.refresh(); err != nil {
				log.Output(1, fmt.Sprintf("Error loading files: %s", err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		if r.Header.Get("Accept") == "application/json" {
			w.Write(c.JSON)
		} else {
			w.Write(c.HTML)
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
	log.Output(1, fmt.Sprintf("Server listening on http://0.0.0.0%s", p))
	http.ListenAndServe(p, nil)
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
	startServer(c, p)
	return nil
}
