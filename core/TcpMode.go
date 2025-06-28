package core

import (
	"fmt"
	"net"
	"time"
)

func TCPconnect(ip string, port uint16, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
