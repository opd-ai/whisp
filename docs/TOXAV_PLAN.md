# ToxAV Implementation Plan for Whisp

## Overview

This document outlines the comprehensive plan for implementing P2P voice and video calls in Whisp using the ToxAV protocol. The implementation requires coordinated development across two repositories:

1. **opd-ai/toxcore**: Add ToxAV bindings and Go wrapper functions
2. **opd-ai/whisp**: Implement call management system and UI integration

## Current State Analysis

### Existing Infrastructure
- ✅ **Audio System**: Mock audio recorder/player with WAV support (`internal/core/audio/`)
- ✅ **Tox Integration**: Complete Tox messaging and file transfer (`internal/core/tox/`)
- ✅ **UI Framework**: Adaptive UI system with platform detection
- ✅ **Database Layer**: SQLite with encrypted storage for call history
- ✅ **Notification System**: Cross-platform notifications for incoming calls

### Missing Components
- ❌ **ToxAV Bindings**: No audio/video calling support in toxcore library
- ❌ **Codec Support**: No audio/video encoding/decoding capabilities
- ❌ **Call State Management**: No call session handling
- ❌ **Real Audio/Video**: Current audio system is mock implementation

## Technical Requirements

### ToxAV Protocol Requirements
- **Audio Codec**: Opus codec for high-quality audio compression
- **Video Codec**: VP8 or H.264 for video compression
- **Network**: UDP-based RTP streaming over Tox network
- **Synchronization**: Audio/video synchronization and jitter buffering
- **Quality Adaptation**: Adaptive bitrate based on network conditions

### Performance Requirements
- **Audio Latency**: <150ms end-to-end latency for voice calls
- **Video Latency**: <200ms end-to-end latency for video calls
- **Bandwidth**: Adaptive bitrate 32-128 kbps audio, 100-2000 kbps video
- **CPU Usage**: <10% CPU usage for audio calls, <25% for video calls
- **Memory**: <50MB additional memory usage during active calls

### Security Requirements
- **End-to-End Encryption**: All audio/video data encrypted using Tox protocol
- **Perfect Forward Secrecy**: New keys for each call session
- **Authentication**: Caller identity verification through Tox friend system
- **Privacy**: No call metadata stored on external servers

## Implementation Plan

### Phase 1: ToxAV Core Integration (opd-ai/toxcore)

#### 1.1 ToxAV C Library Integration
**Objective**: Add ToxAV C library bindings to the toxcore Go wrapper

**Tasks**:
- [ ] **Add ToxAV Dependencies**
  - Include `libtoxav` in build system
  - Add Opus codec library (`libopus`)
  - Add VPX library for VP8 video codec
  - Update CGO build flags and linking

- [ ] **Create ToxAV Bindings** (`toxav.go`)
  ```go
  /*
  #cgo pkg-config: toxav opus vpx
  #include <tox/toxav.h>
  */
  import "C"
  
  type ToxAV struct {
      instance *C.ToxAV
      tox      *Tox
  }
  
  func NewToxAV(tox *Tox) (*ToxAV, error)
  func (av *ToxAV) Call(friendID uint32, audioBitrate, videoBitrate uint32) error
  func (av *ToxAV) Answer(friendID uint32, audioBitrate, videoBitrate uint32) error
  func (av *ToxAV) CallControl(friendID uint32, control CallControl) error
  ```

- [ ] **Implement Callback System**
  ```go
  type CallbacksAV struct {
      OnCall        func(friendID uint32, audioEnabled, videoEnabled bool)
      OnCallState   func(friendID uint32, state CallState)
      OnAudioFrame  func(friendID uint32, pcm []int16, sampleCount int, channels, sampleRate uint32)
      OnVideoFrame  func(friendID uint32, width, height uint16, y, u, v []byte)
  }
  ```

