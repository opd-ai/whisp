# Development Plan

## Project Overview

The Whisp project is a cross-platform secure messaging application with a **complete architectural foundation** already implemented. The project has achieved approximately **85% completion of core infrastructure** with clean interfaces, proper database design, and comprehensive build system. The primary gap is transitioning from placeholder implementations to fully functional features with real Tox library integration and complete UI implementation.

**Current state**: Well-architected foundation with placeholder Tox integration, structured UI components, and comprehensive tooling.  
**Completion percentage**: 85% foundation, 15% functional implementation  
**Critical path**: Tox library integration → UI completion → Platform builds → Security hardening

## Codebase Analysis

### Existing Components
- **Core Application Architecture** ✅: Complete with clean interfaces (`internal/core/app.go`)
- **Platform Detection System** ✅: Runtime platform adaptation (`ui/adaptive/platform.go`)
- **Database Layer** ✅: SQLite/SQLCipher schema with encryption support (`internal/storage/database.go`)
- **Message Management** ✅: Full CRUD operations, history, search interfaces (`internal/core/message/manager.go`)
- **Contact Management** ✅: Friend requests, status tracking, verification (`internal/core/contact/manager.go`)
- **Security Framework** ✅: Encryption, key derivation, secure storage (`internal/core/security/manager.go`)
- **Build System** ✅: Cross-platform Make-based build with packaging (`Makefile`, `scripts/`)
- **Tox Integration Interface** 🔄: Complete interface with placeholder implementation (`internal/core/tox/manager.go`)
- **UI Component Structure** 🔄: Fyne components with incomplete implementation (`ui/shared/components.go`)
- **Configuration System** ✅: YAML-based configuration with platform paths (`config.yaml`)

### Missing Components
- **Real Tox Protocol**: Placeholder implementation needs real `github.com/opd-ai/toxcore` integration
- **Complete UI Implementation**: Chat view, contact dialogs, settings panels need full Fyne widgets
- **File Transfer System**: Interface exists but file handling logic incomplete
- **Notification System**: Platform-specific notification implementations missing
- **Biometric Authentication**: Mobile platform biometric integration missing
- **App Store Packaging**: Distribution packages for mobile platforms missing

## Step-by-Step Implementation Plan

### Phase 1: Foundation Completion (Priority: Critical)

#### 1. **Replace Tox Placeholder Implementation**
   - Description: Integrate real `github.com/opd-ai/toxcore` library replacing placeholder methods
   - Files affected: `internal/core/tox/manager.go`, `go.mod`
   - Dependencies: Update toxcore library to latest version, verify API compatibility
   - Estimated time: 12 hours
   - Success criteria: Real Tox instance creation, friend requests work, basic messaging functional

#### 2. **Implement File I/O for Tox State Management**
   - Description: Complete the `save()` and `loadSavedata()` methods with actual file operations
   - Files affected: `internal/core/tox/manager.go` (lines 370-385)
   - Dependencies: File system permissions, encryption key from security manager
   - Estimated time: 4 hours
   - Success criteria: Tox state persists across application restarts, encrypted savedata files

#### 3. **Complete Database Encryption Integration**
   - Description: Integrate SQLCipher for database encryption using security manager keys
   - Files affected: `internal/storage/database.go`, `internal/core/security/manager.go`
   - Dependencies: SQLCipher bindings, key derivation from security manager
   - Estimated time: 8 hours
   - Success criteria: Database files are encrypted, performance impact < 10%

#### 4. **Implement Core Message Persistence**
   - Description: Complete database operations for message storage and retrieval
   - Files affected: `internal/core/message/manager.go` (saveMessage, loadMessages methods)
   - Dependencies: Database schema finalization, UUID generation
   - Estimated time: 6 hours
   - Success criteria: Messages persist across sessions, search functionality works

### Phase 2: Core Features (Priority: High)

