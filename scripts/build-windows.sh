#!/bin/bash
# Windows build script for Whisp

set -e

echo "🪟 Building Whisp for Windows..."

# Check if we're on Windows or using cross-compilation
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" ]]; then
    echo "Building on Windows"
    NATIVE=true
else
    echo "Cross-compiling for Windows"
    NATIVE=false
fi

# Build configuration
APP_NAME="whisp"
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}

# Build flags
LDFLAGS="-ldflags -X main.version=$VERSION -X main.buildTime=$BUILD_TIME -X main.gitCommit=$GIT_COMMIT -H windowsgui"

# Create build directory
mkdir -p build/windows

# Build application
echo "Building executable..."
GOOS=windows GOARCH=amd64 go build $LDFLAGS -o "build/windows/${APP_NAME}.exe" ./cmd/whisp

# Copy resources
echo "Copying resources..."
mkdir -p build/windows/resources
if [ -d "resources" ]; then
    cp -r resources/* build/windows/resources/
fi

# Create installer (if on Windows with tools available)
if [ "$NATIVE" = true ] && command -v makensis &> /dev/null; then
    echo "Creating NSIS installer..."
    # Copy NSIS script to build directory
    cp "scripts/whisp.nsi" "build/windows/"
    cd build/windows

    # Create installer
    makensis whisp.nsi

    if [ -f "whisp-windows-installer.exe" ]; then
        echo "✅ NSIS installer created: whisp-windows-installer.exe"
        mv "whisp-windows-installer.exe" "../whisp-windows-installer-${VERSION}.exe"
    else
        echo "⚠️  NSIS installer creation failed"
    fi

    cd ../..
else
    echo "⚠️  NSIS not available, skipping installer creation"
    echo "   To create installer: Install NSIS and run: makensis scripts/whisp.nsi"
fi

# Create zip package
echo "Creating zip package..."
cd build/windows
zip -r "../whisp-windows-amd64-${VERSION}.zip" ./*
cd ../..

echo "✅ Windows build complete!"
echo "   Executable: build/windows/${APP_NAME}.exe"
echo "   Package: build/whisp-windows-amd64-${VERSION}.zip"
