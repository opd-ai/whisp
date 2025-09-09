# Whisp Architecture Document

## System Overview

Whisp is designed as a modular, cross-platform messaging application with a clean separation of concerns. The architecture prioritizes code reuse, security, and platform adaptation.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    User Interface Layer                     │
├─────────────────────┬───────────────────┬───────────────────┤
│   Desktop UI        │    Mobile UI      │   Adaptive UI     │
│  (Windows/Mac/Linux)│  (Android/iOS)    │   Components      │
└─────────────────────┴───────────────────┴───────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                    Core Application Layer                   │
├─────────────────────┬───────────────────┬───────────────────┤
│   Message Manager   │  Contact Manager  │ Security Manager  │
└─────────────────────┴───────────────────┴───────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                    Protocol Layer                          │
├─────────────────────┬───────────────────┬───────────────────┤
│   Tox Manager       │   Network Layer   │   Crypto Layer    │
└─────────────────────┴───────────────────┴───────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                    Storage Layer                           │
├─────────────────────┬───────────────────┬───────────────────┤
│   Database          │   File System     │   Secure Storage  │
│   (SQLite/Cipher)   │   (Media/Files)   │   (Keys/Config)   │
└─────────────────────┴───────────────────┴───────────────────┘
```

## Component Architecture

### 1. User Interface Layer

#### Adaptive UI System
The UI layer adapts to platform conventions while maintaining functional consistency.

**Key Components:**
- **Platform Detection**: Runtime detection of execution environment
- **Layout Adaptation**: Different layouts for mobile vs desktop
- **Theme System**: Light/dark/system theme support with platform colors
- **Input Handling**: Touch, mouse, keyboard input normalization

**Implementation:**
```go
type Platform interface {
    IsMobile() bool
    IsDesktop() bool
    GetSystemTheme() Theme
    GetNativeControls() ControlSet
}

type AdaptiveUI struct {
    platform Platform
    theme    *ThemeManager
    layout   LayoutManager
}
```

#### Desktop UI (Windows/macOS/Linux)
- **Layout**: Split-pane with resizable contact list and chat area
- **Navigation**: Menu bar, keyboard shortcuts, context menus
- **Integration**: System tray, native notifications, file associations
- **Window Management**: Remember size/position, minimize to tray

#### Mobile UI (Android/iOS)
- **Layout**: Tab-based navigation with bottom tab bar
- **Navigation**: Swipe gestures, pull-to-refresh, long-press menus
- **Integration**: Share sheets, background notifications, quick actions
- **Responsive Design**: Adapts to different screen sizes and orientations

### 2. Core Application Layer

#### Message Manager
Handles all message-related operations including sending, receiving, storage, and presentation.

**Responsibilities:**
- Message sending with delivery tracking
- Incoming message processing
- Message history management
- Search and filtering
- Message editing and deletion
- File attachment handling

**Key Interfaces:**
```go
type MessageManager interface {
    SendMessage(friendID uint32, content string, msgType MessageType) (*Message, error)
    GetMessages(friendID uint32, limit, offset int) ([]*Message, error)
    SearchMessages(query string) ([]*Message, error)
    EditMessage(messageID int64, newContent string) error
    DeleteMessage(messageID int64) error
}
```

#### Contact Manager
Manages friend relationships, contact information, and presence status.

**Responsibilities:**
- Contact addition and removal
- Friend request handling
- Contact information storage
- Presence status tracking
- Contact verification
- Blocking and privacy controls

**Key Interfaces:**
```go
type ContactManager interface {
    AddContact(toxID, message string) (*Contact, error)
    GetContacts() []*Contact
    UpdateContactStatus(friendID uint32, status Status)
    BlockContact(friendID uint32) error
    VerifyContact(friendID uint32, method VerificationMethod) error
}
```

#### Security Manager
Handles encryption, authentication, and security-related operations.

**Responsibilities:**
- Local data encryption
- Key management
- Biometric authentication
- Secure storage
- Privacy settings enforcement
- Security audit logging

**Key Interfaces:**
```go
type SecurityManager interface {
    EncryptData(data []byte, context string) ([]byte, error)
    DecryptData(encryptedData []byte, context string) ([]byte, error)
    AuthenticateUser(method AuthMethod) error
    SecureStore(key, value string) error
    SecureRetrieve(key string) (string, error)
}
```

### 3. Protocol Layer

#### Tox Manager
Provides a clean abstraction over the Tox protocol implementation.

**Responsibilities:**
- Tox instance lifecycle management
- Network bootstrapping
- Friend management via Tox protocol
- Message sending/receiving
- File transfer coordination
- Connection status monitoring

**Architecture:**
```go
type ToxManager struct {
    instance     *toxcore.Tox
    config       *ToxConfig
    saveFile     string
    callbacks    *CallbackManager
    networking   *NetworkManager
}

