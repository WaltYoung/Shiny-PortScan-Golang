package utils

import (
	"fmt"
	"strings"
)

func ParseTargetPorts(targetPorts string) ([]uint16, error) {
	var ports []uint16
	if targetPorts == "" {
		for i := 1; i <= 65535; i++ {
			ports = append(ports, uint16(i))
		}
	} else if strings.Contains(targetPorts, "-") {
		parts := strings.Split(targetPorts, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid port range: %s", targetPorts)
		}
		var start, end int
		fmt.Sscanf(parts[0], "%d", &start)
		fmt.Sscanf(parts[1], "%d", &end)
		if start < 1 || end > 65535 || start > end {
			return nil, fmt.Errorf("port range out of bounds: %s", targetPorts)
		}
		for i := start; i <= end; i++ {
			ports = append(ports, uint16(i))
		}
	} else if strings.Contains(targetPorts, ",") {
		parts := strings.Split(targetPorts, ",")
		for _, part := range parts {
			var port uint16
			fmt.Sscanf(part, "%d", &port)
			ports = append(ports, port)
		}
	} else {
		var port uint16
		fmt.Sscanf(targetPorts, "%d", &port)
		if port < 1 || port > 65535 {
			return nil, fmt.Errorf("invalid target ports: %s", targetPorts)
		}
		ports = append(ports, port)
	}
	return ports, nil
}
