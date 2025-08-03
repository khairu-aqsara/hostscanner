package scanner

import (
	"errors"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Common errors returned by this package.
var (
	ErrUnsupportedOS = errors.New("unsupported operating system")
)

// Host represents a discovered host on the network.
type Host struct {
	IP       net.IP        `json:"ip"`
	Hostname string        `json:"hostname"`
	MAC      string        `json:"mac"`
	Vendor   string        `json:"vendor"`
	Latency  time.Duration `json:"latency"`
	IsAlive  bool          `json:"is_alive"`
	Error    error         `json:"error,omitempty"`
}

// ScanResult represents the complete network scan results.
type ScanResult struct {
	NetworkRange string        `json:"network_range"`
	TotalHosts   int           `json:"total_hosts"`
	AliveHosts   int           `json:"alive_hosts"`
	Hosts        []Host        `json:"hosts"`
	ScanTime     time.Duration `json:"scan_time"`
}

// ScanNetwork scans a network range for active hosts.
// It uses a worker pool pattern for concurrent scanning.
func ScanNetwork(ips []net.IP, timeout time.Duration, maxWorkers int) *ScanResult {
	start := time.Now()
	result := &ScanResult{
		TotalHosts: len(ips),
		Hosts:      make([]Host, 0, len(ips)),
	}

	// Create worker pool
	jobs := make(chan net.IP, len(ips))
	results := make(chan Host, len(ips))

	// Start workers
	var wg sync.WaitGroup
	for w := 0; w < maxWorkers; w++ {
		wg.Add(1)
		go worker(jobs, results, timeout, &wg)
	}

	// Send jobs
	go func() {
		for _, ip := range ips {
			jobs <- ip
		}
		close(jobs)
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Process results
	for host := range results {
		result.Hosts = append(result.Hosts, host)
		if host.IsAlive {
			result.AliveHosts++
		}
	}

	result.ScanTime = time.Since(start)
	return result
}

// worker performs host discovery for each IP.
func worker(jobs <-chan net.IP, results chan<- Host, timeout time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()
	
	for ip := range jobs {
		host := scanHost(ip, timeout)
		results <- host
	}
}

// scanHost checks if a host is alive and gathers information.
func scanHost(ip net.IP, timeout time.Duration) Host {
	host := Host{
		IP:      ip,
		IsAlive: false,
	}

	// Ping the host
	start := time.Now()
	isAlive, err := pingHost(ip.String(), timeout)
	host.Latency = time.Since(start)
	host.IsAlive = isAlive
	host.Error = err

	if isAlive {
		// Try to resolve hostname
		if names, err := net.LookupAddr(ip.String()); err == nil && len(names) > 0 {
			host.Hostname = strings.TrimSuffix(names[0], ".")
		}

		// Try to get MAC address (works better on local network)
		if mac := getMACAddress(ip.String()); mac != "" {
			host.MAC = mac
			host.Vendor = getVendorFromMAC(mac)
		}
	}

	return host
}

// pingHost pings a host to check if it's alive.
func pingHost(ip string, timeout time.Duration) (bool, error) {
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("ping", "-n", "1", "-w", fmt.Sprintf("%.0f", timeout.Seconds()*1000), ip)
	case "darwin", "linux":
		cmd = exec.Command("ping", "-c", "1", "-W", fmt.Sprintf("%.0f", timeout.Seconds()*1000), ip)
	default:
		return false, fmt.Errorf("%w: %s", ErrUnsupportedOS, runtime.GOOS)
	}

	err := cmd.Run()
	return err == nil, err
}

// getMACAddress attempts to get MAC address using ARP table.
// It returns an empty string if the MAC address cannot be determined.
func getMACAddress(ip string) string {
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("arp", "-a", ip)
	case "darwin", "linux":
		cmd = exec.Command("arp", "-n", ip)
	default:
		return ""
	}

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Parse ARP output to extract MAC address
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, ip) {
			fields := strings.Fields(line)
			for _, field := range fields {
				if strings.Count(field, ":") == 5 || strings.Count(field, "-") == 5 {
					return strings.ToUpper(field)
				}
			}
		}
	}

	return ""
}

// getVendorFromMAC returns vendor information based on MAC address OUI.
// It returns "Unknown" if the vendor cannot be determined.
func getVendorFromMAC(mac string) string {
	if len(mac) < 8 {
		return "Unknown"
	}

	// Extract OUI (first 3 octets)
	oui := strings.ReplaceAll(mac[:8], ":", "")
	oui = strings.ReplaceAll(oui, "-", "")
	oui = strings.ToUpper(oui)

	// Common vendor mappings based on OUI database
	vendors := map[string]string{
		"00:50:56": "VMware",
		"08:00:27": "Oracle VirtualBox",
		"52:54:00": "QEMU/KVM",
		"B8:27:EB": "Raspberry Pi Foundation",
		"DC:A6:32": "Raspberry Pi Foundation",
		"E4:5F:01": "Raspberry Pi Foundation",
		"00:16:3E": "Xen",
		"00:1C:42": "Parallels",
		"AC:DE:48": "Apple",
		"F8:FF:C2": "Apple",
		"28:CD:C1": "Apple",
		"3C:07:54": "Apple",
	}

	for ouiPrefix, vendor := range vendors {
		if strings.HasPrefix(mac, ouiPrefix) {
			return vendor
		}
	}

	return "Unknown"
}
