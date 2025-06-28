package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

func ParseTarget(target string) ([]net.IP, string, error) {
	ip := net.ParseIP(target)
	if ip != nil {
		return []net.IP{ip}, "IP", nil
	}

	ips, err := net.LookupIP(target)
	if err == nil {
		return ips, "DOMAIN", nil
	}

	_, ipNet, err := net.ParseCIDR(target)
	if err == nil {
		return expandCIDR(ipNet), "CIDR", nil
	}

	return nil, "", fmt.Errorf("invalid target: %s", target)
}

func expandCIDR(ipNet *net.IPNet) []net.IP {
	start := ipv4ToUint32(ipNet.IP)
	ones, bits := ipNet.Mask.Size()
	ipCount := 1 << (bits - ones)
	var ips []net.IP
	for i := 1; i < ipCount-1; i++ {
		ip := uint32ToIPv4(start + uint32(i))
		if ip != nil {
			ips = append(ips, ip)
		}
	}
	return ips
}

func ipv4ToUint32(ip net.IP) uint32 {
	ip = ip.To4()
	if ip == nil {
		return 0
	}
	var ret uint32
	err := binary.Read(bytes.NewBuffer(ip), binary.BigEndian, &ret)
	if err != nil {
		return 0
	}
	return ret
}

func uint32ToIPv4(IPv4Int uint32) net.IP {
	return net.ParseIP(fmt.Sprintf("%d.%d.%d.%d", byte(IPv4Int>>24), byte(IPv4Int>>16), byte(IPv4Int>>8), byte(IPv4Int)))
}
