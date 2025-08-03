package network_test

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"hostscanner/network"
)

func TestParseIPRange_CIDR(t *testing.T) {
	// Test CIDR notation
	ipRange, err := network.ParseIPRange("192.168.1.0/24")
	assert.NoError(t, err)
	assert.NotNil(t, ipRange)
	
	ips := ipRange.GenerateIPs()
	assert.Equal(t, 256, len(ips))
	assert.Equal(t, "192.168.1.0", ips[0].String())
	assert.Equal(t, "192.168.1.255", ips[255].String())
}

func TestParseIPRange_Range(t *testing.T) {
	// Test range notation
	ipRange, err := network.ParseIPRange("192.168.1.1-192.168.1.5")
	assert.NoError(t, err)
	assert.NotNil(t, ipRange)
	
	ips := ipRange.GenerateIPs()
	assert.Equal(t, 5, len(ips))
	assert.Equal(t, "192.168.1.1", ips[0].String())
	assert.Equal(t, "192.168.1.5", ips[4].String())
}

func TestParseIPRange_SingleIP(t *testing.T) {
	// Test single IP
	ipRange, err := network.ParseIPRange("192.168.1.1")
	assert.NoError(t, err)
	assert.NotNil(t, ipRange)
	
	ips := ipRange.GenerateIPs()
	assert.Equal(t, 1, len(ips))
	assert.Equal(t, "192.168.1.1", ips[0].String())
}

func TestParseIPRange_InvalidIP(t *testing.T) {
	// Test invalid IP
	_, err := network.ParseIPRange("invalid.ip.address")
	assert.Error(t, err)
}

func TestParseIPRange_InvalidCIDR(t *testing.T) {
	// Test invalid CIDR
	_, err := network.ParseIPRange("192.168.1.0/40")
	assert.Error(t, err)
}

func TestGetLocalNetworkRange(t *testing.T) {
	// Test getting local network range
	localNetwork, err := network.GetLocalNetworkRange()
	
	// This might fail in some environments (like CI), so we'll check if we get a result or a reasonable error
	if err != nil {
		t.Logf("Could not detect local network: %v", err)
		return
	}
	
	assert.NotEmpty(t, localNetwork)
	
	// Verify it's a valid CIDR
	_, _, err = net.ParseCIDR(localNetwork)
	assert.NoError(t, err)
}