#### 5. **Complete Chat View Implementation**
   - Description: Finish chat interface with message display, input handling, real-time updates
   - Files affected: `ui/shared/components.go` (ChatView struct, lines 1-130)
   - Dependencies: Message manager integration, Fyne widget customization
   - Estimated time: 16 hours
   - Success criteria: Messages display correctly, input sends messages, scroll behavior works

#### 6. **Implement Add Friend Dialog**
   - Description: Create modal dialog for adding friends via Tox ID with validation
   - Files affected: `ui/shared/components.go` (showAddFriendDialog method, line 187)
   - Dependencies: Tox ID validation, contact manager integration
   - Estimated time: 8 hours
   - Success criteria: Dialog appears, validates Tox IDs, successfully adds friends

#### 7. **Complete Contact List Integration**
   - Description: Connect contact list to real contact manager data with real-time updates
   - Files affected: `ui/shared/components.go` (ContactList, RefreshContacts method, line 195)
   - Dependencies: Contact manager callbacks, status change notifications
   - Estimated time: 10 hours
   - Success criteria: Contacts display correctly, status updates in real-time, selection works

#### 8. **Implement Settings Panel**
   - Description: Create settings interface for configuration, preferences, and security options
   - Files affected: `ui/adaptive/ui.go` (createMenuBar method, line 153), new `ui/shared/settings.go`
   - Dependencies: Configuration system integration, platform-specific settings
   - Estimated time: 12 hours
   - Success criteria: Settings persist, platform adaptation works, security options functional

### Phase 3: Platform Integration (Priority: High)

#### 9. **Complete Desktop UI Implementation**
   - Description: Finalize desktop-specific features like menus, keyboard shortcuts, window management
   - Files affected: `ui/adaptive/ui.go` (createDesktopLayout, createMenuBar methods)
   - Dependencies: Fyne menu system, keyboard event handling
   - Estimated time: 14 hours
   - Success criteria: Menu bar functional, keyboard shortcuts work, window state persists

#### 10. **Implement Mobile UI Adaptations**
   - Description: Complete mobile-specific UI patterns, gestures, and navigation
   - Files affected: `ui/adaptive/ui.go` (createMobileLayout method), `ui/adaptive/platform.go`
   - Dependencies: Mobile platform detection, touch gesture handling
   - Estimated time: 16 hours
   - Success criteria: Touch navigation works, mobile layouts adapt correctly, performance acceptable

#### 11. **Platform-Specific Notification System**
   - Description: Implement native notifications for each platform (Windows, macOS, Linux, Android, iOS)
   - Files affected: New `platform/notifications/` directory with platform-specific implementations
   - Dependencies: Platform notification APIs, permission handling
   - Estimated time: 20 hours
   - Success criteria: Notifications appear natively on each platform, respect user preferences

#### 12. **Implement Secure Storage Integration**
   - Description: Connect security manager to platform-specific secure storage (Keychain, Credential Manager, etc.)
   - Files affected: `internal/core/security/manager.go`, new platform-specific storage files
   - Dependencies: Platform-specific secure storage APIs
   - Estimated time: 18 hours
   - Success criteria: Keys stored securely per platform, biometric authentication works on mobile

### Phase 4: Advanced Features (Priority: Medium)

#### 13. **File Transfer Implementation**
   - Description: Complete file sending/receiving with progress tracking and resumption
   - Files affected: `internal/core/message/manager.go` (file transfer methods), new `internal/core/transfer/` package
   - Dependencies: Tox file transfer protocol, file system operations, progress callbacks
   - Estimated time: 24 hours
   - Success criteria: Files transfer reliably, progress indication works, large files supported

#### 14. **Message Search and History**
   - Description: Implement full-text search across message history with performance optimization
   - Files affected: `internal/core/message/manager.go` (SearchMessages method), database schema updates
   - Dependencies: SQLite FTS extension, indexing strategy
   - Estimated time: 12 hours
   - Success criteria: Search is fast (<100ms), results are accurate, handles large message history

