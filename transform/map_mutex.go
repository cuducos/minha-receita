package transform

import (
	"sync"
)

type mapMutex struct {
	keys  map[string]struct{}
	mutex sync.Mutex
}

func (m *mapMutex) lock(k string) {
	exists := func(m *mapMutex, k string) bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()
		_, exists := m.keys[k]
		return exists
	}

	for {
		if !exists(m, k) {
			m.mutex.Lock()
			defer m.mutex.Unlock()
			m.keys[k] = struct{}{}
			return
		}
	}
}

func (m *mapMutex) unlock(k string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.keys, k)
}

func newMapMutex() mapMutex {
	return mapMutex{keys: make(map[string]struct{})}
}
