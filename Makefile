# OpenTUI Go Build System
# Mirrors the build system from the original anomalyco/opentui repository

.PHONY: all build test vet fmt clean demo install run release

GO ?= go
GOFLAGS ?= -mod=mod
BUILD_DIR ?= build
BINARY ?= opentui-demo

all: build

build:
	$(GO) build $(GOFLAGS) ./...

demo:
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/opentui-demo/

run: demo
	./$(BUILD_DIR)/$(BINARY)

test:
	$(GO) test $(GOFLAGS) ./...

vet:
	$(GO) vet $(GOFLAGS) ./...

fmt:
	$(GO) fmt ./...

fmt-check:
	@$(GO) fmt -l ./... | grep -v "^$$" || true

clean:
	rm -rf $(BUILD_DIR)
	$(GO) clean -cache

install: build
	$(GO) install $(GOFLAGS) ./cmd/opentui-demo/

# Cross-platform builds
build-linux:
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/opentui-demo-linux-amd64 ./cmd/opentui-demo/

build-darwin:
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/opentui-demo-darwin-amd64 ./cmd/opentui-demo/

build-windows:
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/opentui-demo-windows-amd64.exe ./cmd/opentui-demo/

# Module management
tidy:
	$(GO) mod tidy

download:
	$(GO) mod download

# Verify module integrity
verify:
	$(GO) mod verify

# Default target
all: build test vet
