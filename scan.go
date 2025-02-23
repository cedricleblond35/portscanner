package main

import (
	"fmt"
	"os"
	"strconv"

	scanner "portscanner/cmd"
)

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
	scanner := scanner.NewPortScanner(host, startPort, endPort)
	fmt.Printf("Scanning %s from port %d to %d...\n", host, startPort, endPort)
	results := scanner.Scan()

	for _, result := range results {
		fmt.Println(result)
	}
	fmt.Printf("Open ports found: %v\n", scanner.OpenPorts)
}
