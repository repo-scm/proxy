#!/bin/bash

# Test script to run the proxy server with test data
echo "Building proxy..."
go build -o bin/proxy .

echo "Starting proxy server with test data..."
echo "The server will run on http://localhost:9090"
echo "Access the web UI at http://localhost:9090/ui"
echo "Press Ctrl+C to stop the server"
echo ""

./bin/proxy serve --test --config config/test.yaml --address :9090
