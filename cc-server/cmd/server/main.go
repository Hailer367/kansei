package main

import (
	"flag"
	"log"
	"os"

	"github.com/user/cc-server/internal/server"
)

var (
	addr         string
	supabaseURL  string
	supabaseKey  string
)

func main() {
	// Use the provided Supabase credentials
	supabaseURL = "https://ektnvfvzjdoyjwunpykh.supabase.co"
	supabaseKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImVrdG52ZnZ6amRveWp3dW5weWtoIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjI2MTk3MzYsImV4cCI6MjA3ODE5NTczNn0.deNHeJoTZuCytumXVkhRTK0sKlBHMH5jE0kYAWw1bnU"
	
	flag.StringVar(&addr, "addr", ":8080", "Server address")
	flag.Parse()

	log.Printf("Starting C&C server on %s", addr)
	log.Printf("Using Supabase URL: %s", supabaseURL)

	s := server.NewServer(addr, supabaseURL, supabaseKey)
	
	if err := s.Start(); err != nil {
		log.Printf("Server error: %v", err)
		os.Exit(1)
	}
}