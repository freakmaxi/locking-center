package common

import (
	"net"
	"strings"
)

func ExtractSourceAddr(conn net.Conn) string {
	sourceAddr := conn.RemoteAddr().String()
	idxColon := strings.Index(sourceAddr, ":")
	if idxColon > -1 {
		sourceAddr = sourceAddr[:idxColon]
	}
	return sourceAddr
}
