package network

import (
	"net"
)

// GetLANIPs returns active non-loopback IPv4 addresses found on the system (deduplicated).
func GetLANIPs() []string {
	var ips []string
	seen := make(map[string]bool)

	ifaces, err := net.Interfaces()
	if err != nil {
		return []string{"127.0.0.1"}
	}

	for _, iface := range ifaces {
		// Skip interfaces that are down or loopback
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			// Ensure it's an IPv4 address
			ip4 := ip.To4()
			if ip4 == nil {
				continue
			}

			// Filter out link-local addresses (169.254.x.x)
			if ip4.IsLinkLocalUnicast() {
				continue
			}

			ipStr := ip4.String()
			if !seen[ipStr] {
				seen[ipStr] = true
				ips = append(ips, ipStr)
			}
		}
	}

	if len(ips) == 0 {
		ips = append(ips, "127.0.0.1")
	}

	return ips
}
