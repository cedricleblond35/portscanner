package portscanner

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type PortScanner struct {
	Host       string
	StartPort  int
	EndPort    int
	Timeout    time.Duration
	OpenPorts  []int
	mu         sync.Mutex
	stopChan   chan struct{}
	running    bool
	totalPorts int
	scanned    int
}

type DialerFunc func(network, address string, timeout time.Duration) (net.Conn, error)

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

func (ps *PortScanner) ScanPort(port int, wg *sync.WaitGroup, results chan<- string, dialer DialerFunc) {
	defer wg.Done()

	select {
	case <-ps.stopChan:
		return
	default:
	}
	address := fmt.Sprintf("%s:%d", ps.Host, port)
	conn, err := dialer("tcp", address, ps.Timeout)
	if err == nil {
		conn.Close()
		ps.mu.Lock()
		ps.OpenPorts = append(ps.OpenPorts, port)
		ps.mu.Unlock()
		results <- fmt.Sprintf("Port %d : OUVERT", port)
	} else {
		results <- fmt.Sprintf("Port %d : fermé ou filtré", port)
	}

	ps.mu.Lock()
	ps.scanned++
	ps.mu.Unlock()
}

// Scan lance le scan sur toute la plage de ports
func (ps *PortScanner) Scan() []string {
	ps.running = true
	var wg sync.WaitGroup
	results := make(chan string, ps.EndPort-ps.StartPort+1)
	var output []string

	for port := ps.StartPort; port <= ps.EndPort; port++ {
		wg.Add(1)
		go ps.ScanPort(port, &wg, results, net.DialTimeout)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		output = append(output, result)
	}

	ps.running = false
	return output
}

// Stop arrête le scan
func (ps *PortScanner) Stop() {
	if ps.running {
		close(ps.stopChan)
		ps.running = false
		ps.stopChan = make(chan struct{})
	}
}
