# Implementation Plan: Equinox 2.0 Evolution

## Current State
- ✅ Secure Multi-User Auth (JWT + BCrypt).
- ✅ Full Data Isolation (UserID-based DB filtering).
- ✅ Multi-container Docker environment (Go 1.23 + Node).

## Proposed Changes: Phase 3-4

### [Infra] Oracle Cloud Deployment
- **Goal**: Move from local dev to a persistent remote cloud node.
- **Components**:
    - Compute Instance (OCI x86_64 or ARM).
    - Persisted Volume for `vault.db`.
    - Docker Hub or Manual transfer for images.

### Phase 6: Whisper.cpp Integration (Privacy & Stability)

#### 1. Task Queue (Stability Engine)
- **Goal**: Prevent server lockup during heavy transcription.
- **Implementation**:
    - **Worker Pool**: A Go service managing a `chan *WhisperTask`.
    - **Concurrency**: Default to `1` worker to prevent CPU/GPU contention.
    - **Persistence**: Task state held in RAM; lost on restart (intended for privacy).

#### 2. Zero-Persistence & E2EE Pipeline (Privacy Engine)
- **Client-Side Encryption**:
    - Frontend encrypts `.wav` blob via `AES-256-GCM` using the user's Master Key (derived from passphrase).
- **Streaming Decryption**:
    - Backend receives encrypted stream.
    - `SecurityService` provides an `io.Reader` that decrypts on-the-fly (`StreamDecryptReader`).
- **RAM-only I/O**:
    - Resulting plaintext is piped directly to Whisper container's stdin via HTTP chunked upload or Unix Socket.
    - No temp files. No disk I/O for audio.
- **Memory Zeroing**: 
    - Critical buffers are overwritten with zeros immediately after the task completes.

#### 3. Docker Isolation (The Bunker)
- **Container**: `whisper-engine` (based on `ghcr.io/ggerganov/whisper.cpp:main`).
- **Limits**: `cpus: 1.5`, `mem_limit: 512mb` (for `small` model).
- **Networking**: Internal-only bridge; no internet access for Whisper container.

#### 4. Frontend Integration
- **Store**: `whisper.ts` handles `MediaRecorder`, encryption, and polling logic.
- **UI Components**:
    - **Global HUD**: A sci-fi microphone button in the QuickAction area for instant task creation.
    - **Module Context**: A subtle microphone icon in the bottom-right of the **Description/Text** module for appending voice notes to existing tasks.
    - **Visuals**: Pulsing glow during transcription, "breathing" cyan during recording.


## Verification Plan

### Stage 1: Documentation & Sync Verification
- [ ] Verify `global_context_map.md` matches current `tasks.ts` store logic.
- [ ] Verify `MASTER_SPEC.md` reflects current `gemma2:2b` model.

### Stage 2:- Whisper.cpp (small Q5_0) in an isolated container.
 with resource limits.
- [ ] Verify RAM-only processing using a test stream.
- [ ] Test the Go worker queue by submitting multiple files simultaneously.
