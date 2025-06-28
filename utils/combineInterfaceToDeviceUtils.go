package utils

import (
	"PortScan/global"
	"github.com/google/gopacket/pcap"
	"net"
)

func CombineInterfaceToDevice() (map[string]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	devices, err := pcap.FindAllDevs()
	if err != nil {
		panic(err)
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, device := range devices {
			if len(device.Addresses) == 0 {
				continue
			}
			for _, addr := range addrs {
				if device.Addresses[0].IP.String() == addr.(*net.IPNet).IP.String() {
					global.InterfaceToDeviceDict[iface.Name] = device.Name
					break
				}
			}
		}
	}
	return global.InterfaceToDeviceDict, err
}
