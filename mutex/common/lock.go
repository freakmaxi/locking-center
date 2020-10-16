package common

import (
	"net"
	"sort"
	"strings"
	"sync"
	"time"
)

type Lock struct {
	mutex    *sync.Mutex
	channels map[string]*Channel
}

func NewLock() *Lock {
	return &Lock{
		mutex:    &sync.Mutex{},
		channels: make(map[string]*Channel),
	}
}

func (l *Lock) channel(key string) *Channel {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if _, has := l.channels[key]; !has {
		l.channels[key] = NewChannel(key)
	}

	return l.channels[key]
}

func (l *Lock) Lock(key string, sourceAddr string, remoteAddr net.Addr) (locked bool) {
	defer func() {
		if r := recover(); r != nil {
			locked = false
		}
	}() // Handle in case of reset
	l.channel(key).Push(&Request{
		SourceAddr: sourceAddr,
		RemoteAddr: remoteAddr,
		Stamp:      time.Now().UTC(),
	})
	return true
}

func (l *Lock) Unlock(key string) {
	l.channel(key).Pull()
}

func (l *Lock) ResetByKey(key string) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if _, has := l.channels[key]; !has {
		return false
	}

	l.channels[key].Close()
	delete(l.channels, key)

	return true
}

func (l *Lock) ResetBySource(sourceAddr string) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	reset := false
	resettingKeys := make([]string, 0)

	for key, channel := range l.channels {
		if channel.Latest == nil {
			continue
		}

		if strings.Compare(channel.Latest.SourceAddr, sourceAddr) != 0 {
			continue
		}

		resettingKeys = append(resettingKeys, key)
		reset = true
	}

	for len(resettingKeys) > 0 {
		key := resettingKeys[0]

		l.channels[key].Close()
		delete(l.channels, key)

		resettingKeys = resettingKeys[1:]
	}

	return reset
}

func (l *Lock) Keys() ChannelReports {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	reports := make(ChannelReports, 0)
	for k := range l.channels {
		report := l.channels[k].Report()
		if report == nil {
			continue
		}
		reports = append(reports, report)
	}
	sort.Sort(reports)

	return reports
}
