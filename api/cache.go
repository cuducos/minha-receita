package api

import (
	"sync"
	"time"
)

const expiresIn = 12 // hours

type cacheEntry struct {
	data      []byte
	expiresOn time.Time
}

type cache struct {
	entries map[string]cacheEntry
	mutex   sync.Mutex
}

func (c *cache) read(k string) ([]byte, bool) {
	e, ok := c.entries[k]
	if !ok {
		return []byte{}, false
	}
	if time.Now().After(e.expiresOn) {
		c.mutex.Lock()
		delete(c.entries, k)
		c.mutex.Unlock()
		return []byte{}, false
	}
	return e.data, true
}

func (c *cache) save(k string, b []byte) {
	c.mutex.Lock()
	c.entries[k] = cacheEntry{b, time.Now().Add(expiresIn * time.Hour)}
	c.mutex.Unlock()
}

func newCache() cache {
	return cache{entries: make(map[string]cacheEntry)}
}
