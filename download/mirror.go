package download

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const mirrorURL = "https://mirror.minhareceita.org"

type MirrorFile struct {
	URL  string `json:"url"`
	Size uint   `json:"size"`
}

type MirrorGroup struct {
	Name string       `json:"name"`
	URLs []MirrorFile `json:"urls"`
}

type MirrorResponse struct {
	Data []MirrorGroup `json:"data"`
}

func getMirrorURLs(t string) ([]string, error) {
	c := &http.Client{}
	req, err := http.NewRequest("GET", mirrorURL, nil)
	if err != nil {
		return []string{}, fmt.Errorf("error creating request for mirror: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	r, err := c.Do(req)
	if err != nil {
		return []string{}, fmt.Errorf("error sending request to mirror: %w", err)
	}
	defer r.Body.Close()
	var gs MirrorResponse
	if err = json.NewDecoder(r.Body).Decode(&gs); err != nil {
		return []string{}, fmt.Errorf("error decoding response body: %w", err)
	}
	var urls []string
	var opts []string
	for _, g := range gs.Data {
		if g.Name == t {
			for _, u := range g.URLs {
				if u.Size > 0 {
					urls = append(urls, u.URL)
				}
			}
			break
		}
		opts = append(opts, g.Name)
	}
	if len(urls) == 0 {
		return []string{}, fmt.Errorf("unknown mirror identifier `%s`, options are: %s", t, strings.Join(opts, ", "))
	}
	return urls, nil
}
