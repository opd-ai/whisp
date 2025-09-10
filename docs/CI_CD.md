# CI/CD Pipeline Documentation

This document describes the GitHub Actions CI/CD pipeline for the Whisp messaging application.

## Overview

The project uses three main GitHub Actions workflows:

1. **CI Pipeline** (`ci.yml`) - Fast feedback on code quality
2. **Build Pipeline** (`build.yml`) - Cross-platform package building
3. **Dependency Updates** (`deps.yml`) - Automated dependency management

## Workflows

### 1. Continuous Integration (ci.yml)

**Triggers:**
- Push to `main`, `develop`, `feature/*` branches
- Pull requests to `main`, `develop`

**Jobs:**
- **Validate**: Code formatting, vet, tests, security scanning
- **Build Check**: Verify builds for major platforms
- **Security**: Gosec security scanner and vulnerability checking

**Features:**
- Fast feedback (< 5 minutes)
- Code quality enforcement
- Multi-platform build verification
- Security vulnerability scanning

### 2. Build and Package (build.yml)

**Triggers:**
- Push to `main`, `develop` branches
- Pull requests to `main`
- Release publications

**Platform Matrix:**
- **Linux**: amd64, arm64
- **Windows**: amd64
- **macOS**: amd64, arm64
- **Android**: APK
- **iOS**: IPA (macOS runners only)

**Build Features:**
- Native CGO compilation for GUI components
- Proper ldflags injection for version information
- Artifact collection and storage
- Automatic release creation for tags

**Artifacts:**
- Desktop executables for all platforms
- Mobile packages (APK/IPA)
- Compressed archives for releases

### 3. Dependency Updates (deps.yml)

**Triggers:**
- Weekly schedule (Mondays 9 AM UTC)
- Manual workflow dispatch

**Features:**
- Automatic Go module updates
- Test verification before PR creation
- Security vulnerability checking
- Automated pull request creation

## Build Matrix Strategy

The pipeline uses GitHub's matrix strategy for efficient parallel builds:

```yaml
strategy:
  matrix:
    arch: [amd64, arm64]  # For Linux/macOS
    target:               # For build verification
      - { goos: linux, goarch: amd64 }
      - { goos: windows, goarch: amd64 }
      - { goos: darwin, goarch: amd64 }
      - { goos: darwin, goarch: arm64 }
```

## Cross-Platform Considerations

### CGO and GUI Dependencies

Whisp uses Fyne for GUI components, which requires CGO. This affects cross-compilation:

- **Linux**: Native builds on Ubuntu runners with system dependencies
- **Windows**: Native builds on Windows runners with CGO enabled
- **macOS**: Native builds on macOS runners for both architectures
- **Mobile**: Uses Fyne's mobile packaging tools

### Dependencies by Platform

**Linux:**
```bash
sudo apt-get install -y gcc libc6-dev libgl1-mesa-dev xorg-dev
```

**Windows:**
- Uses Windows runners with Go's default CGO setup
- No additional system dependencies required

**macOS:**
- Uses macOS runners with Xcode toolchain
- Supports both Intel and Apple Silicon

**Mobile:**
- Android: Requires Java 17 for Android toolchain
- iOS: Requires macOS with Xcode (developer account needed for distribution)

## Security Features

### Code Scanning
- **Gosec**: Static security analysis for Go code
- **govulncheck**: Known vulnerability database checking
- **Dependency scanning**: Automated security updates

### Build Security
- **Minimal permissions**: Workflows use least-privilege access
- **Artifact signing**: Prepared for code signing integration
- **Secure secrets**: Uses GitHub secrets for sensitive data

## Performance Optimizations

### Caching Strategy
- **Go module cache**: Speeds up dependency downloads
- **Build cache**: Reduces compilation time
- **Artifact retention**: 30-day retention for development builds

### Parallel Execution
- **Matrix builds**: All platforms build simultaneously
- **Job dependencies**: Optimized dependency graph
- **Fast feedback**: CI completes in under 5 minutes

## Release Automation

### Automatic Releases
When a tag is pushed:
1. All platform builds execute
2. Artifacts are downloaded and archived
3. Release is created with generated notes
4. All platform packages are attached

### Versioning
Version information is injected via ldflags:
- `main.version`: Git tag or commit hash
- `main.buildTime`: ISO 8601 build timestamp
- `main.gitCommit`: Short commit hash

Example:
```bash
go build -ldflags "-X main.version=v1.0.0 -X main.buildTime=2025-09-09T12:00:00Z -X main.gitCommit=abc123"
```

## Local Development

### Testing the CI Pipeline Locally

```bash
# Run the same tests as CI
make test
make lint

# Test cross-platform builds (requires platform-specific tools)
make build-all

# Verify workflow syntax
go test ./cmd/whisp -run TestWorkflow -v
```

### Build Requirements

Refer to the main README for platform-specific build requirements. The CI pipeline matches these requirements exactly.

## Troubleshooting

### Common Build Issues

**CGO Cross-Compilation Errors:**
- Solution: Use native runners for each platform
- Fallback: Build with `CGO_ENABLED=0` for syntax checking

**Mobile Build Failures:**
- Android: Ensure Java 17 is available
- iOS: Requires macOS runner and valid developer setup

**Dependency Issues:**
- Update Go modules: `go get -u ./... && go mod tidy`
- Check vulnerabilities: `govulncheck ./...`

### Workflow Debugging

View detailed logs in GitHub Actions tab:
1. Navigate to repository → Actions
2. Select failed workflow run
3. Expand job and step logs
4. Check artifact uploads for build outputs

## Future Enhancements

### Planned Improvements
- **Code signing**: Platform-specific binary signing
- **App store automation**: Automated app store uploads
- **Performance benchmarks**: Automated performance regression testing
- **Multi-architecture**: ARM64 Linux builds on native runners

### Integration Opportunities
- **Docker builds**: Containerized build environments
- **Nix builds**: Reproducible builds with Nix
- **Release notes**: Automated changelog generation
- **Security scanning**: Enhanced SAST/DAST integration

## Configuration

### Environment Variables
- `GO_VERSION`: Go version for all builds (currently 1.21)
- `APP_NAME`: Application name for artifacts (whisp)

### Secrets (when needed)
- `ANDROID_SIGNING_KEY`: For signed Android releases
- `APPLE_DEVELOPER_*`: For iOS app store releases
- `CODE_SIGNING_*`: For desktop application signing

This CI/CD pipeline provides comprehensive automated testing, building, and release management for the Whisp cross-platform messaging application.
