#!/bin/bash

# run_tests.sh

set -e  # Exit immediately if a command exits with a non-zero status.

echo "Building the project..."
go build -v ./...

echo "Running all tests..."
go test -v ./...

echo "Running specific command tests..."
go test -v ./cmd -run TestDumpCommand
go test -v ./cmd -run TestReconstructCommand

echo "Running utils tests..."
go test -v ./utils -run TestMatchesPatterns

echo "All tests completed."