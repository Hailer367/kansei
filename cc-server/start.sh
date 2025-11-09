#!/bin/bash

# Simple startup script for the C&C server

echo "Starting C&C Server..."

# Run the server in the background
./bin/cc-server &

SERVER_PID=$!

echo "C&C Server started with PID: $SERVER_PID"

# Function to stop the server
cleanup() {
    echo "Stopping C&C Server..."
    kill $SERVER_PID
    exit 0
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

# Wait for server to finish (it won't unless killed)
wait $SERVER_PID