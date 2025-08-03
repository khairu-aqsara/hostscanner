package scanner_test

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"hostscanner/scanner"
)

func TestScanNetwork(t *testing.T) {
	// Test with localhost
	ips := []net.IP{
		net.ParseIP("127.0.0.1"),
	}
	
	timeout := 500 * time.Millisecond
	maxWorkers := 10

	result := scanner.ScanNetwork(ips, timeout, maxWorkers)

	assert.NotNil(t, result)
	assert.Equal(t, 1, result.TotalHosts)
	assert.Equal(t, 1, len(result.Hosts))
	assert.True(t, result.Hosts[0].IsAlive, "Localhost should be alive")
	assert.Equal(t, "127.0.0.1", result.Hosts[0].IP.String())
}