type CallbackManager struct {
    friendRequest    func([32]byte, string)
    friendMessage    func(uint32, string)
    connectionStatus func(uint32, ConnectionStatus)
    fileTransfer     func(uint32, FileTransfer)
}
```

#### Network Layer
Handles network connectivity, proxy support, and connection management.

**Features:**
- Bootstrap node management
- Proxy support (SOCKS5, HTTP)
- IPv4/IPv6 dual stack
- Connection quality monitoring
- Network change detection
- Offline message queuing

#### Crypto Layer
Implements additional cryptographic operations beyond Tox's built-in encryption.

**Features:**
- Local database encryption (SQLCipher)
- Key derivation and management
- Random number generation
- Hash functions for integrity
- Digital signatures for verification

### 4. Storage Layer

#### Database (SQLite + SQLCipher)
Provides encrypted, reliable storage for application data.

**Schema Design:**
```sql
-- Contacts table
CREATE TABLE contacts (
    id INTEGER PRIMARY KEY,
    tox_id TEXT UNIQUE,
    public_key BLOB NOT NULL,
    friend_id INTEGER UNIQUE,
    name TEXT NOT NULL,
    status_message TEXT,
    avatar BLOB,
    status INTEGER,
    is_blocked BOOLEAN DEFAULT 0,
    created_at DATETIME,
    updated_at DATETIME
);

-- Messages table
CREATE TABLE messages (
    id INTEGER PRIMARY KEY,
    friend_id INTEGER REFERENCES contacts(friend_id),
    content TEXT NOT NULL,
    message_type INTEGER,
    is_outgoing BOOLEAN,
    timestamp DATETIME,
    delivered_at DATETIME,
    read_at DATETIME,
    edited_at DATETIME,
    file_path TEXT,
    reply_to_id INTEGER REFERENCES messages(id)
);

-- Settings table for configuration persistence
CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value TEXT,
    updated_at DATETIME
);
```

#### File System Storage
Manages media files, downloads, and temporary data.

**Structure:**
```
~/.local/share/whisp/  (Linux)
~/Library/Application Support/Whisp/  (macOS)  
%APPDATA%/Whisp/  (Windows)
├── whisp.db          # Main database
├── media/            # Received media files
├── downloads/        # File downloads
├── cache/           # Temporary data
├── logs/            # Application logs
└── backup/          # Database backups
```

#### Secure Storage
Platform-specific secure storage for sensitive data.

**Platform Integration:**
- **Windows**: Windows Credential Manager / DPAPI
- **macOS**: Keychain Services
- **Linux**: Secret Service API / gnome-keyring
- **Android**: Android Keystore
- **iOS**: iOS Keychain

## Data Flow Architecture

### Message Flow
```
User Input → UI Component → Core App → Message Manager → Tox Manager → Network
                                   ↓
Message Storage ← Database ← Storage Layer ← Message Manager
```

### Contact Management Flow
```
Add Contact → UI → Contact Manager → Tox Manager → Send Friend Request
                                  ↓
