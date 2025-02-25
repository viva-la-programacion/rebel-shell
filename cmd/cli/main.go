package main

import (
	"fmt"
	"os"
	"strconv"

	"rebel-shell/internal/app"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "rebel-shell",
		Short: "A CLI tool for ethical blackhat operations",
	}

	// Add version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Display the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Rebelshell CLI version 0.1.0")
		},
	})

	// Add config command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "config [key] [value]",
		Short: "Set or retrieve configuration",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Simulate setting a config (adapt for your app logic)
			key, value := args[0], args[1]
			fmt.Printf("Configuration set: %s = %s\n", key, value)
		},
	})
	// TCP scanner command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "scan [network] [host] [startPort] [endPort] [concurrency] [throttle]",
		Short: "Scan range of ports on a host",
		Args:  cobra.ExactArgs(6),
		Run: func(cmd *cobra.Command, args []string) {
			network := args[0]
			host := args[1]
			startPort, _ := strconv.Atoi(args[2])
			endPort, _ := strconv.Atoi(args[3])
			concurrency, _ := strconv.Atoi(args[4])
			throttle := args[5] == "true"

			scanner := app.PortScanner{
				Host:        host,
				StartPort:   startPort,
				EndPort:     endPort,
				Concurrency: concurrency,
				Throttle:    throttle,
				Network:     network,
			}

			if err := scanner.Scan(); err != nil {
				fmt.Println("Error:", err)
				return
			}

			for _, result := range scanner.Results {
				fmt.Printf("Port %d: %s\n", result.Port, result.Status)
			}

		},
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
