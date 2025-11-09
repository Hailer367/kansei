#!/bin/bash

echo "Building C&C Server components..."

# Build the server
echo "Building server..."
go build -o bin/cc-server cmd/server/main.go

# Build the client
echo "Building client..."
go build -o bin/cc-client cmd/client/main.go

# Build the CLI
echo "Building CLI..."
go build -o bin/cc-cli cmd/cli/main.go

echo "Build completed. Binaries are in the bin/ directory."
echo "To run the server: ./bin/cc-server"
echo "To run a client: ./bin/cc-client -token YOUR_TOKEN"
echo "To use the CLI: ./bin/cc-cli list-clients"