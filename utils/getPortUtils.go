package utils

import "net"

func GetPort(srcip string) (port uint16, err error) {
	addr, err := net.ResolveTCPAddr("tcp", srcip+":0")
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}

	defer listener.Close()
	return uint16(listener.Addr().(*net.TCPAddr).Port), nil
}
