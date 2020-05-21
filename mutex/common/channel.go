package common

type Channel struct {
	Key    string
	Latest *Request

	mutexChan chan *Request
}

func NewChannel(key string) *Channel {
	return &Channel{
		Key:       key,
		mutexChan: make(chan *Request, 1),
	}
}

func (c *Channel) Push(r *Request) {
	c.mutexChan <- r
	c.Latest = r
}

func (c *Channel) Pull() {
	if len(c.mutexChan) == 0 { // Avoid deadlock on empty channel
		return
	}
	<-c.mutexChan
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

func (c *Channel) Close() {
	close(c.mutexChan)

	for {
		select {
		case _, more := <-c.mutexChan:
			if !more {
				return
			}
		}
	}
}
