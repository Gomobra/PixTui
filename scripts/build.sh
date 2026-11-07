#!/bin/bash
# PixTui Build Script
# Usage: ./scripts/build.sh [output-dir]

set -e

OUTPUT_DIR=${1:-build}
mkdir -p "$OUTPUT_DIR"

echo "=== Building PixTui ==="

# Build all packages
go build ./...
echo "All packages built successfully."

# Build demo binary
go build -o "$OUTPUT_DIR/pixtui-demo" ./cmd/pixtui-demo/
echo "Demo binary built: $OUTPUT_DIR/pixtui-demo"

echo "=== Build complete ==="
