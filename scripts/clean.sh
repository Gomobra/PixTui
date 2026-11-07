#!/bin/bash
# PixTui Clean Script
# Removes build artifacts and caches

set -e

echo "=== Cleaning PixTui ==="

rm -rf build/ dist/
go clean -cache -testcache

echo "=== Clean complete ==="