- [ ] **Audio/Video Frame APIs**
  ```go
  func (av *ToxAV) AudioSendFrame(friendID uint32, pcm []int16, sampleCount int, channels, sampleRate uint32) error
  func (av *ToxAV) VideoSendFrame(friendID uint32, width, height uint16, y, u, v []byte) error
  ```

**Libraries Required**:
- `libtoxav`: ToxAV C library
- `libopus`: Opus audio codec
- `libvpx`: VP8 video codec

**Estimated Time**: 2-3 weeks

#### 1.2 Go API Design
**Objective**: Create clean Go interfaces for ToxAV functionality

**Key Interfaces**:
```go
type AudioVideoManager interface {
    Call(friendID uint32, audio, video bool) (*CallSession, error)
    Answer(callID string, audio, video bool) error
    Hangup(callID string) error
    
    SendAudioFrame(callID string, frame AudioFrame) error
    SendVideoFrame(callID string, frame VideoFrame) error
    
    SetCallbacks(callbacks CallbacksAV)
}

type CallSession interface {
    ID() string
    FriendID() uint32
    State() CallState
    AudioEnabled() bool
    VideoEnabled() bool
    SetAudioEnabled(enabled bool) error
    SetVideoEnabled(enabled bool) error
}
```

**Estimated Time**: 1 week

### Phase 2: Audio/Video Codec Integration (opd-ai/whisp)

#### 2.1 Audio Processing System
**Objective**: Replace mock audio system with real codec support

**Components**:
- [ ] **Audio Capture** (`internal/core/audio/capture.go`)
  ```go
  type AudioCapture interface {
      Start(sampleRate, channels int) error
      Stop() error
      ReadFrame() (AudioFrame, error)
  }
  ```

- [ ] **Audio Playback** (`internal/core/audio/playback.go`)
  ```go
  type AudioPlayback interface {
      Start(sampleRate, channels int) error
      Stop() error
      WriteFrame(frame AudioFrame) error
  }
  ```

- [ ] **Opus Codec Integration**
  ```go
  // Library: github.com/hraban/opus
  // License: MIT
  // Import: "github.com/hraban/opus"
  // Why: Pure Go Opus codec implementation, well-maintained
  
  type OpusEncoder interface {
      Encode(pcm []int16) ([]byte, error)
  }
  
  type OpusDecoder interface {
      Decode(data []byte) ([]int16, error)
  }
  ```

**Libraries Required**:
```
Library: github.com/hraban/opus
License: MIT
Import: "github.com/hraban/opus"
Why: Pure Go Opus codec, no C dependencies, actively maintained

Library: github.com/gordonklaus/portaudio
License: MIT  
Import: "github.com/gordonklaus/portaudio"
Why: Cross-platform audio I/O, industry standard, well-documented
```

**Estimated Time**: 2 weeks

#### 2.2 Video Processing System
**Objective**: Implement video capture, encoding, and rendering

**Components**:
- [ ] **Video Capture** (`internal/core/video/capture.go`)
  ```go
  type VideoCapture interface {
      Start(width, height, fps int) error
      Stop() error
      ReadFrame() (VideoFrame, error)
  }
  ```

- [ ] **Video Rendering** (`internal/core/video/renderer.go`)
  ```go
  type VideoRenderer interface {
      Start() error
      Stop() error
      RenderFrame(frame VideoFrame) error
  }
  ```

- [ ] **VP8 Codec Integration**
  ```go
  // Library: github.com/pion/webrtc/v3 (includes VP8)
  // License: MIT
  // Import: "github.com/pion/webrtc/v3/pkg/media"
  // Why: Battle-tested WebRTC implementation with VP8 support
  
  type VP8Encoder interface {
      Encode(frame VideoFrame) ([]byte, error)
  }
  
  type VP8Decoder interface {
      Decode(data []byte) (VideoFrame, error)
  }
  ```

