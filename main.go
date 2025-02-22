package portscanner

import (
	"net"
	"sync"
	"time"
)

type Portscanner struct {
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

func NewPortScanner(host string, start, end int) *Portscanner {
	return &Portscanner{
		Host:       host,
		StartPort:  start,
		EndPort:    end,
		Timeout:    100 * time.Millisecond,
		OpenPorts:  []int{},
		stopChan:   make(chan struct{}),
		totalPorts: end - start + 1,
	}
}

func (ps *Portscanner) ScanPort(port int, wg *sync.WaitGroup, results chan<- string, dialer DialerFunc) {
	defer wg.Done()

	select {
	case <-ps.stopChan:
		return
	default:
	}
	
}
