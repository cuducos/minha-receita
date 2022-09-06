package api

import (
	"testing"
	"time"
)

func TestCacheRead(t *testing.T) {
	k := "answer"
	v := []byte("42")
	future := time.Now().Add(1 * time.Hour)
	past := time.Now().Add(-1 * time.Hour)
	t.Run("existing value", func(t *testing.T) {
		c := newCache()
		c.entries[k] = cacheEntry{data: v, expiresOn: future}
		got, ok := c.read(k)
		if !ok {
			t.Errorf("expected to be able to read for the cache, but got ok = %t", ok)
		}
		if string(got) != string(v) {
			t.Errorf("expected to read %s from cache, got %s", string(v), string(got))
		}
	})
	t.Run("expired value", func(t *testing.T) {
		c := newCache()
		c.entries[k] = cacheEntry{data: v, expiresOn: past}
		_, ok := c.read(k)
		if ok {
			t.Errorf("expected not to be able to read for the cache, but got ok = %t", ok)
		}
	})
	t.Run("non-existing value", func(t *testing.T) {
		c := newCache()
		_, ok := c.read(k)
		if ok {
			t.Errorf("expected not to be able to read for the cache, but got ok = %t", ok)
		}
	})
}

func TestCacheSave(t *testing.T) {
	k := "answer"
	v := []byte("42")
	t.Run("creates new key", func(t *testing.T) {
		c := newCache()
		c.save(k, v)
		got, ok := c.entries[k]
		if !ok {
			t.Errorf("expected key to be created in the cache, but got ok = %t", ok)
		}
		if string(got.data) != string(v) {
			t.Errorf("expected to read %s from cache, got %s", string(v), string(got.data))
		}
	})
	t.Run("updates existing key", func(t *testing.T) {
		c := newCache()
		c.save(k, []byte("not me"))
		c.save(k, v)
		got, ok := c.entries[k]
		if !ok {

			t.Errorf("expected key to be created in the cache, but got ok = %t", ok)
		}
		if string(got.data) != string(v) {
			t.Errorf("expected to read %s from cache, got %s", string(v), string(got.data))
		}
	})
	t.Run("sets the expires on accordingly", func(t *testing.T) {
		c := newCache()
		c.save(k, v)
		got, ok := c.entries[k]
		if !ok {
			t.Errorf("expected key to be created in the cache, but got ok = %t", ok)
		}
		if got.expiresOn.Before(time.Now()) {
			t.Errorf("expected expires on to be in the future, got %v", got.expiresOn)
		}
	})
}
