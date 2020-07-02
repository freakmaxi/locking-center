package common

import (
	"net"
	"time"
)

type Request struct {
	SourceAddr string
	RemoteAddr net.Addr
	Stamp      time.Time
}
