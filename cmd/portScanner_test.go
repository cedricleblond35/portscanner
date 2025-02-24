package cmd

import (
	"fmt"
	"net"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewPortScanner(t *testing.T) {
	host := "localhost"
	startPort := 1
	endPort := 100
	scanner := NewPortScanner(host, startPort, endPort)

	if scanner.Host != host {
		t.Errorf("Expected Host to be %s, got %s", host, scanner.Host)
	}

	if scanner.StartPort != startPort {
		t.Errorf("Expected StartPort to be %d, got %d", startPort, scanner.StartPort)
	}
	if scanner.Timeout != 2*time.Second {
		t.Errorf("Expected Timeout to be 2s, got %v", scanner.Timeout)
	}
	if len(scanner.OpenPorts) != 0 {
		t.Errorf("Expected OpenPorts to be empty, got %v", scanner.OpenPorts)
	}
	if scanner.totalPorts != endPort-startPort+1 {
		t.Errorf("Expected totalPorts to be %d, got %d", endPort-startPort+1, scanner.totalPorts)
	}
	if scanner.running {
		t.Error("Expected running to be false")
	}
}

// TestScanPort_Open tests ScanPort with an open port
func TestScanPort_Open(t *testing.T) {
	scanner := NewPortScanner("localhost", 8080, 8080)
	wg := sync.WaitGroup{}
	results := make(chan string, 1)

	dialer := func(network, address string, timeout time.Duration) (net.Conn, error) {
		return &net.TCPConn{}, nil
	}

	wg.Add(1)
	go scanner.ScanPort(8080, &wg, results, dialer)
	wg.Wait()
	close(results)

	result := <-results
	if !strings.Contains(result, "OPEN") {
		t.Errorf("Expected 'OPEN' in result, got %s", result)
	}
	if len(scanner.OpenPorts) != 1 || scanner.OpenPorts[0] != 8080 {
		t.Errorf("Expected OpenPorts to be [8080], got %v", scanner.OpenPorts)
	}
	if scanner.scanned != 1 {
		t.Errorf("Expected scanned to be 1, got %d", scanner.scanned)
	}
}

// TestScanPort_Closed tests ScanPort with a closed port
func TestScanPort_Closed(t *testing.T) {
	scanner := NewPortScanner("localhost", 8080, 8080)
	wg := sync.WaitGroup{}
	results := make(chan string, 1)

	dialer := func(network, address string, timeout time.Duration) (net.Conn, error) {
		return nil, fmt.Errorf("connection refused")
	}

	wg.Add(1)
	go scanner.ScanPort(8080, &wg, results, dialer)
	wg.Wait()
	close(results)

	result := <-results
	if !strings.Contains(result, "closed or filtered") {
		t.Errorf("Expected 'closed or filtered' in result, got %s", result)
	}
	if len(scanner.OpenPorts) != 0 {
		t.Errorf("Expected OpenPorts to be empty, got %v", scanner.OpenPorts)
	}
	if scanner.scanned != 1 {
		t.Errorf("Expected scanned to be 1, got %d", scanner.scanned)
	}
}

// TestScan tests the full Scan method with a small range
func TestScan(t *testing.T) {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	scanner := NewPortScanner("localhost", 8079, 8081)
	results := scanner.Scan()

	// Check results
	expected := []string{
		"Port 8079: closed or filtered",
		"Port 8080: OPEN",
		"Port 8081: closed or filtered",
	}
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
	for _, exp := range expected {
		found := false
		for _, res := range results {
			if res == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected result %s not found in %v", exp, results)
		}
	}
	if !reflect.DeepEqual(scanner.OpenPorts, []int{8080}) {
		t.Errorf("Expected OpenPorts to be [8080], got %v", scanner.OpenPorts)
	}
	if scanner.scanned != 3 {
		t.Errorf("Expected scanned to be 3, got %d", scanner.scanned)
	}
}

// TestStop tests the Stop method
func TestStop(t *testing.T) {
	scanner := NewPortScanner("localhost", 1, 10)
	scanner.running = true	// Simulate a running scan

	scanner.Stop()

	if scanner.running {
		t.Error("Expected running to be false after Stop")
	}
	select {
	case <-scanner.stopChan:
		// Channel should be closed
	default:
		t.Error("Expected stopChan to be closed after Stop")
	}

	// Verify reset for future scans
	scanner.running = true
	scanner.Stop()
	if scanner.running {
		t.Error("Expected running to be false after second Stop")
	}
	if scanner.stopChan == nil {
		t.Error("Expected stopChan to be reinitialized")
	}
}
