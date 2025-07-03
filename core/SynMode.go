package core

import (
	"PortScan/global"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"net"
	"time"
)

func SYNscan(srcIP string, srcPort uint16, dstIP string, dstPort uint16, iface string, srcMac []byte, dstMac []byte) (bool, error) {
	handle, err := pcap.OpenLive(global.InterfaceToDeviceDict[iface], 65536, true, pcap.BlockForever)
	if err != nil {
		return false, fmt.Errorf("failed to open device: %v", err)
	}
	defer handle.Close()
	eth := &layers.Ethernet{
		SrcMAC:       srcMac,
		DstMAC:       dstMac,
		EthernetType: layers.EthernetTypeIPv4,
	}
	ip4 := &layers.IPv4{
		Version:  4,
		SrcIP:    net.ParseIP(srcIP).To4(),
		DstIP:    net.ParseIP(dstIP).To4(),
		Protocol: layers.IPProtocolTCP,
	}
	tcp := &layers.TCP{
		SrcPort: layers.TCPPort(srcPort),
		DstPort: layers.TCPPort(dstPort),
		SYN:     true,
		Window:  14600,
	}
	tcp.SetNetworkLayerForChecksum(ip4)
	buffer := gopacket.NewSerializeBuffer()
	opt := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	err = gopacket.SerializeLayers(buffer, opt, eth, ip4, tcp)
	if err != nil {
		return false, fmt.Errorf("failed to create serialize buffer: %v", err)
	}
	err = handle.WritePacketData(buffer.Bytes())
	if err != nil {
		return false, fmt.Errorf("failed to send SYN packet: %v", err)
	}
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for {
		select {
		case packet := <-packetSource.Packets():
			if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
				ipContent, _ := ipLayer.(*layers.IPv4)
				if !ipContent.SrcIP.Equal(net.ParseIP(dstIP)) || !ipContent.DstIP.Equal(net.ParseIP(srcIP)) {
					continue
				}
			}
			if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
				tcpContent, _ := tcpLayer.(*layers.TCP)
				if tcpContent.DstPort == layers.TCPPort(srcPort) && tcpContent.SYN && tcpContent.ACK {
					return true, nil
				}
				if tcpContent.RST {
					return false, nil
				}
			}
		case <-time.After(global.Timeout):
			return false, fmt.Errorf("timeout waiting for response from %s:%d", dstIP, dstPort)
		}
	}
}
