package utils

import (
	"PortScan/global"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"net"
	"time"
)

func GetInterfaceMacAddr(iface string) ([]byte, error) {
	netInterface, err := net.InterfaceByName(iface)
	if err != nil {
		return nil, err
	}
	return netInterface.HardwareAddr, nil
}

func ArpGetMacAddr(srcIP string, srcMac []byte, dstIP string, iface string) ([]byte, error) {
	handle, err := pcap.OpenLive(global.InterfaceToDeviceDict[iface], 65536, true, pcap.BlockForever)
	if err != nil {
		return nil, fmt.Errorf("failed to open device: %v", err)
	}
	defer handle.Close()
	eth := &layers.Ethernet{
		SrcMAC:       srcMac,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}
	arp := &layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPRequest,
		SourceHwAddress:   srcMac,
		SourceProtAddress: net.ParseIP(srcIP).To4(),
		DstHwAddress:      []byte{0, 0, 0, 0, 0, 0},
		DstProtAddress:    net.ParseIP(dstIP).To4(),
	}
	buffer := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{}
	err = gopacket.SerializeLayers(buffer, opts, eth, arp)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize ARP packet: %v", err)
	}
	err = handle.WritePacketData(buffer.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to send ARP packet: %v", err)
	}
	fmt.Println("ARP request sent for IP:", dstIP)
	// 监听并处理ARP响应包
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for {
		select {
		case packet := <-packetSource.Packets():
			arpLayer := packet.Layer(layers.LayerTypeARP)
			if arpLayer != nil {
				arpContent, _ := arpLayer.(*layers.ARP)
				if arpContent.Operation == layers.ARPReply && net.IP(arpContent.SourceProtAddress).Equal(net.ParseIP(dstIP)) {
					fmt.Printf("ARP responce Received: %s -> %s\n", net.IP(arpContent.SourceProtAddress), net.HardwareAddr(arpContent.SourceHwAddress))
					return arpContent.SourceHwAddress, nil
				}
			}
		case <-time.After(global.Timeout):
			return nil, fmt.Errorf("No ARP response received for IP: %s (timeout)", dstIP)
		}
	}
}
