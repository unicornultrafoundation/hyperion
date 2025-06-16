package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/0xsoniclabs/hyperion/driver"
	"github.com/0xsoniclabs/hyperion/driver/network/external"
	"github.com/0xsoniclabs/hyperion/driver/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: external_chain <rpc_endpoint> [chain_id]")
		fmt.Println("Example: external_chain http://localhost:18545 4002")
		os.Exit(1)
	}

	rpcEndpoint := os.Args[1]
	chainID := int64(4002) // Default to a common test chain ID

	if len(os.Args) > 2 {
		fmt.Sscanf(os.Args[2], "%d", &chainID)
	}

	log.Printf("Connecting to external chain at %s (Chain ID: %d)", rpcEndpoint, chainID)

	// Create external network configuration
	config := &external.ExternalNetworkConfig{
		NetworkConfig: driver.NetworkConfig{
			Validators:    []driver.Validator{}, // No validators needed for external
			RoundTripTime: 0,
			NetworkRules:  map[string]string{},
			OutputDir:     "/tmp/hyperion_external",
		},
		RpcEndpoints: []string{rpcEndpoint},
		ChainID:      chainID,
	}

	// Create external network
	network, err := external.NewExternalNetwork(config)
	if err != nil {
		log.Fatalf("Failed to create external network: %v", err)
	}
	defer network.Shutdown()

	// Create a simple counter application
	appConfig := &driver.ApplicationConfig{
		Name:  "counter-test",
		Type:  "counter",
		Users: 5,
		Rate: &parser.Rate{
			Constant: func() *float32 { f := float32(2); return &f }(), // 2 transactions per second
		},
	}

	log.Printf("Creating application: %s", appConfig.Name)
	app, err := network.CreateApplication(appConfig)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	log.Printf("Starting application...")
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	log.Printf("Application started successfully!")
	log.Printf("Sending transactions to your local chain...")
	log.Printf("Press Ctrl+C to stop")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Print periodic status updates
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			log.Printf("Received interrupt signal, shutting down...")
			app.Stop()
			return
		case <-ticker.C:
			// Print status update
			sent, err := app.GetReceivedTransactions()
			if err != nil {
				log.Printf("Error getting transaction count: %v", err)
			} else {
				log.Printf("Transactions processed: %d", sent)
			}
		}
	}
}
