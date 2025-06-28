package global

import "time"

const Timeout = 2 * time.Second
const Iface = "WLAN"

var SysType string
var SrcIP string
var SrcPort uint16
var SrcMac []byte
var DstMac []byte
var InterfaceToDeviceDict map[string]string
var GatewayIpv4Addr string
var GatewayMacAddr []byte

func init() {
	InterfaceToDeviceDict = make(map[string]string)
}
