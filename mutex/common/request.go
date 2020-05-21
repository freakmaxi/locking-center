package common

import (
	"net"
	"time"
)

type Request struct {
	RemoteAddr net.Addr
	Stamp      time.Time
}