#### 15. **Voice Message Support**
   - Description: Add voice message recording, playback, and waveform visualization
   - Files affected: New `internal/core/voice/` package, UI components for recording/playback
   - Dependencies: Audio recording/playback libraries, compression, UI controls
   - Estimated time: 28 hours
   - Success criteria: Voice messages record/playback correctly, file sizes reasonable, cross-platform support

#### 16. **Theme System Implementation**
   - Description: Complete light/dark/system theme support with custom color schemes
   - Files affected: `ui/adaptive/ui.go`, new `ui/theme/` package, configuration integration
   - Dependencies: Fyne theme system, system theme detection, user preferences
   - Estimated time: 14 hours
   - Success criteria: Themes switch correctly, system theme auto-detection works, custom themes supported

## Technical Considerations

### Architecture Decisions Needed
- **Tox Library Version**: Verify compatibility with `github.com/opd-ai/toxcore` latest version
- **Database Migration Strategy**: Plan for schema updates and data migration
- **Mobile Platform Build**: Decide on Fyne mobile vs native UI approach for iOS/Android
- **File Storage Strategy**: Implement secure file storage with encryption for media files

### Technology Stack Gaps
- **Mobile Biometric Libraries**: Platform-specific biometric authentication integration
- **Notification Libraries**: Native notification system integration per platform
- **Audio Processing**: Voice message recording/playback library selection
- **File System Encryption**: Secure file storage for media and attachments

### Integration Requirements
- **CI/CD Pipeline**: GitHub Actions for automated building and testing
- **Code Signing**: Platform-specific code signing for distribution
- **App Store Compliance**: Ensure compliance with app store requirements
- **Performance Monitoring**: Add telemetry for performance optimization

## Risk Assessment

### High Risk
- **Tox Library API Changes**: Real library may have different API than placeholder
  - *Mitigation*: Create adapter interfaces, thorough integration testing
- **Mobile Platform Restrictions**: iOS/Android app store approval challenges
  - *Mitigation*: Early compliance review, alternative distribution channels
- **Performance on Mobile**: Resource constraints may affect functionality
  - *Mitigation*: Performance profiling, optimization strategies, feature flags

### Medium Risk
- **Cross-Platform UI Consistency**: Fyne limitations on platform adaptation
  - *Mitigation*: Custom widget development, platform-specific UI branches
- **Database Migration Complexity**: Schema changes may break existing data
  - *Mitigation*: Migration testing, backup strategies, version compatibility
- **File Transfer Reliability**: Large file transfers may fail on poor connections
  - *Mitigation*: Resume capability, chunk verification, fallback protocols

### Low Risk
- **Build System Complexity**: Multiple platform builds may be fragile
  - *Mitigation*: Containerized builds, comprehensive testing, documentation
- **Configuration Management**: User preferences may not persist correctly
  - *Mitigation*: Configuration validation, default value handling, migration support

## Timeline

**Total estimated completion**: 12-14 weeks

### Critical Milestones

- **Week 2**: Phase 1 complete - Real Tox integration working, basic messaging functional
- **Week 6**: Phase 2 complete - Core UI implemented, friend management working
- **Week 10**: Phase 3 complete - Platform builds working, notifications implemented
- **Week 14**: Phase 4 complete - Advanced features implemented, ready for distribution

### Development Phases Timeline

- **Phase 1** (Foundation): 2 weeks - Critical path items, blockers resolved
- **Phase 2** (Core Features): 4 weeks - Basic functionality complete and testable
- **Phase 3** (Platform Integration): 4 weeks - All platforms building and working
- **Phase 4** (Advanced Features): 4 weeks - Polish and advanced functionality

### Quality Gates
- **End of Phase 1**: Core messaging works reliably
- **End of Phase 2**: Basic application usable for daily messaging
- **End of Phase 3**: All platforms build and install correctly
- **End of Phase 4**: Feature-complete application ready for public release

The project has an excellent foundation and clear path to completion. The architecture is sound, dependencies are manageable, and the modular design enables incremental development with testable milestones.
