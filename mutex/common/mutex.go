package common

import (
	"net"
	"sort"
	"sync"
	"time"
)

type Mutex struct {
	mutex    *sync.Mutex
	channels map[string]*Channel
}

func NewLock() *Mutex {
	return &Mutex{
		mutex:    &sync.Mutex{},
		channels: make(map[string]*Channel),
	}
}

func (m *Mutex) channel(key string) *Channel {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, has := m.channels[key]; !has {
		m.channels[key] = NewChannel(key)
	}

	return m.channels[key]
}

func (m *Mutex) Lock(key string, remoteAddr net.Addr) (locked bool) {
	defer func() {
		if r := recover(); r != nil {
			locked = false
		}
	}() // Handle in case of reset
	m.channel(key).Push(&Request{
		RemoteAddr: remoteAddr,
		Stamp:      time.Now().UTC(),
	})

	locked = true
	return
}

func (m *Mutex) Unlock(key string) {
	m.channel(key).Pull()
}

func (m *Mutex) Reset(key string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, has := m.channels[key]; !has {
		return
	}

	m.channels[key].Close()
	delete(m.channels, key)
}

func (m *Mutex) Keys() ChannelReports {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	reports := make(ChannelReports, 0)
	for k := range m.channels {
		report := m.channels[k].Report()
		if report == nil {
			continue
		}
		reports = append(reports, report)
	}
	sort.Sort(reports)

	return reports
}
