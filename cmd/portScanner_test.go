package cmd

import (
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
