package utils

import (
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"net"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

func GetInterfaceIpv4(iface string) (addr string, mask string, err error) {
	netInterface, err := net.InterfaceByName(iface)
	if err != nil {
		return
	}
	addrs, err := netInterface.Addrs()
	if err != nil {
		return
	}
	var ipv4Addr net.IP
	var ipv4Mask net.IP
	for _, addr := range addrs {
		ipv4Addr = addr.(*net.IPNet).IP.To4()
		if ipv4Addr != nil {
			ipv4Mask = net.IP(addr.(*net.IPNet).Mask).To4()
			break
		}
	}
	if addrs == nil {
		return "", "", fmt.Errorf("interface %s don't have an ipv4 address\n", iface)
	}
	return ipv4Addr.String(), ipv4Mask.String(), nil
}

// GetGatewayForInterface 获取指定网卡的默认网关（跨平台实现）
func GetGatewayForInterface(ifaceName string) (string, error) {
	switch runtime.GOOS {
	case "windows":
		return GetWindowsGateway(ifaceName)
	case "linux", "darwin":
		return GetUnixGateway(ifaceName)
	default:
		return "", fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// GetUnixGateway 获取Unix-like系统（Linux/macOS）的网关
func GetUnixGateway(ifaceName string) (string, error) {
	var cmd *exec.Cmd

	if runtime.GOOS == "linux" {
		cmd = exec.Command("ip", "route", "show", "default")
	} else { // macOS
		cmd = exec.Command("netstat", "-rn")
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("命令执行失败: %v", err)
	}

	return parseUnixGateway(string(output), ifaceName)
}

// parseUnixGateway 解析Unix系统的网关信息
func parseUnixGateway(output, ifaceName string) (string, error) {
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.Contains(line, "default") || strings.Contains(line, "0.0.0.0") {
			// Linux: "default via 192.168.1.1 dev eth0"
			// macOS: "default            192.168.1.1        UGSc           en0"
			if strings.Contains(line, ifaceName) {
				re := regexp.MustCompile(`\b(?:via)?\s*(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\b`)
				matches := re.FindStringSubmatch(line)
				if len(matches) > 1 {
					return matches[1], nil
				}
			}
		}
	}
	return "", fmt.Errorf("网关未找到")
}

// GetWindowsGateway 获取Windows系统的网关
func GetWindowsGateway(ifaceName string) (string, error) {
	// 获取网卡的IP地址列表
	ifaceIPs, err := getInterfaceIPs(ifaceName)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("route", "print", "-4")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("route命令执行失败: %v", err)
	}
	output_utf8, _, _ := transform.Bytes(simplifiedchinese.GBK.NewDecoder(), output) // GBK编码转换为UTF-8编码

	return parseWindowsRoute(string(output_utf8), ifaceIPs)
}

// getInterfaceIPs 获取指定网卡的所有IPv4地址
func getInterfaceIPs(ifaceName string) ([]string, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return nil, fmt.Errorf("找不到网卡: %v", err)
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, fmt.Errorf("获取IP地址失败: %v", err)
	}

	var ips []string
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}

		if ipNet.IP.To4() != nil {
			ips = append(ips, ipNet.IP.String())
		}
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("网卡没有IPv4地址")
	}

	return ips, nil
}

// parseWindowsRoute 解析Windows路由表
func parseWindowsRoute(output string, ifaceIPs []string) (string, error) {
	lines := strings.Split(output, "\r\n")
	inTable := false

	for _, line := range lines {
		// 找到IPv4路由表开始位置
		if strings.Contains(line, "IPv4 路由表") || strings.Contains(line, "IPv4 Route Table") {
			inTable = true
			continue
		}

		if !inTable {
			continue
		}

		// 匹配默认路由行: 0.0.0.0 0.0.0.0
		if strings.HasPrefix(line, "          0.0.0.0          0.0.0.0") {
			fields := strings.Fields(line)
			if len(fields) < 5 {
				continue
			}

			gateway := fields[2]
			interfaceIP := fields[3]

			// 检查接口IP是否属于该网卡
			for _, ip := range ifaceIPs {
				if ip == interfaceIP {
					return gateway, nil
				}
			}
		}
	}

	return "", fmt.Errorf("默认网关未找到")
}
