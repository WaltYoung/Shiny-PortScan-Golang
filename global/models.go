package global

import "net"

type Task struct {
	IP   net.IP
	Port uint16
}

type Result struct {
	Task Task
	Flag bool
	Err  error
}
