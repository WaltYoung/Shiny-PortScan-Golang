package core

import (
	"PortScan/global"
	"PortScan/utils"
	"fmt"
	"sync"
)

func Worker(scanMode string, inSubnet bool, tasks <-chan global.Task, results chan<- global.Result, wg *sync.WaitGroup) {
	defer wg.Done()
	if scanMode == "tcp" {
		for task := range tasks {
			flag := TCPconnect(task.IP.String(), task.Port, global.Timeout)
			results <- global.Result{
				Task: global.Task{
					IP:   task.IP,
					Port: task.Port,
				},
				Flag: flag,
				Err:  nil,
			}
		}
	} else {
		var err error
		var flag bool
		if inSubnet {
			for task := range tasks {
				global.DstMac, err = utils.ArpGetMacAddr(global.SrcIP, global.SrcMac, task.IP.String(), global.Iface)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				flag, err = SYNscan(global.SrcIP, global.SrcPort, task.IP.String(), task.Port, global.Iface, global.SrcMac, global.DstMac)
				results <- global.Result{
					Task: global.Task{
						IP:   task.IP,
						Port: task.Port,
					},
					Flag: flag,
					Err:  err,
				}
			}
		} else {
			global.DstMac = global.GatewayMacAddr
			for task := range tasks {
				flag, err = SYNscan(global.SrcIP, global.SrcPort, task.IP.String(), task.Port, global.Iface, global.SrcMac, global.DstMac)
				results <- global.Result{
					Task: global.Task{
						IP:   task.IP,
						Port: task.Port,
					},
					Flag: flag,
					Err:  err,
				}
			}
		}
	}
}
