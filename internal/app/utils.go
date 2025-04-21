package app

import (
	"net"
	"strings"
)

// support three formats addr:
// 1. ipaddress(10.0.0.2)
// 2. ipaddr:port(10.0.0.2:8080)
// 3. file patch(/var/run/test.sock)
func networkType(addr string) string {
	addr = strings.Split(addr, ":")[0]
	if net.ParseIP(addr) != nil {
		return "tcp"
	} else {
		return "unix"
	}
}
