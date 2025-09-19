#!/bin/bash

# Performance measurement script for Whisp
# Measures startup time, memory usage, and other performance metrics

echo "=== Whisp Performance Analysis ==="
echo

# Build the application
echo "Building application..."
BUILD_START=$(date +%s.%3N)
go build -o /tmp/whisp-perf ./cmd/whisp
BUILD_END=$(date +%s.%3N)
BUILD_TIME=$(echo "$BUILD_END - $BUILD_START" | bc)

echo "Build time: ${BUILD_TIME}s"

# Check binary size
BINARY_SIZE=$(stat -c%s /tmp/whisp-perf)
BINARY_SIZE_MB=$(echo "scale=2; $BINARY_SIZE / 1048576" | bc)

echo "Binary size: ${BINARY_SIZE_MB}MB"

# Test startup time (headless mode if possible)
echo
echo "Testing startup performance..."

echo "Running initialization performance test..."
go run scripts/perf_benchmark.go

echo
echo "=== Performance Summary ==="
echo "Build Time: ${BUILD_TIME}s (Target: Fast builds)"
echo "Binary Size: ${BINARY_SIZE_MB}MB (Target: <50MB)"
echo "Note: Full startup time requires GUI testing in proper environment"
