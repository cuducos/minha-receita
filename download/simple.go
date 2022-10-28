package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func simpleDownload(c *http.Client, f file) error {
	h, err := os.Create(f.path)
	if err != nil {
		return fmt.Errorf("could not create %s: %w", f.path, err)
	}
	defer h.Close()
	resp, err := c.Get(f.url)
	if err != nil {
		return fmt.Errorf("error requesting %s: %w", f.url, err)

	}
	defer resp.Body.Close()
	_, err = io.Copy(h, resp.Body)
	if err != nil {
		return fmt.Errorf("error writing to %s: %w", f.path, err)
	}
	return nil
}
