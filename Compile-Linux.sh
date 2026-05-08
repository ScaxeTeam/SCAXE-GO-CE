#!/bin/bash

echo "Go version:"
go version

echo -e "\n\nDownloading dependencies..."
go mod tidy

echo "Compiling for Linux..."
go build -o scaxe-server ./cmd/server

if [ $? -eq 0 ]; then
    echo "Done. Build successful!"
else
    echo "Error: Compilation failed."
    exit 1
fi
