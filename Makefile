#!/usr/bin/env make

# Whisp - Cross-platform Tox Messenger
# Build system for all supported platforms

.PHONY: all clean build test coverage lint install deps
.PHONY: build-windows build-macos build-linux build-android build-ios build-all
.PHONY: run run-debug dev test-integration package-all

# Variables
APP_NAME := whisp
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null | sed 's/[^0-9.]*\([0-9.]*\).*/\1/' || echo "1.0.0")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"
BUILD_DIR := build
DIST_DIR := dist

# Platform detection
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# Default target
all: build

# Dependencies
deps:
	@echo "📦 Installing dependencies..."
	go mod download
	go mod verify

# Build for current platform
build: deps
	@echo "🔨 Building $(APP_NAME) for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./cmd/whisp

# Run application
run: build
	@echo "🚀 Running $(APP_NAME)..."
	./$(BUILD_DIR)/$(APP_NAME)

# Run with debug logging
run-debug: build
	@echo "🐛 Running $(APP_NAME) with debug logging..."
	./$(BUILD_DIR)/$(APP_NAME) -debug

# Development mode with hot reload (desktop only)
dev:
	@echo "🔥 Starting development mode..."
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

# Testing
test:
	@echo "🧪 Running tests..."
	go test -v ./...

test-coverage:
	@echo "📊 Running tests with coverage..."
	@mkdir -p $(BUILD_DIR)
	go test -v -coverprofile=$(BUILD_DIR)/coverage.out ./...
	go tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "Coverage report generated at $(BUILD_DIR)/coverage.html"

test-integration:
	@echo "🔬 Running integration tests..."
	go test -v -tags=integration ./...

# Linting
lint:
	@echo "🔍 Running linters..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin)
	golangci-lint run

# Platform-specific builds

# Windows build
build-windows:
	@echo "🪟 Building for Windows..."
	@mkdir -p $(BUILD_DIR)/windows
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/windows/$(APP_NAME).exe ./cmd/whisp
	@echo "✅ Windows build complete: $(BUILD_DIR)/windows/$(APP_NAME).exe"

# macOS build
build-macos:
	@echo "🍎 Building for macOS..."
	@mkdir -p $(BUILD_DIR)/macos
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/macos/$(APP_NAME)-amd64 ./cmd/whisp
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/macos/$(APP_NAME)-arm64 ./cmd/whisp
	@echo "✅ macOS builds complete"

# Linux build
build-linux:
	@echo "🐧 Building for Linux..."
	@mkdir -p $(BUILD_DIR)/linux
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/linux/$(APP_NAME)-amd64 ./cmd/whisp
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/linux/$(APP_NAME)-arm64 ./cmd/whisp
	@echo "✅ Linux builds complete"

# Android build
build-android:
	@echo "🤖 Building for Android..."
	@which fyne > /dev/null || (echo "Installing fyne CLI..." && go install fyne.io/fyne/v2/cmd/fyne@latest)
	@mkdir -p $(BUILD_DIR)/android
	go build $(LDFLAGS) -o $(BUILD_DIR)/android/$(APP_NAME) ./cmd/whisp
	cp assets/icons/icon-192.png $(BUILD_DIR)/android/Icon.png
	cd $(BUILD_DIR)/android && fyne package -os android -appBuild 1 -appVersion $(VERSION) -appID io.whisp.app -icon Icon.png
	@echo "✅ Android build complete: $(BUILD_DIR)/android/$(APP_NAME).apk"

# iOS build (requires macOS)
build-ios:
	@echo "📱 Building for iOS..."
ifeq ($(GOOS),darwin)
	@which fyne > /dev/null || (echo "Installing fyne CLI..." && go install fyne.io/fyne/v2/cmd/fyne@latest)
	@mkdir -p $(BUILD_DIR)/ios
	go build $(LDFLAGS) -o $(BUILD_DIR)/ios/$(APP_NAME) ./cmd/whisp
	cp assets/icons/icon-192.png $(BUILD_DIR)/ios/Icon.png
	cd $(BUILD_DIR)/ios && fyne package -os ios -appBuild 1 -appVersion $(VERSION) -appID io.whisp.app -icon Icon.png
	@echo "✅ iOS build complete: $(BUILD_DIR)/ios/$(APP_NAME).ipa"
else
	@echo "❌ iOS builds require macOS"
	@exit 1
endif

# Build all platforms
build-all: build-windows build-macos build-linux build-android
ifeq ($(GOOS),darwin)
	@$(MAKE) build-ios
endif
	@echo "🎉 All platform builds complete!"

# Packaging

package-windows: build-windows
	@echo "📦 Creating Windows installer..."
	@mkdir -p $(DIST_DIR)/windows
	@echo "TODO: Create MSI/MSIX installer"

package-macos: build-macos
	@echo "📦 Creating macOS package..."
	@mkdir -p $(DIST_DIR)/macos
	@echo "TODO: Create DMG installer"

package-linux: build-linux
	@echo "📦 Creating Linux packages..."
	@mkdir -p $(DIST_DIR)/linux
	@echo "TODO: Create AppImage, Flatpak, Snap, deb, rpm"

package-all: package-windows package-macos package-linux
	@echo "🎁 All packages created!"

# Installation
install: build
	@echo "📦 Installing $(APP_NAME)..."
	@mkdir -p ~/.local/bin
	cp $(BUILD_DIR)/$(APP_NAME) ~/.local/bin/
	@echo "✅ $(APP_NAME) installed to ~/.local/bin/"

# Cleanup
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	go clean

# Help
help:
	@echo "🔧 Whisp Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  build         - Build for current platform"
	@echo "  run           - Build and run application"
	@echo "  run-debug     - Run with debug logging"
	@echo "  dev           - Development mode with hot reload"
	@echo "  test          - Run unit tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run code linters"
	@echo "  deps          - Install dependencies"
	@echo ""
	@echo "Platform builds:"
	@echo "  build-windows - Build for Windows"
	@echo "  build-macos   - Build for macOS (Intel + Apple Silicon)"
	@echo "  build-linux   - Build for Linux (x64 + ARM64)"
	@echo "  build-android - Build for Android"
	@echo "  build-ios     - Build for iOS (macOS only)"
	@echo "  build-all     - Build for all platforms"
	@echo ""
	@echo "Packaging:"
	@echo "  package-all   - Create installers for all platforms"
	@echo ""
	@echo "Utilities:"
	@echo "  install       - Install to ~/.local/bin"
	@echo "  clean         - Clean build artifacts"
	@echo "  help          - Show this help"
