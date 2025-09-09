#!/bin/bash
# macOS build script for Whisp

set -e

echo "🍎 Building Whisp for macOS..."

# Check if we're on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    echo "❌ macOS builds require macOS"
    exit 1
fi

# Build configuration
APP_NAME="Whisp"
BUNDLE_ID="com.opd-ai.whisp"
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}

# Build flags
LDFLAGS="-ldflags -X main.version=$VERSION -X main.buildTime=$BUILD_TIME -X main.gitCommit=$GIT_COMMIT"

# Create build directory
mkdir -p build/macos

echo "Building universal binary..."

# Build for Intel and Apple Silicon
echo "Building for Intel (amd64)..."
GOOS=darwin GOARCH=amd64 go build $LDFLAGS -o "build/macos/${APP_NAME}-amd64" ./cmd/whisp

echo "Building for Apple Silicon (arm64)..."
GOOS=darwin GOARCH=arm64 go build $LDFLAGS -o "build/macos/${APP_NAME}-arm64" ./cmd/whisp

# Create universal binary
echo "Creating universal binary..."
lipo -create -output "build/macos/${APP_NAME}" \
    "build/macos/${APP_NAME}-amd64" \
    "build/macos/${APP_NAME}-arm64"

# Create app bundle
echo "Creating app bundle..."
APP_BUNDLE="build/macos/${APP_NAME}.app"
mkdir -p "$APP_BUNDLE/Contents/MacOS"
mkdir -p "$APP_BUNDLE/Contents/Resources"

# Copy binary
cp "build/macos/${APP_NAME}" "$APP_BUNDLE/Contents/MacOS/"

# Create Info.plist
cat > "$APP_BUNDLE/Contents/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>${APP_NAME}</string>
    <key>CFBundleIdentifier</key>
    <string>${BUNDLE_ID}</string>
    <key>CFBundleName</key>
    <string>${APP_NAME}</string>
    <key>CFBundleVersion</key>
    <string>${VERSION}</string>
    <key>CFBundleShortVersionString</key>
    <string>${VERSION}</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleSignature</key>
    <string>????</string>
    <key>LSMinimumSystemVersion</key>
    <string>11.0</string>
    <key>NSHumanReadableCopyright</key>
    <string>Copyright © 2024 OPD AI. All rights reserved.</string>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
EOF

# Copy icon if available
if [ -f "resources/icons/icon.icns" ]; then
    cp "resources/icons/icon.icns" "$APP_BUNDLE/Contents/Resources/"
    echo "    <key>CFBundleIconFile</key>" >> "$APP_BUNDLE/Contents/Info.plist.tmp"
    echo "    <string>icon</string>" >> "$APP_BUNDLE/Contents/Info.plist.tmp"
fi

# Sign the application (if developer certificate is available)
if command -v codesign &> /dev/null; then
    echo "Signing application..."
    if [ -n "$DEVELOPER_ID" ]; then
        codesign --force --deep --sign "$DEVELOPER_ID" "$APP_BUNDLE"
        echo "✅ Application signed with Developer ID: $DEVELOPER_ID"
    else
        codesign --force --deep --sign - "$APP_BUNDLE"
        echo "⚠️  Application signed with ad-hoc signature (development only)"
    fi
fi

# Create DMG (if create-dmg is available)
if command -v create-dmg &> /dev/null; then
    echo "Creating DMG..."
    create-dmg \
        --volname "$APP_NAME" \
        --window-pos 200 120 \
        --window-size 600 300 \
        --icon-size 100 \
        --icon "$APP_NAME.app" 175 120 \
        --hide-extension "$APP_NAME.app" \
        --app-drop-link 425 120 \
        "build/whisp-macos-${VERSION}.dmg" \
        "$APP_BUNDLE"
else
    echo "⚠️  create-dmg not found, creating zip package instead"
    cd build/macos
    zip -r "../whisp-macos-${VERSION}.zip" "${APP_NAME}.app"
    cd ../..
fi

# Cleanup intermediate files
rm -f "build/macos/${APP_NAME}-amd64"
rm -f "build/macos/${APP_NAME}-arm64"
rm -f "build/macos/${APP_NAME}"

echo "✅ macOS build complete!"
echo "   App Bundle: $APP_BUNDLE"
if [ -f "build/whisp-macos-${VERSION}.dmg" ]; then
    echo "   DMG: build/whisp-macos-${VERSION}.dmg"
elif [ -f "build/whisp-macos-${VERSION}.zip" ]; then
    echo "   Package: build/whisp-macos-${VERSION}.zip"
fi
