#!/bin/bash
# Linux build script for Whisp

set -e

echo "🐧 Building Whisp for Linux..."

# Build configuration
APP_NAME="whisp"
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}

# Build flags
LDFLAGS="-ldflags -X main.version=$VERSION -X main.buildTime=$BUILD_TIME -X main.gitCommit=$GIT_COMMIT"

# Create build directory
mkdir -p build/linux

echo "Building for multiple architectures..."

# Build for amd64
echo "Building for x86_64..."
GOOS=linux GOARCH=amd64 go build $LDFLAGS -o "build/linux/${APP_NAME}-amd64" ./cmd/whisp

# Build for arm64
echo "Building for ARM64..."
GOOS=linux GOARCH=arm64 go build $LDFLAGS -o "build/linux/${APP_NAME}-arm64" ./cmd/whisp

# Create AppImage (if appimagetool is available)
create_appimage() {
    local arch=$1
    local binary=$2
    
    echo "Creating AppImage for $arch..."
    
    # Create AppDir structure
    APPDIR="build/linux/Whisp-$arch.AppDir"
    mkdir -p "$APPDIR/usr/bin"
    mkdir -p "$APPDIR/usr/share/applications"
    mkdir -p "$APPDIR/usr/share/icons/hicolor/256x256/apps"
    
    # Copy binary
    cp "$binary" "$APPDIR/usr/bin/whisp"
    chmod +x "$APPDIR/usr/bin/whisp"
    
    # Create desktop file
    cat > "$APPDIR/usr/share/applications/whisp.desktop" << EOF
[Desktop Entry]
Type=Application
Name=Whisp
Comment=Secure cross-platform messaging
Exec=whisp
Icon=whisp
Categories=Network;InstantMessaging;
Terminal=false
EOF
    
    # Copy icon if available
    if [ -f "assets/icons/icon.png" ]; then
        cp "assets/icons/icon.png" "$APPDIR/usr/share/icons/hicolor/256x256/apps/whisp.png"
    elif [ -f "assets/icons/icon-192.png" ]; then
        cp "assets/icons/icon-192.png" "$APPDIR/usr/share/icons/hicolor/256x256/apps/whisp.png"
    else
        echo "⚠️  No suitable icon found for AppImage"
    fi
    
    # Create AppRun
    cat > "$APPDIR/AppRun" << 'EOF'
#!/bin/bash
HERE="$(dirname "$(readlink -f "${0}")")"
export PATH="${HERE}/usr/bin:${PATH}"
exec "${HERE}/usr/bin/whisp" "$@"
EOF
    chmod +x "$APPDIR/AppRun"
    
    # Symlink desktop file and icon
    ln -sf "usr/share/applications/whisp.desktop" "$APPDIR/"
    if [ -f "$APPDIR/usr/share/icons/hicolor/256x256/apps/whisp.png" ]; then
        ln -sf "usr/share/icons/hicolor/256x256/apps/whisp.png" "$APPDIR/"
    fi
    
    # Create AppImage if appimagetool is available
    if command -v appimagetool &> /dev/null; then
        appimagetool "$APPDIR" "build/whisp-linux-$arch-${VERSION}.AppImage"
        echo "✅ AppImage created: build/whisp-linux-$arch-${VERSION}.AppImage"
    else
        echo "⚠️  appimagetool not found, AppDir created at $APPDIR"
    fi
}

# Create AppImages
create_appimage "amd64" "build/linux/${APP_NAME}-amd64"
create_appimage "arm64" "build/linux/${APP_NAME}-arm64"

# Create Flatpak manifest (if flatpak is available)
create_flatpak_manifest() {
    echo "Creating Flatpak manifest..."
    
    mkdir -p build/linux/flatpak
    
    cat > "build/linux/flatpak/com.opd-ai.whisp.yml" << EOF
app-id: com.opd-ai.whisp
runtime: org.freedesktop.Platform
runtime-version: '23.08'
sdk: org.freedesktop.Sdk
command: whisp

finish-args:
  - --share=network
  - --share=ipc
  - --socket=fallback-x11
  - --socket=wayland
  - --device=dri
  - --filesystem=xdg-download:rw
  - --talk-name=org.freedesktop.Notifications

modules:
  - name: whisp
    buildsystem: simple
    build-commands:
      - install -Dm755 whisp /app/bin/whisp
      - install -Dm644 whisp.desktop /app/share/applications/com.opd-ai.whisp.desktop
      - install -Dm644 whisp.png /app/share/icons/hicolor/256x256/apps/com.opd-ai.whisp.png
    sources:
      - type: file
        path: ../whisp-amd64
        dest-filename: whisp
      - type: file
        path: ../../resources/whisp.desktop
      - type: file
        path: ../../assets/icons/icon-192.png
        dest-filename: whisp.png
EOF
    
    echo "✅ Flatpak manifest created: build/linux/flatpak/com.opd-ai.whisp.yml"
}

# Create tar.gz packages
echo "Creating tar.gz packages..."
cd build/linux
tar -czf "whisp-linux-amd64-${VERSION}.tar.gz" "${APP_NAME}-amd64"
tar -czf "whisp-linux-arm64-${VERSION}.tar.gz" "${APP_NAME}-arm64"
cd ../..

# Create Flatpak manifest
create_flatpak_manifest

echo "✅ Linux build complete!"
echo "   Binaries:"
echo "     - build/linux/${APP_NAME}-amd64"
echo "     - build/linux/${APP_NAME}-arm64"
echo "   Packages:"
echo "     - build/linux/whisp-linux-amd64-${VERSION}.tar.gz"
echo "     - build/linux/whisp-linux-arm64-${VERSION}.tar.gz"

if [ -f "build/whisp-linux-amd64-${VERSION}.AppImage" ]; then
    echo "   AppImages:"
    echo "     - build/whisp-linux-amd64-${VERSION}.AppImage"
    echo "     - build/whisp-linux-arm64-${VERSION}.AppImage"
fi
