package transform

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type cache struct {
	dir   string
	mutex sync.Mutex
}

func (c *cache) pathFor(k string) string {
	if len(k) != 14 {
		return filepath.Join(c.dir, k)
	}
	return filepath.Join(c.dir, k[0:2], k[2:5], k[5:8], k[8:12], k[12:14])
}

func (c *cache) check(k string) (bool, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	_, err := os.Stat(c.pathFor(k))
	if err == nil {
		return true, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("error reading cache file %s: %w", c.pathFor(k), err)
	}
	if err := c.save(k); err != nil {
		return false, err
	}
	return false, nil
}

func (c *cache) save(k string) error {
	p := c.pathFor(k)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return fmt.Errorf("could not create cache directory %s: %w", filepath.Dir(p), err)
	}
	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("error creating cache file %s: %w", p, err)
	}
	defer f.Close()
	return nil
}
