package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/user/cc-server/internal/client"
)

var (
	serverAddr string
	regToken   string
)

func main() {
	flag.StringVar(&serverAddr, "server", "http://localhost:8080", "Server address to connect to")
	flag.StringVar(&regToken, "token", "", "Registration token")
	flag.Parse()

	if regToken == "" {
		log.Fatal("Registration token is required (-token flag)")
	}

	// Initialize client
	c, err := client.NewClient(serverAddr)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Register with server
	if err := c.Register(regToken); err != nil {
		log.Fatalf("Failed to register client: %v", err)
	}

	// Connect to server
	if err := c.Connect(); err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}

	// Wait for interrupt signal to gracefully shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Clean shutdown
	c.Close()
	log.Println("Client shutdown gracefully")
}