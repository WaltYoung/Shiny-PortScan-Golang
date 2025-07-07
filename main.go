package main

import (
	"PortScan/core"
	"PortScan/global"
	"PortScan/utils"
	"encoding/hex"
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"
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
	global.SrcIP, global.SrcSubnetMask, err = utils.GetInterfaceIpv4(global.Iface)
	if err != nil {
		fmt.Println(err)
	}
	global.SrcPort, err = utils.GetPort(global.SrcIP)
	fmt.Println("Using interface:", global.Iface, "with IP:", global.SrcIP, "Mask:", global.SrcSubnetMask, "and source port:", global.SrcPort)
	global.SrcMac, err = utils.GetInterfaceMacAddr(global.Iface)
	fmt.Println("Using interface:", global.Iface, "with MAC:", hex.EncodeToString(global.SrcMac), "and source mac:", global.SrcMac)
	global.GatewayIpv4Addr, err = utils.GetGatewayForInterface(global.Iface)
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
	ips, err := utils.ParseTarget(target)
	if err != nil {
		fmt.Println("Error:", err)
		return
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
	if scanMode == "tcp" {
		fmt.Println("Using TCP connect scan mode")
	} else {
		fmt.Println("Using SYN scan mode")
	}
	start := time.Now()
	var inSubnet bool
	maskLen, err := utils.SubNetMaskToLen(global.SrcSubnetMask)
	inSubnet = utils.IsIPInSubnet(global.SrcIP+"/"+strconv.Itoa(maskLen), ips[0])
	var wg sync.WaitGroup
	taskCount := len(ips) * len(ports)
	tasks := make(chan global.Task, taskCount)
	results := make(chan global.Result, taskCount)
	// 启动协程池
	for i := 0; i < global.WorkNum; i++ {
		wg.Add(1)
		go core.Worker(scanMode, inSubnet, tasks, results, &wg)
	}
	// 向协程池提交任务
	for _, ip := range ips {
		for _, port := range ports {
			tasks <- global.Task{
				IP:   ip,
				Port: port,
			}
		}
	}
	close(tasks)
	wg.Wait()      // 等待所有 Worker 协程完成
	close(results) // 确保所有结果写入后关闭通道
	for res := range results {
		if res.Flag {
			fmt.Printf("Port %d is open on %s\n", res.Task.Port, res.Task.IP)
		} else {
			//fmt.Printf("Port %d is closed on %s\n", res.Task.Port, res.Task.IP)
			if res.Err != nil {
				fmt.Printf("Error: %v Port %d is closed on %s\n", res.Err, res.Task.Port, res.Task.IP)
			}
		}
	}
	end := time.Now()
	fmt.Printf("Scan completed in %v seconds\n", end.Sub(start).Seconds())
}
