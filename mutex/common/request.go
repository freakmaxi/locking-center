package common

import (
	"net"
	"time"

	"github.com/google/uuid"
)

type Request struct {
	Id    string
	Stamp time.Time

	SourceAddr string
	RemoteAddr net.Addr
}

func NewRequest(sourceAddr string, remoteAddr net.Addr) *Request {
	return &Request{
		Id:         uuid.NewString(),
		Stamp:      time.Now().UTC(),
		SourceAddr: sourceAddr,
		RemoteAddr: remoteAddr,
	}
}
