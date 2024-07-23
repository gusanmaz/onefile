#!/bin/bash

# Ensure we're in the project root
cd "$(dirname "$0")"

# Build the main onefile command
go build -o bin/onefile main.go

# Build individual commands
go build -o bin/dump cmd/dump/main.go
go build -o bin/github2file cmd/github2file/main.go
go build -o bin/json2md cmd/json2md/main.go
go build -o bin/pypi2file cmd/pypi2file/main.go
go build -o bin/reconstruct cmd/reconstruct/main.go

echo "All commands have been built and placed in the bin directory."