package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
)

func ParseTarget(target string) ([]net.IP, error) {
	ip := net.ParseIP(target)
	if ip != nil {
		return []net.IP{ip}, nil
	}

	ips, err := net.LookupIP(target)
	if err == nil {
		return ips, nil
	}

	_, ipNet, err := net.ParseCIDR(target)
	if err == nil {
		return expandCIDR(ipNet), nil
	}

	return nil, fmt.Errorf("invalid target: %s", target)
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

// 子网掩码地址转换为网络位长度，如 255.255.255.0 对应的网络位长度为 24
func SubNetMaskToLen(netmask string) (int, error) {
	ipSplitArr := strings.Split(netmask, ".")
	if len(ipSplitArr) != 4 {
		return 0, fmt.Errorf("netmask:%v is not valid, pattern should like: 255.255.255.0", netmask)
	}
	ipv4MaskArr := make([]byte, 4)
	for i, value := range ipSplitArr {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return 0, fmt.Errorf("ipMaskToInt call strconv.Atoi error:[%v] string value is: [%s]", err, value)
		}
		if intValue > 255 {
			return 0, fmt.Errorf("netmask cannot greater than 255, current value is: [%s]", value)
		}
		ipv4MaskArr[i] = byte(intValue)
	}

	ones, _ := net.IPv4Mask(ipv4MaskArr[0], ipv4MaskArr[1], ipv4MaskArr[2], ipv4MaskArr[3]).Size()
	return ones, nil
}

// 网络位长度转换为子网掩码地址，如 24 对应的子网掩码地址为 255.255.255.0
func LenToSubNetMask(subnet int) string {
	var buff bytes.Buffer
	for i := 0; i < subnet; i++ {
		buff.WriteString("1")
	}
	for i := subnet; i < 32; i++ {
		buff.WriteString("0")
	}
	masker := buff.String()
	a, _ := strconv.ParseUint(masker[:8], 2, 64)
	b, _ := strconv.ParseUint(masker[8:16], 2, 64)
	c, _ := strconv.ParseUint(masker[16:24], 2, 64)
	d, _ := strconv.ParseUint(masker[24:32], 2, 64)
	resultMask := fmt.Sprintf("%v.%v.%v.%v", a, b, c, d)
	return resultMask
}

func IsIPInSubnet(cidr string, ip net.IP) bool {
	_, subnet, _ := net.ParseCIDR(cidr)
	return subnet.Contains(ip)
}
