package app_test

import (
	"fmt"
	"rebel-shell/internal/app"
	"sync"
	"testing"
	"time"
)

func TestPortScannerValidate(t *testing.T) {
	tests := []struct {
		scanner  app.PortScanner
		expected error
	}{
		{app.PortScanner{Host: "127.0.0.1", StartPort: 1, EndPort: 65535}, nil},
		{app.PortScanner{Host: "invalid_host", StartPort: 1, EndPort: 65535}, app.ErrInvalidHost},
		{app.PortScanner{Host: "127.0.0.1", StartPort: 0, EndPort: 65535}, app.ErrInvalidPortRange},
		{app.PortScanner{Host: "127.0.0.1", StartPort: 1, EndPort: 0}, app.ErrInvalidPortRange},
		{app.PortScanner{Host: "127.0.0.1", StartPort: 65535, EndPort: 1}, app.ErrStartPortGreater},
	}

	for _, test := range tests {
		result := test.scanner.Validate()
		if result != test.expected {
			t.Errorf("Validate() = %v; want %v", result, test.expected)
		}
	}
}
func TestPortScannerThrottle(t *testing.T) {
	scanner := app.PortScanner{
		Host:        "127.0.0.1",
		StartPort:   1,
		EndPort:     5,
		Concurrency: 1,
		Throttle:    true,
		CheckPort: func(network, host string, port int) (string, error) {
			return app.CLOSED, nil
		},
	}

	start := time.Now()
	scanner.Scan()
	elapsed := time.Since(start)

	if elapsed < time.Duration(250*time.Millisecond) {
		t.Errorf("expected throttling to add delay; took %s", elapsed)
	}
}

func TestPortScannerScan(t *testing.T) {
	// Mock CheckPort function
	mockCheckPort := func(network, host string, port int) (string, error) {
		if port%2 == 0 {
			return app.OPEN, nil
		}
		return app.CLOSED, nil
	}

	// Initialize the scanner with the mock function
	scanner := app.PortScanner{
		Host:        "127.0.0.1",
		StartPort:   1,
		EndPort:     10,
		Concurrency: 2,
		Throttle:    false,
		CheckPort:   mockCheckPort, // Inject mock function
	}

	err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	// Debugging output
	t.Logf("Scan Results: %+v", scanner.Results)

	expectedOpenPorts := 5 // Even ports from 1 to 10
	actualOpenPorts := 0
	for _, port := range scanner.Results {
		if port.Status == app.OPEN {
			actualOpenPorts++
		}
	}

	if actualOpenPorts != expectedOpenPorts {
		t.Errorf("expected %d open ports; got %d", expectedOpenPorts, actualOpenPorts)
	}
}
func TestPortScannerConcurrency(t *testing.T) {
	scannedPorts := make(map[int]bool)
	var mu sync.Mutex

	// Mock CheckPort function
	mockCheckPort := func(network, host string, port int) (string, error) {
		mu.Lock()
		defer mu.Unlock()
		if scannedPorts[port] {
			t.Errorf("port %d scanned multiple times", port)
		}
		scannedPorts[port] = true
		return app.CLOSED, nil
	}

	scanner := app.PortScanner{
		Host:        "127.0.0.1",
		StartPort:   1,
		EndPort:     10,
		Concurrency: 5,
		Throttle:    false,
		CheckPort:   mockCheckPort,
	}

	err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	// Ensure all ports were scanned
	for port := 1; port <= 10; port++ {
		if !scannedPorts[port] {
			t.Errorf("port %d was not scanned", port)
		}
	}
}

func TestPortScannerSinglePort(t *testing.T) {
	mockCheckPort := func(network, host string, port int) (string, error) {
		if port == 80 {
			return app.OPEN, nil
		}
		return app.CLOSED, nil
	}

	scanner := app.PortScanner{
		Host:        "127.0.0.1",
		StartPort:   80,
		EndPort:     80,
		Concurrency: 1,
		Throttle:    false,
		CheckPort:   mockCheckPort,
	}

	err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	if len(scanner.Results) != 1 {
		t.Errorf("expected 1 result; got %d", len(scanner.Results))
	}

	if scanner.Results[0].Port != 80 || scanner.Results[0].Status != app.OPEN {
		t.Errorf("expected port 80 to be open; got %+v", scanner.Results[0])
	}
}

func TestPortScannerResultsConsistency(t *testing.T) {
	mockCheckPort := func(network, host string, port int) (string, error) {
		if port%3 == 0 {
			return "", fmt.Errorf("mock error for port %d", port)
		}
		return app.CLOSED, nil
	}

	scanner := app.PortScanner{
		Host:        "127.0.0.1",
		StartPort:   1,
		EndPort:     10,
		Concurrency: 2,
		Throttle:    false,
		CheckPort:   mockCheckPort,
	}

	err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	if len(scanner.Results) != 10 {
		t.Errorf("expected 10 results; got %d", len(scanner.Results))
	}

	// Ensure results are sorted by port number
	for i := 1; i < len(scanner.Results); i++ {
		if scanner.Results[i-1].Port > scanner.Results[i].Port {
			t.Errorf("Results not sorted: %+v", scanner.Results)
			break
		}
	}
}

func BenchmarkPortScanner(b *testing.B) {
	mockCheckPort := func(network, host string, port int) (string, error) {
		return app.CLOSED, nil
	}

	for i := 1; i <= 10; i++ {
		b.Run(fmt.Sprintf("Concurrency%d", i), func(b *testing.B) {
			scanner := app.PortScanner{
				Host:        "127.0.0.1",
				StartPort:   1,
				EndPort:     100,
				Concurrency: i,
				Throttle:    false,
				CheckPort:   mockCheckPort,
			}

			for n := 0; n < b.N; n++ {
				_ = scanner.Scan()
			}
		})
	}
}

func TestPortScannerLargeRange(t *testing.T) {
	mockCheckPort := func(network, host string, port int) (string, error) {
		return app.CLOSED, nil
	}

	scanner := app.PortScanner{
		Host:        "127.0.0.1",
		StartPort:   1,
		EndPort:     10000, // Large range
		Concurrency: 100,
		Throttle:    false,
		CheckPort:   mockCheckPort,
	}

	err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	if len(scanner.Results) != 10000 {
		t.Errorf("expected 10000 results; got %d", len(scanner.Results))
	}
}

func TestPortScannerUnreachableHost(t *testing.T) {
	scanner := app.PortScanner{
		Host:        "unreachable.host",
		StartPort:   1,
		EndPort:     5,
		Concurrency: 2,
		Throttle:    false,
		CheckPort: func(network, host string, port int) (string, error) {
			return app.CLOSED, fmt.Errorf("mock unreachable host error")
		},
	}

	err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	for _, result := range scanner.Results {
		if result.Status != app.CLOSED {
			t.Errorf("expected CLOSED for port %d; got %s", result.Port, result.Status)
		}
	}
}
