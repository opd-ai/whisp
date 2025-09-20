#!/bin/bash
# CI/CD Pipeline Validation Script
# This script validates the GitHub Actions workflow files and simulates their execution

set -e

echo "🔍 Validating GitHub Actions workflow files..."

# Check if workflow files exist
WORKFLOW_DIR=".github/workflows"
if [ ! -d "$WORKFLOW_DIR" ]; then
    echo "❌ Error: $WORKFLOW_DIR directory not found"
    exit 1
fi

# Find all workflow files
WORKFLOW_FILES=$(find $WORKFLOW_DIR -name "*.yml" -o -name "*.yaml")

if [ -z "$WORKFLOW_FILES" ]; then
    echo "❌ Error: No workflow files found in $WORKFLOW_DIR"
    exit 1
fi

echo "📋 Found workflow files:"
for file in $WORKFLOW_FILES; do
    echo "  - $file"
done

echo ""
echo "🧪 Simulating CI/CD pipeline steps..."

# Simulate dependency installation (key step that could fail)
echo "📦 Testing dependency installation..."
go mod download
go mod verify

# Simulate code quality checks
echo "🔍 Testing code quality checks..."
go vet ./...

# Check code formatting
echo "📝 Testing code formatting..."
UNFORMATTED=$(gofmt -s -l . | wc -l)
if [ "$UNFORMATTED" -gt 0 ]; then
    echo "❌ Code formatting issues found:"
    gofmt -s -l .
    exit 1
fi

# Run tests
echo "🧪 Running test suite..."
go test -v ./...

# Test building for current platform
echo "🔨 Testing build process..."
make build

# Test mobile build commands (without actual mobile packaging)
echo "📱 Testing mobile build commands..."
make build-android

# Check if icon files are accessible
echo "🖼️ Validating icon files..."
if [ ! -f "assets/icons/icon-192.png" ]; then
    echo "❌ Error: Required icon file assets/icons/icon-192.png not found"
    exit 1
fi

echo "✅ All CI/CD pipeline validation checks passed!"
echo ""
echo "🎯 Pipeline is ready for GitHub Actions execution"
echo "   - Code quality checks: ✅"
echo "   - Dependencies: ✅"
echo "   - Build process: ✅"
echo "   - Icon files: ✅"
echo "   - Mobile build setup: ✅"
echo ""
echo "🚀 The next push to main branch will trigger the full CI/CD pipeline"