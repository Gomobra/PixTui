#!/bin/bash
# PixTui Test Script
# Runs all tests with coverage

set -e

echo "=== Running PixTui Tests ==="

go test -v -cover ./...
echo "=== Tests complete ==="
