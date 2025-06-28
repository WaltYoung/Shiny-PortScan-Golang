package main

import (
	"PortScan/core"
	"PortScan/global"
	"PortScan/utils"
	"encoding/hex"
	"fmt"
	"runtime"
)

func init() {
	var err error
	global.SysType = runtime.GOOS
	if global.SysType == "windows" {
		global.InterfaceToDeviceDict, err = utils.CombineInterfaceToDevice()
		if err != nil {
			fmt.Println("Error combining interface to device:", err)
			return
		}
		for iface, device := range global.InterfaceToDeviceDict {
			fmt.Printf("Interface: %s, Device: %s\n", iface, device)
		}
	} else if global.SysType == "linux" {
	} else if global.SysType == "ios" {
	}
	global.SrcIP, err = utils.GetInterfaceIpv4Addr(global.Iface)
	if err != nil {
		fmt.Println(err)
	}
	global.SrcPort, err = utils.GetPort(global.SrcIP)
	fmt.Println("Using interface:", global.Iface, "with IP:", global.SrcIP, "and source port:", global.SrcPort)
	global.SrcMac, err = utils.GetInterfaceMacAddr(global.Iface)
	fmt.Println("Using interface:", global.Iface, "with MAC:", hex.EncodeToString(global.SrcMac), "and source mac:", global.SrcMac)
	global.GatewayIpv4Addr, err = utils.GetGatewayIpv4Addr(global.Iface)
	if err != nil {
		fmt.Println("Error getting gateway IP:", err)
		return
	}
	fmt.Println("Gateway IP:", global.GatewayIpv4Addr)
	global.GatewayMacAddr, err = utils.ArpGetMacAddr(global.SrcIP, global.SrcMac, global.GatewayIpv4Addr, global.Iface)
	if err != nil {
		fmt.Println("Error getting gateway MAC address:", err)
		return
	}
	fmt.Println("Gateway MAC:", hex.EncodeToString(global.GatewayMacAddr))
}

func main() {
	fmt.Print("Enter a target (IP, domain, or CIDR):")
	var target string
	fmt.Scanln(&target)
	ips, tag, err := utils.ParseTarget(target)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	switch tag {
	case "IP":
	case "DOMAIN":
	case "CIDR":
	}
	fmt.Println("IPs:", ips)
	fmt.Print("Enter target ports (22, 80,443 or 1-1024) Default 1-65535:")
	var targetPorts string
	fmt.Scanln(&targetPorts)
	ports, err := utils.ParseTargetPorts(targetPorts)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Ports:", ports)
	fmt.Print("Enter scan mode (syn, tcp) Default syn:")
	var scanMode string
	fmt.Scanln(&scanMode)
	var flag bool
	if scanMode == "tcp" {
		fmt.Println("Using TCP connect scan mode")
		for _, ip := range ips {
			for _, port := range ports {
				flag = core.TCPconnect(ip.String(), port, global.Timeout)
				if flag {
					fmt.Printf("Port %d is open on %s\n", port, ip)
				} else {
					fmt.Printf("Port %d is closed on %s\n", port, ip)
				}
			}
		}
	} else {
		fmt.Println("Using SYN scan mode")
		for _, dstIP := range ips {
			global.DstMac, err = utils.ArpGetMacAddr(global.SrcIP, global.SrcMac, dstIP.String(), global.Iface)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			for _, dstPort := range ports {
				flag, err = core.SYNscan(global.SrcIP, global.SrcPort, dstIP.String(), dstPort, global.Iface, global.SrcMac, global.DstMac)
				if flag {
					fmt.Printf("Port %d is open on %s\n", dstPort, dstIP)
				} else {
					fmt.Printf("Port %d is closed on %s\n", dstPort, dstIP)
					fmt.Println("Error:", err)
				}
			}
		}
	}
}
