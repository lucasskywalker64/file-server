package network

import (
	"net"
	"testing"
)

func TestGetLANIPs(t *testing.T) {
	ips := GetLANIPs()
	if len(ips) == 0 {
		t.Fatalf("expected at least one IP address, got 0")
	}

	seen := make(map[string]bool)
	for _, ipStr := range ips {
		// Verify valid IP format
		ip := net.ParseIP(ipStr)
		if ip == nil {
			t.Errorf("invalid IP string returned: %q", ipStr)
		}

		// Ensure IPv4
		if ip.To4() == nil {
			t.Errorf("expected IPv4 address, got %q", ipStr)
		}

		// Ensure deduplication
		if seen[ipStr] {
			t.Errorf("duplicate IP found: %q", ipStr)
		}
		seen[ipStr] = true
	}
}
