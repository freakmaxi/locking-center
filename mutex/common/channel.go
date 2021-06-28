package common

import (
	"strings"
	"sync"
)

type Channel struct {
	Key    string
	Latest *Request

	mutexChan chan bool

	queueLock sync.Mutex
	queueMap  map[string]*Request
}

func NewChannel(key string) *Channel {
	return &Channel{
		Key:       key,
		mutexChan: make(chan bool, 1),
		queueLock: sync.Mutex{},
		queueMap:  make(map[string]*Request),
	}
}

func (c *Channel) pushToQueue(r *Request) {
	c.queueLock.Lock()
	defer c.queueLock.Unlock()

	c.queueMap[r.Id] = r
}

func (c *Channel) pullFromQueue(requestId string) *Request {
	c.queueLock.Lock()
	defer c.queueLock.Unlock()

	request, has := c.queueMap[requestId]
	if !has {
		return nil
	}
	delete(c.queueMap, requestId)

	return request
}

func (c *Channel) Push(r *Request) {
	defer func() { _ = recover() }() // Handle close channel exception

	c.pushToQueue(r)
	c.mutexChan <- true
	c.Latest = c.pullFromQueue(r.Id)

	if c.Latest == nil {
		c.Pull()
	}
}

func (c *Channel) Pull() {
	c.Latest = nil
	if len(c.mutexChan) > 0 { // Avoid deadlock on empty channel
		<-c.mutexChan
	}
}

func (c *Channel) Report() *ChannelReport {
	if len(c.mutexChan) == 0 || c.Latest == nil {
		return nil
	}
	return &ChannelReport{
		Key:     c.Key,
		Current: c.Latest,
	}
}

func (c *Channel) Reset(sourceAddr string) {
	c.queueLock.Lock()
	defer c.queueLock.Unlock()

	resettingRequestIds := make([]string, 0)
	for requestId, request := range c.queueMap {
		if strings.Compare(request.SourceAddr, sourceAddr) != 0 {
			continue
		}
		resettingRequestIds = append(resettingRequestIds, requestId)
	}

	for len(resettingRequestIds) > 0 {
		delete(c.queueMap, resettingRequestIds[0])
		resettingRequestIds = resettingRequestIds[1:]
	}

	if c.Latest != nil && strings.Compare(c.Latest.SourceAddr, sourceAddr) == 0 {
		c.Pull()
	}
}

func (c *Channel) Close() {
	c.queueLock.Lock()
	defer c.queueLock.Unlock()

	c.queueMap = make(map[string]*Request)
	close(c.mutexChan)
}
