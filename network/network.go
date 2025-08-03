package network

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

// Common errors returned by this package.
var (
	ErrInvalidIPAddress = errors.New("invalid IP address")
	ErrInvalidCIDR      = errors.New("invalid CIDR notation")
	ErrInvalidRange     = errors.New("invalid IP range format")
	ErrNoLocalNetwork   = errors.New("no local network found")
)

// IPRange represents an IP range
type IPRange struct {
	StartIP net.IP
	EndIP   net.IP
}

// ParseIPRange parses different IP range formats
func ParseIPRange(ipRange string) (*IPRange, error) {
	// Check if it's CIDR notation (e.g., 192.168.1.0/24)
	if strings.Contains(ipRange, "/") {
		return parseCIDR(ipRange)
	}
	
	// Check if it's range notation (e.g., 192.168.1.1-192.168.1.255)
	if strings.Contains(ipRange, "-") {
		return parseRange(ipRange)
	}
	
	// Single IP
	ip := net.ParseIP(ipRange)
	if ip == nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidIPAddress, ipRange)
	}
	
	return &IPRange{
		StartIP: ip,
		EndIP:   ip,
	}, nil
}

// parseCIDR parses CIDR notation.
func parseCIDR(cidr string) (*IPRange, error) {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidCIDR, cidr)
	}
	
	// Calculate start and end IPs
	startIP := ip.Mask(ipNet.Mask)
	endIP := make(net.IP, len(startIP))
	copy(endIP, startIP)
	
	// Calculate broadcast address
	for i := 0; i < len(startIP); i++ {
		endIP[i] = startIP[i] | ^ipNet.Mask[i]
	}
	
	return &IPRange{
		StartIP: startIP,
		EndIP:   endIP,
	}, nil
}

// parseRange parses range notation (e.g., 192.168.1.1-192.168.1.255).
func parseRange(rangeStr string) (*IPRange, error) {
	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("%w: %s", ErrInvalidRange, rangeStr)
	}
	
	startIP := net.ParseIP(strings.TrimSpace(parts[0]))
	endIP := net.ParseIP(strings.TrimSpace(parts[1]))
	
	if startIP == nil || endIP == nil {
		return nil, fmt.Errorf("%w: invalid IP addresses in range %s", ErrInvalidRange, rangeStr)
	}
	
	return &IPRange{
		StartIP: startIP,
		EndIP:   endIP,
	}, nil
}

// GenerateIPs generates all IPs in the range.
// It returns an empty slice for IPv6 addresses or invalid IP ranges.
func (r *IPRange) GenerateIPs() []net.IP {
	// Convert to 4-byte representation for easier arithmetic
	startIP := r.StartIP.To4()
	endIP := r.EndIP.To4()
	
	if startIP == nil || endIP == nil {
		// Handle IPv6 or invalid IPs
		return nil
	}
	
	// Convert to uint32 for easier arithmetic
	start := ipToUint32(startIP)
	end := ipToUint32(endIP)
	
	// Pre-allocate slice with known capacity for better performance
	capacity := int(end - start + 1)
	ips := make([]net.IP, 0, capacity)
	
	for i := start; i <= end; i++ {
		ips = append(ips, uint32ToIP(i))
	}
	
	return ips
}

// ipToUint32 converts IPv4 to uint32
func ipToUint32(ip net.IP) uint32 {
	ip = ip.To4()
	return uint32(ip[0])<<24 + uint32(ip[1])<<16 + uint32(ip[2])<<8 + uint32(ip[3])
}

// uint32ToIP converts uint32 to IPv4
func uint32ToIP(n uint32) net.IP {
	return net.IPv4(byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}

// GetLocalNetworkRange attempts to detect the local network range.
// It returns the first non-loopback IPv4 network found in CIDR format.
func GetLocalNetworkRange() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("failed to get interface addresses: %w", err)
	}
	
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				// Return the network in CIDR format
				return ipNet.String(), nil
			}
		}
	}
	
	return "", ErrNoLocalNetwork
}