**Libraries Required**:
```
Library: github.com/pion/webrtc/v3
License: MIT
Import: "github.com/pion/webrtc/v3/pkg/media"
Why: Industry-standard WebRTC implementation with VP8/H.264 codecs

Library: github.com/vladimirvivien/go4vl
License: Apache 2.0
Import: "github.com/vladimirvivien/go4vl/v4l2"
Why: Video4Linux support for camera access on Linux

Library: github.com/kbinani/screenshot
License: MIT
Import: "github.com/kbinani/screenshot"
Why: Cross-platform screen capture for screen sharing
```

**Estimated Time**: 3 weeks

### Phase 3: Call Management System (opd-ai/whisp)

#### 3.1 Call State Management
**Objective**: Implement call session handling and state management

**Components**:
- [ ] **Call Manager** (`internal/core/calls/manager.go`)
  ```go
  type CallManager struct {
      tox          *tox.ToxManager
      toxav        *toxcore.ToxAV
      audioCapture AudioCapture
      audioPlayback AudioPlayback
      videoCapture VideoCapture
      videoRenderer VideoRenderer
      
      activeCalls  map[string]*CallSession
      mu          sync.RWMutex
  }
  
  func (cm *CallManager) InitiateCall(friendID uint32, audio, video bool) (*CallSession, error)
  func (cm *CallManager) AcceptCall(callID string, audio, video bool) error
  func (cm *CallManager) RejectCall(callID string) error
  func (cm *CallManager) EndCall(callID string) error
  ```

- [ ] **Call Session** (`internal/core/calls/session.go`)
  ```go
  type CallSession struct {
      id          string
      friendID    uint32
      state       CallState
      audioEnabled bool
      videoEnabled bool
      startTime   time.Time
      
      // Statistics
      audioStats  AudioStats
      videoStats  VideoStats
      networkStats NetworkStats
      
      mu sync.RWMutex
  }
  ```

- [ ] **Call Types** (`internal/core/calls/types.go`)
  ```go
  type CallState int
  const (
      CallStateRinging CallState = iota
      CallStateConnecting
      CallStateActive
      CallStateEnding
      CallStateEnded
  )
  
  type AudioFrame struct {
      PCM        []int16
      SampleRate int
      Channels   int
      Timestamp  time.Time
  }
  
  type VideoFrame struct {
      Y, U, V   []byte
      Width     int
      Height    int
      Timestamp time.Time
  }
  ```

**Estimated Time**: 2 weeks

#### 3.2 Database Schema Extensions
**Objective**: Add call history and statistics storage

**Database Changes**:
```sql
-- Call history table
CREATE TABLE calls (
    id TEXT PRIMARY KEY,
    friend_id INTEGER NOT NULL,
    direction INTEGER NOT NULL, -- 0=outgoing, 1=incoming
    call_type INTEGER NOT NULL, -- 0=audio, 1=video, 2=both
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    duration INTEGER, -- seconds
    status INTEGER NOT NULL, -- 0=completed, 1=missed, 2=rejected
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (friend_id) REFERENCES contacts(id)
);

-- Call statistics table
CREATE TABLE call_stats (
    call_id TEXT NOT NULL,
    metric_name TEXT NOT NULL,
    metric_value TEXT NOT NULL,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (call_id, metric_name, timestamp),
    FOREIGN KEY (call_id) REFERENCES calls(id)
);

-- Indexes for performance
CREATE INDEX idx_calls_friend_id ON calls(friend_id);
CREATE INDEX idx_calls_start_time ON calls(start_time);
CREATE INDEX idx_call_stats_call_id ON call_stats(call_id);
```

**Estimated Time**: 1 week

### Phase 4: UI Integration (opd-ai/whisp)

#### 4.1 Call Interface Components
**Objective**: Create user interface for voice and video calls

**Components**:
- [ ] **Call Window** (`ui/shared/call_window.go`)
  ```go
  type CallWindow struct {
      container    *container.VBox
      videoView    *widget.Card
      audioControls *container.HBox
      callInfo     *widget.Label
      
      muteButton   *widget.Button
      videoButton  *widget.Button
      hangupButton *widget.Button
      
      session *calls.CallSession
  }
  ```

