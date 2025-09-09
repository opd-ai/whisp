# Whisp

A secure, cross-platform messaging application built with Go, featuring end-to-end encryption via the Tox protocol and complete feature parity across Windows, macOS, Linux, Android, and iOS.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue.svg)
![Platform Support](https://img.shields.io/badge/platforms-windows%20%7C%20macos%20%7C%20linux%20%7C%20android%20%7C%20ios-lightgrey.svg)

## Features

### 🔒 Security First
- **End-to-end encryption** using the Tox protocol
- **No central servers** - completely peer-to-peer
- **Perfect forward secrecy** for all messages
- **Encrypted local storage** with SQLCipher
- **Biometric authentication** support
- **Disappearing messages** with custom timers

### 💬 Complete Messaging
- **Text messages** with markdown formatting
- **Voice messages** with playback controls
- **File sharing** up to 2GB with progress tracking
- **Rich media** support with in-app preview
- **Message editing** and deletion with history
- **Offline message** queuing and delivery

### 🌍 True Cross-Platform
- **100% feature parity** across all platforms
- **Native performance** with shared Go codebase
- **Adaptive UI** that follows platform conventions
- **Synchronized** experience across devices
- **Single codebase** with 85%+ code reuse

## Quick Start

### Prerequisites
- Go 1.21 or higher
- Platform-specific development tools:
  - **Windows**: Visual Studio 2019+ with C++ tools
  - **macOS**: Xcode 14+ and Command Line Tools
  - **Linux**: gcc, make, and development headers
  - **Android**: Android Studio with NDK
  - **iOS**: Xcode 14+ with valid developer account

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/whisp.git
cd whisp

# Install dependencies
go mod download

# Build for current platform
make build

# Run the application
./build/whisp
```

### Platform-Specific Builds

```bash
# Desktop platforms
make build-windows   # Creates .exe and installer
make build-macos     # Creates .app bundle
make build-linux     # Creates AppImage

# Mobile platforms
make build-android   # Creates APK/AAB
make build-ios       # Creates IPA (requires macOS)

# Build all platforms
make build-all
```

## Development

### Project Structure
```
whisp/
├── cmd/whisp/          # Application entry point
├── internal/           # Core business logic
│   ├── core/          # Tox protocol and messaging
│   ├── storage/       # Database and file handling
│   └── sync/          # Message synchronization
├── ui/                # User interface
│   ├── shared/        # Shared UI components
│   ├── adaptive/      # Platform adaptation layer
│   └── themes/        # Theme definitions
└── platform/          # Platform-specific code
```

### Running Tests
```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package tests
go test ./internal/core/message

# Run integration tests
make test-integration
```

### Development Mode
```bash
# Run with hot reload (desktop only)
make dev

# Run with debug logging
make run-debug

# Run mobile simulator
make run-android    # Requires Android emulator
make run-ios        # Requires iOS simulator (macOS only)
```

## Configuration

Whisp uses sensible defaults, but you can customize settings via `config.yaml`:

```yaml
# Network settings
network:
  bootstrap_nodes:
    - "tox.node1.example.com:33445:KEY1"
    - "tox.node2.example.com:33445:KEY2"
  enable_ipv6: true
  enable_udp: true
  enable_tcp: true
  proxy:
    type: "none"  # Options: none, socks5, http
    address: ""
    port: 0

# Storage settings
storage:
  data_dir: "~/.whisp"  # Platform-specific defaults used
  enable_encryption: true
  max_file_size: 2147483648  # 2GB in bytes

# UI settings
ui:
  theme: "system"  # Options: system, light, dark, amoled
  language: "en"
  font_size: "medium"
  enable_animations: true

# Privacy settings
privacy:
  save_message_history: true
  typing_indicators: true
  read_receipts: true
  auto_accept_files: false
  screenshot_protection: true
```

## Building from Source

### Desktop Build Requirements

**Windows:**
```powershell
# Install dependencies
choco install golang git visualstudio2019community

# Build
./scripts/build-windows.ps1
```

**macOS:**
```bash
# Install dependencies
brew install go git

# Build and sign
./scripts/build-macos.sh --sign "Developer ID"
```

**Linux:**
```bash
# Install dependencies (Ubuntu/Debian)
sudo apt-get install golang git build-essential libgl1-mesa-dev

# Build
./scripts/build-linux.sh
```

### Mobile Build Requirements

See [docs/MOBILE_BUILD.md](docs/MOBILE_BUILD.md) for detailed mobile build instructions.

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md).

### Development Workflow
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Style
- Follow standard Go formatting (`go fmt`)
- Pass all linters (`make lint`)
- Maintain test coverage above 80%
- Document exported functions

## Security

### Reporting Security Issues
Please report security issues to security@whisp.app. Do not create public issues for security vulnerabilities.

### Security Features
- **No phone numbers or email required**
- **No metadata collection**
- **Decentralized architecture**
- **Open source and auditable**
- **Regular security updates**

See [SECURITY.md](SECURITY.md) for details.

## Documentation

- [User Guide](docs/USER_GUIDE.md) - Complete feature documentation
- [API Reference](https://pkg.go.dev/github.com/yourusername/whisp) - Go package documentation
- [Protocol Spec](docs/PROTOCOL.md) - Tox protocol implementation details
- [Architecture](docs/ARCHITECTURE.md) - System design and decisions

## Support

- **Issues**: [GitHub Issues](https://github.com/yourusername/whisp/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/whisp/discussions)
- **Chat**: Join us on Whisp! ID: `XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX`

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Tox Protocol](https://tox.chat/) for the secure communication protocol
- [Fyne Framework](https://fyne.io/) for the cross-platform UI toolkit
- [opd-ai/toxcore](https://github.com/opd-ai/toxcore) for the Go Tox implementation
- All [contributors](https://github.com/yourusername/whisp/graphs/contributors) who have helped improve this project

---

<p align="center">
  Made with ❤️ by the Whisp Team
</p>