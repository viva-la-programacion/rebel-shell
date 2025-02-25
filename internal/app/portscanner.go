package app

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"sort"
	"sync"
	"time"
)

type PortScanner struct {
	Network     string
	Host        string
	StartPort   int
	EndPort     int
	Concurrency int
	Throttle    bool
	Results     []PortStatus
	ElapsedTime time.Duration

	CheckPort func(network, host string, port int) (string, error)
}

type PortStatus struct {
	Port   int    `json:"port"`
	Status string `json:"status"`
	Error  string `json:"error"`
}

const (
	OPEN    = "open"
	CLOSED  = "closed"
	TIMEOUT = "timeout"
)

var (
	ErrInvalidHost = errors.New(
		"invalid host",
	)
	ErrInvalidPortRange = errors.New(
		"invalid port range, valid range (1-65535)",
	)
	ErrStartPortGreater = errors.New(
		"start port cannot be greater than end port",
	)
)

// Validate checks the validity of scanner parameters.
func (ps *PortScanner) Validate() error {
	if ps.CheckPort == nil {
		ps.CheckPort = ps.checkPort
	}
	if !validateHost(ps.Host) {
		return ErrInvalidHost
	}
	if !validatePort(ps.StartPort) || !validatePort(ps.EndPort) {
		return ErrInvalidPortRange
	}
	if ps.StartPort > ps.EndPort {
		return ErrStartPortGreater
	}
	if ps.Network == "" {
		ps.Network = "tcp"
	}
	return nil
}

// Scan performs the port scan and stores results.
func (ps *PortScanner) Scan() error {
	if err := ps.Validate(); err != nil {
		return err
	}

	startTime := time.Now()
	portCh := make(chan int, ps.Concurrency)
	resultCh := make(chan PortStatus, ps.EndPort-ps.StartPort+1)
	var wg sync.WaitGroup
	var errorCount int

	// Worker goroutines for port scanning.
	for i := 0; i < ps.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for port := range portCh {
				status, portErr := ps.CheckPort(ps.Network, ps.Host, port)
				if portErr != nil {
					errorCount++
				}
				resultCh <- PortStatus{
					Port:   port,
					Status: status,
					Error: func() string {
						if portErr != nil {
							return portErr.Error()
						} else {
							return ""
						}
					}(),
				}

				if ps.Throttle {
					// Randomized delay with adjustment based on error count
					delay := time.Duration(
						rand.Intn(100)+50,
					) * time.Millisecond
					if errorCount > 5 {
						delay += time.Duration(
							errorCount*10,
						) * time.Millisecond
					}
					time.Sleep(delay)
				}
			}
		}()
	}

	// Send ports to the portCh channel.
	go func() {
		for port := ps.StartPort; port <= ps.EndPort; port++ {
			portCh <- port
		}
		close(portCh)
	}()

	// Close the result channel once all workers are done.
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results from the result channel.
	for result := range resultCh {
		ps.Results = append(ps.Results, result)
	}
	sort.Slice(ps.Results, func(i, j int) bool {
		return ps.Results[i].Port < ps.Results[j].Port
	})

	elapsedTime := time.Since(startTime)
	fmt.Printf("Scan completed in %s\n", elapsedTime)
	ps.ElapsedTime = elapsedTime

	return nil
}

// checkPort attempts to connect to a port and determines its status.
func (ps *PortScanner) checkPort(
	network string,
	host string,
	port int,
) (string, error) {
	conn, err := net.Dial(network, fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return TIMEOUT, err
		}
		return CLOSED, err
	}
	conn.Close()
	return OPEN, nil
}

// Helper functions
func validatePort(port int) bool {
	return port > 0 && port < 65536
}

func validateHost(host string) bool {
	if net.ParseIP(host) != nil {
		return true
	}
	_, err := net.LookupHost(host)
	return err == nil
}
