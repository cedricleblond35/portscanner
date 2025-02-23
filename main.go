package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

// PortScanner holds the logic for scanning ports
type PortScanner struct {
	Host       string        // Target host (e.g., "localhost" or "192.168.1.1")
	StartPort  int           // Starting port number
	EndPort    int           // Ending port number
	Timeout    time.Duration // Timeout for each connection attempt
	OpenPorts  []int         // List of open ports found
	mu         sync.Mutex    // Mutex to protect OpenPorts from concurrent access
	stopChan   chan struct{} // Channel to signal stop
	running    bool          // Indicates if the scan is running
	totalPorts int           // Total number of ports to scan
	scanned    int           // Number of ports scanned
}

// DialerFunc is an abstraction for net.DialTimeout, useful for testing
type DialerFunc func(network, address string, timeout time.Duration) (net.Conn, error)

// NewPortScanner initializes a new port scanner
func NewPortScanner(host string, start, end int) *PortScanner {
	return &PortScanner{
		Host:       host,
		StartPort:  start,
		EndPort:    end,
		Timeout:    100 * time.Millisecond,
		OpenPorts:  []int{},
		stopChan:   make(chan struct{}),
		totalPorts: end - start + 1,
	}
}

// ScanPort checks if a specific port is open
func (ps *PortScanner) ScanPort(port int, wg *sync.WaitGroup, results chan<- string, dialer DialerFunc) {
	defer wg.Done()

	// Check if the scan should stop
	select {
	case <-ps.stopChan:
		return
	default:
	}

	// Construct the target address (e.g., "localhost:80")
	address := fmt.Sprintf("%s:%d", ps.Host, port)
	conn, err := dialer("tcp", address, ps.Timeout)
	if err == nil {
		conn.Close()
		ps.mu.Lock()
		ps.OpenPorts = append(ps.OpenPorts, port)
		ps.mu.Unlock()
		results <- fmt.Sprintf("Port %d : OPEN", port)
	} else {
		results <- fmt.Sprintf("Port %d : closed or filtered", port)
	}

	ps.mu.Lock()
	ps.scanned++
	ps.mu.Unlock()
}

// Scan runs the port scan over the specified range
func (ps *PortScanner) Scan() []string {
	ps.running = true
	var wg sync.WaitGroup
	results := make(chan string, ps.EndPort-ps.StartPort+1)
	var output []string

	// Launch a goroutine for each port
	for port := ps.StartPort; port <= ps.EndPort; port++ {
		wg.Add(1)
		go ps.ScanPort(port, &wg, results, net.DialTimeout)
	}

	// Close the results channel when all goroutines finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	for result := range results {
		output = append(output, result)
	}

	ps.running = false
	return output
}

// Stop terminates the scan
func (ps *PortScanner) Stop() {
	if ps.running {
		close(ps.stopChan)
		ps.running = false
		ps.stopChan = make(chan struct{}) // Reset for future scans
	}
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run portscanner.go <host> <start-port> <end-port>")
		fmt.Println("Example: go run portscanner.go localhost 1 1024")
	}

	host := os.Args[1]
	startPort, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Error: invalid start port")
		os.Exit(1)
	}
	endPort, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Println("Error: invalid end port")
		os.Exit(1)
	}

	if startPort < 1 || endPort > 65535 || startPort > endPort {
		println("Error: port range must be between 1 and 65535, with start <= end")
		os.Exit(1)
	}

	// Create and run the scanner
	scanner := NewPortScanner(host, startPort, endPort)
	fmt.Printf("Scanning %s from port %d to %d...\n", host, startPort, endPort)
	results := scanner.Scan()

	for _, result := range results {
		fmt.Println(result)
	}
	fmt.Printf("Open ports found: %v\n", scanner.OpenPorts)
}
