package common

import (
	"sort"
	"sync"
)

type Mutex struct {
	mutex    *sync.Mutex
	channels map[string]chan bool
}

func NewLock() *Mutex {
	return &Mutex{
		mutex:    &sync.Mutex{},
		channels: make(map[string]chan bool),
	}
}

func (m *Mutex) channel(key string) chan bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, has := m.channels[key]; !has {
		m.channels[key] = make(chan bool, 1)
	}

	return m.channels[key]
}

func (m *Mutex) Lock(key string) {
	defer recover() // Handle in case of reset
	m.channel(key) <- true
}

func (m *Mutex) Unlock(key string) {
	c := m.channel(key)
	if len(c) == 0 { // Avoid deadlock on empty channel
		return
	}
	<-m.channel(key)
}

func (m *Mutex) Reset(key string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, has := m.channels[key]; !has {
		return
	}

	close(m.channels[key])
	delete(m.channels, key)
}

func (m *Mutex) Keys() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	keys := make([]string, 0)
	for k := range m.channels {
		if len(m.channels[k]) == 0 {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}