Contact Storage ← Database ← Contact Manager ← Friend Request Response
```

### Configuration Flow
```
Settings UI → Validation → Security Manager → Encrypted Storage
                        ↓
Application Config ← Config Manager ← Secure Retrieval
```

## Security Architecture

### Defense in Depth
1. **Network Layer**: Tox protocol end-to-end encryption
2. **Transport Layer**: TLS for bootstrap connections
3. **Application Layer**: Message authentication and integrity
4. **Storage Layer**: SQLCipher database encryption
5. **System Layer**: Platform-specific secure storage

### Key Management
```
Master Key (Platform Keystore)
    ├── Database Encryption Key
    ├── Configuration Encryption Key
    └── Backup Encryption Key
```

### Privacy Protection
- **No Metadata Logging**: Application doesn't log message content
- **Forward Secrecy**: Tox protocol provides perfect forward secrecy
- **Local Processing**: All data processing happens locally
- **Minimal Data Collection**: Only necessary data for functionality

## Platform-Specific Considerations

### Windows
- **UI Framework**: Native Windows controls with Fyne
- **Packaging**: MSI installer with proper Windows features
- **Integration**: Windows notifications, file associations
- **Security**: Windows Hello integration, DPAPI usage

### macOS
- **UI Framework**: Native macOS controls with Fyne
- **Packaging**: DMG with proper app bundle structure
- **Integration**: macOS notifications, Spotlight search
- **Security**: Touch ID integration, Keychain usage
- **Sandboxing**: App Store compatibility

### Linux
- **UI Framework**: GTK-based controls with Fyne
- **Packaging**: AppImage, Flatpak, native packages
- **Integration**: FreeDesktop notifications, desktop files
- **Security**: Secret Service API, system keyring

### Android
- **UI Framework**: Material Design with Fyne
- **Packaging**: APK/AAB with Play Store compliance
- **Integration**: Android notifications, sharing, intents
- **Background**: Foreground service for connectivity
- **Security**: Android Keystore, biometric authentication

### iOS
- **UI Framework**: iOS design patterns with Fyne
- **Packaging**: IPA with App Store compliance
- **Integration**: iOS notifications, sharing, shortcuts
- **Background**: Background app refresh limitations
- **Security**: iOS Keychain, Face ID/Touch ID

## Performance Considerations

### Memory Management
- **Message Pagination**: Load messages in chunks
- **Image Caching**: LRU cache for media files
- **Connection Pooling**: Reuse network connections
- **Garbage Collection**: Proper cleanup of resources

### Battery Optimization (Mobile)
- **Background Processing**: Minimize CPU usage when backgrounded
- **Network Efficiency**: Batch network operations
- **Wake Lock Management**: Careful use of device wake locks
- **Notification Batching**: Group notifications to reduce interruptions

### Startup Performance
- **Lazy Loading**: Load components as needed
- **Database Optimization**: Proper indexing and queries
- **Asset Bundling**: Embed critical resources
- **Parallel Initialization**: Initialize components concurrently

## Scalability Architecture

### Message History
- **Pagination**: Efficient message loading
- **Archiving**: Automatic old message cleanup
- **Search Indexing**: FTS for message search
- **Backup Strategy**: Incremental backups

### Contact Management
- **Contact Limits**: Handle large friend lists efficiently
- **Status Caching**: Cache contact status to reduce lookups
- **Avatar Optimization**: Resize and cache contact avatars

### File Transfers
- **Streaming**: Stream large files instead of loading in memory
- **Progress Tracking**: Real-time transfer progress
- **Resumption**: Resume interrupted transfers
- **Cleanup**: Automatic cleanup of failed transfers

This architecture provides a solid foundation for a secure, performant, and maintainable cross-platform messaging application that can scale to meet user demands while maintaining security and privacy.