- [ ] **Incoming Call Dialog** (`ui/shared/incoming_call.go`)
  ```go
  type IncomingCallDialog struct {
      dialog      dialog.Dialog
      friendName  string
      callType    string
      acceptFunc  func(audio, video bool)
      rejectFunc  func()
  }
  ```

- [ ] **Call History View** (`ui/shared/call_history.go`)
  ```go
  type CallHistoryView struct {
      list      *widget.List
      calls     []CallRecord
      onCallBack func(friendID uint32)
  }
  ```

**UI Libraries** (already established):
- Fyne for cross-platform UI
- Existing adaptive UI system
- Platform-specific call notifications

**Estimated Time**: 3 weeks

#### 4.2 Platform-Specific Integration
**Objective**: Integrate with platform calling features

**Platform Features**:
- [ ] **Desktop Integration**
  - System tray call notifications
  - Keyboard shortcuts (Space to mute, Cmd/Ctrl+D to hangup)
  - Screen sharing support
  - Multiple monitor detection

- [ ] **Mobile Integration**
  - CallKit integration (iOS) for native call experience
  - Telecom framework (Android) for system call interface
  - Background calling permissions
  - Proximity sensor handling

- [ ] **Accessibility**
  - Screen reader announcements for call state changes
  - High contrast mode support for call interface
  - Large button mode for easier touch targets
  - Voice command support for call controls

**Estimated Time**: 2 weeks per platform (6 weeks total)

### Phase 5: Quality and Performance Optimization

#### 5.1 Network Optimization
**Objective**: Optimize call quality and adapt to network conditions

**Features**:
- [ ] **Adaptive Bitrate**
  ```go
  type BitrateController struct {
      currentAudioBitrate int
      currentVideoBitrate int
      networkQuality     float64
      packetLoss         float64
  }
  
  func (bc *BitrateController) AdjustBitrates(stats NetworkStats)
  ```

- [ ] **Jitter Buffer**
  ```go
  type JitterBuffer struct {
      frames   []Frame
      maxDelay time.Duration
      minDelay time.Duration
  }
  
  func (jb *JitterBuffer) AddFrame(frame Frame)
  func (jb *JitterBuffer) GetFrame() Frame
  ```

- [ ] **Echo Cancellation**
  ```go
  // Library: github.com/gordonklaus/echo
  // License: MIT
  // Import: "github.com/gordonklaus/echo"
  // Why: Acoustic echo cancellation for better call quality
  
  type EchoCanceller interface {
      ProcessFrame(frame AudioFrame) AudioFrame
  }
  ```

**Estimated Time**: 2 weeks

#### 5.2 Testing and Quality Assurance
**Objective**: Comprehensive testing of call functionality

**Test Coverage**:
- [ ] **Unit Tests** (>90% coverage)
  - Call state transitions
  - Audio/video frame processing
  - Codec encode/decode operations
  - Network adaptation algorithms

- [ ] **Integration Tests**
  - End-to-end call scenarios
  - Multi-platform calling
  - Network condition simulation
  - Call quality metrics validation

- [ ] **Performance Tests**
  - Latency measurements
  - CPU/memory usage profiling
  - Battery usage optimization (mobile)
  - Concurrent call handling

- [ ] **Demo Applications**
  ```bash
  go run ./cmd/demo-calls     # Call system demonstration
  go run ./cmd/demo-audio     # Audio codec testing
  go run ./cmd/demo-video     # Video codec testing
  ```

**Estimated Time**: 2 weeks

## Implementation Timeline

### Total Estimated Time: 20-24 weeks

