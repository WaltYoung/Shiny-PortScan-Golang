package utils

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"net"
)

func GetInterfaceIpv4Addr(iface string) (addr string, err error) {
	netInterface, err := net.InterfaceByName(iface)
	if err != nil {
		return
	}
	addrs, err := netInterface.Addrs()
	if err != nil {
		return
	}
	var ipv4Addr net.IP
	for _, addr := range addrs {
		ipv4Addr = addr.(*net.IPNet).IP.To4()
		if ipv4Addr != nil {
			break
		}
	}
	if addrs == nil {
		return "", fmt.Errorf("interface %s don't have an ipv4 address\n", iface)
	}
	return ipv4Addr.String(), nil
}

func GetGatewayIpv4Addr(ifaceName string) (string, error) {
	// 获取指定网卡
	link, err := netlink.LinkByName(ifaceName)
	if err != nil {
		return "", fmt.Errorf("获取网卡失败: %v", err)
	}

	// 获取Ipv4的所有路由
	routes, err := netlink.RouteList(link, 2) // 2 表示 AF_INET (IPv4)
	if err != nil {
		return "", fmt.Errorf("获取路由失败: %v", err)
	}

	// 查找默认网关（目的地址为 nil 的路由）
	for _, route := range routes {
		if route.Dst == nil {
			if route.Gw != nil {
				return route.Gw.String(), nil // 找到网关 IP
			}
		}
	}

	return "", fmt.Errorf("未找到默认网关")
}
