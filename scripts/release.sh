#!/bin/bash
# PixTui Release Script
# Usage: ./scripts/release.sh <version>
# Example: ./scripts/release.sh v1.0.0

set -e

VERSION=$1
if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v1.0.0"
    exit 1
fi

# Validate version format
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Version must be in format vX.Y.Z (e.g., v1.0.0)"
    exit 1
fi

echo "=== PixTui Release $VERSION ==="

# Ensure clean working tree
if [ -n "$(git status --porcelain)" ]; then
    echo "Error: Working tree is not clean. Commit or stash changes first."
    exit 1
fi

# Run tests
echo "Running tests..."
go test ./...
if [ $? -ne 0 ]; then
    echo "Error: Tests failed"
    exit 1
fi

# Run vet
echo "Running vet..."
go vet ./...
if [ $? -ne 0 ]; then
    echo "Error: Vet failed"
    exit 1
fi

# Build for all platforms
echo "Building binaries..."
mkdir -p dist
BUILD_DIR=dist

# Linux
GOOS=linux GOARCH=amd64 go build -o "$BUILD_DIR/pixtui-demo-linux-amd64" ./cmd/pixtui-demo/
GOOS=linux GOARCH=arm64 go build -o "$BUILD_DIR/pixtui-demo-linux-arm64" ./cmd/pixtui-demo/

# macOS
GOOS=darwin GOARCH=amd64 go build -o "$BUILD_DIR/pixtui-demo-darwin-amd64" ./cmd/pixtui-demo/
GOOS=darwin GOARCH=arm64 go build -o "$BUILD_DIR/pixtui-demo-darwin-arm64" ./cmd/pixtui-demo/

# Windows
GOOS=windows GOARCH=amd64 go build -o "$BUILD_DIR/pixtui-demo-windows-amd64.exe" ./cmd/pixtui-demo/

echo "Builds complete:"
ls -lh "$BUILD_DIR/"

# Create git tag
echo "Creating tag $VERSION..."
git tag -a "$VERSION" -m "Release $VERSION"
git push origin "$VERSION"

echo "=== Release $VERSION created successfully ==="
echo "GitHub Actions will build and publish the release automatically."