| Phase | Duration | Dependencies | Deliverable |
|-------|----------|--------------|-------------|
| **Phase 1: ToxAV Core** | 3-4 weeks | None | ToxAV bindings in toxcore |
| **Phase 2: Codecs** | 5 weeks | Phase 1 | Audio/video processing |
| **Phase 3: Call Management** | 3 weeks | Phase 2 | Call state system |
| **Phase 4: UI Integration** | 9 weeks | Phase 3 | Complete call interface |
| **Phase 5: Optimization** | 4 weeks | Phase 4 | Production-ready quality |

### Milestone Schedule

#### Month 1: ToxAV Foundation
- ✅ ToxAV C library integration
- ✅ Go wrapper API design
- ✅ Basic call initiation/termination

#### Month 2-3: Media Processing
- ✅ Audio capture/playback system
- ✅ Opus codec integration
- ✅ Video capture/rendering system
- ✅ VP8 codec integration

#### Month 4: Call Management
- ✅ Call session handling
- ✅ Database schema updates
- ✅ State management system

#### Month 5-6: UI Development
- ✅ Call interface components
- ✅ Platform-specific integration
- ✅ Accessibility features

#### Month 6: Quality & Launch
- ✅ Performance optimization
- ✅ Comprehensive testing
- ✅ Production deployment

## Technical Risks and Mitigation

### High Risk
1. **ToxAV C Library Complexity**
   - *Risk*: Difficult CGO integration with multiple dependencies
   - *Mitigation*: Start with minimal viable implementation, iterate

2. **Cross-Platform Codec Support**
   - *Risk*: Different codec availability across platforms
   - *Mitigation*: Use pure Go codecs where possible, provide fallbacks

3. **Real-Time Performance**
   - *Risk*: Audio/video latency and quality issues
   - *Mitigation*: Extensive performance testing, adaptive algorithms

### Medium Risk
1. **Mobile Platform Restrictions**
   - *Risk*: iOS/Android background calling limitations
   - *Mitigation*: Follow platform best practices, use native call APIs

2. **Network Adaptation**
   - *Risk*: Poor call quality on unstable networks
   - *Mitigation*: Implement comprehensive network monitoring and adaptation

### Low Risk
1. **UI Complexity**
   - *Risk*: Complex call interface implementation
   - *Mitigation*: Leverage existing Fyne UI system and adaptive components

## License Compliance

All suggested libraries maintain MIT or Apache 2.0 licenses compatible with the existing MIT license:

- **github.com/hraban/opus**: MIT License
- **github.com/gordonklaus/portaudio**: MIT License
- **github.com/pion/webrtc/v3**: MIT License
- **github.com/vladimirvivien/go4vl**: Apache 2.0 License
- **github.com/kbinani/screenshot**: MIT License

## Success Criteria

### Functional Requirements
- ✅ Successful voice calls with <150ms latency
- ✅ Video calls with synchronized audio/video streams
- ✅ Call history and statistics tracking
- ✅ Cross-platform compatibility (Windows, macOS, Linux, Android, iOS)
- ✅ Integration with existing Whisp contact system

### Quality Requirements
- ✅ >90% test coverage for call-related code
- ✅ <10% CPU usage for audio calls
- ✅ <25% CPU usage for video calls
- ✅ Graceful degradation on poor network conditions
- ✅ Accessibility compliance for call interface

### User Experience Requirements
- ✅ Intuitive call interface matching platform conventions
- ✅ Clear call quality indicators and statistics
- ✅ Seamless integration with existing chat interface
- ✅ Reliable call notifications and history

## Next Steps

1. **Review and Approval**: Technical review of this implementation plan
2. **Resource Allocation**: Assign development resources for toxcore and whisp repositories
3. **Phase 1 Kickoff**: Begin ToxAV C library integration in opd-ai/toxcore
4. **Parallel Development**: Set up development environments and testing infrastructure
5. **Regular Checkpoints**: Weekly progress reviews and technical checkpoint meetings

---

*This document serves as the master plan for implementing P2P voice and video calls in Whisp. It should be updated as implementation progresses and requirements evolve.*
