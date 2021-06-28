package common

import (
	"net"
	"sort"
	"sync"
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
	l.channel(key).Push(NewRequest(sourceAddr, remoteAddr))
	return true
}

func (l *Lock) Unlock(key string) {
	l.channel(key).Pull()
}

func (l *Lock) ResetByKey(key string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	channel, has := l.channels[key]
	if !has {
		return
	}
	channel.Close()
	delete(l.channels, key)
}

func (l *Lock) ResetBySource(sourceAddr string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	for _, channel := range l.channels {
		channel.Reset(sourceAddr)
	}
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
