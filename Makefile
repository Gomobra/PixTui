# PixTui Build System
# Mirrors the build system from the original Gomobra/PixTui repository

.PHONY: all build test vet fmt clean demo install run release

GO ?= go
GOFLAGS ?= -mod=mod
BUILD_DIR ?= build
BINARY ?= pixtui-demo

all: build

build:
	$(GO) build $(GOFLAGS) ./...

demo:
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/pixtui-demo/

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
	$(GO) install $(GOFLAGS) ./cmd/pixtui-demo/

# Cross-platform builds
build-linux:
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/pixtui-demo-linux-amd64 ./cmd/pixtui-demo/

build-darwin:
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/pixtui-demo-darwin-amd64 ./cmd/pixtui-demo/

build-windows:
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/pixtui-demo-windows-amd64.exe ./cmd/pixtui-demo/

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
