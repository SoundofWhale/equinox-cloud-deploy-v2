# Task Checklist: Equinox 2.0 Security & Cloud Prep

## Permanent Agent Rules
- [IMPORTANT] **Do not perform browser tests (browser_subagent) unless explicitly requested by the USER.**

- [x] **Phase 1: Authentication Engine**
    - [x] Backend: JWT, Registration, Login.
    - [x] Frontend: Auth Store, Login/Register UI.
    - [x] Security: BCrypt password hashing.
    - [x] Build RAM-only Zero-Persistence pipeline.
    - [x] Docker isolation and resource limits.
    - [x] FFmpeg automatic transcoding (Zero-Persistence Fix).
- [x] **Project Audit & Optimization**
    - [x] Full codebase analysis for logical gaps.
    - [x] Performance bottleneck identification.
    - [x] DB: Add `user_id` to all relevant tables.
    - [x] Backend: context-aware `userID` injection.
    - [x] Service Level: Filter all task operations by `userID`.
    - [x] Verification: User A cannot see User B's data.

- [x] **Phase 3: Bug Fixes & Stability**
    - [x] Fix empty dashboard on task creation (auth headers).
    - [x] Resolve deep subtask duplication (ID mapping & store deduplication).
    - [x] Fix TaskDetailPanel reactivity leakage between different root tasks.
    - [x] Add navigation guards to prevent race conditions during subtask transitions.

- [x] **Phase 4: UX Polish**
    - [x] Swap "Rename" and "Go Deeper" actions for subtasks.
    - [x] Add visual indicators (arrow icon) for navigatable subtasks.
    - [x] Expand subtask click area and add hover effects for better HUD feel.

- [x] **Phase 5: Full Hydration & Data Sync**
    - [x] Connect `tasks.ts` to Go Backend API.
    - [x] Implement JWT-authorized task CRUD.
    - [x] Add "Welcome" task auto-creation for new users.

- [ ] **Phase 6: Whisper.cpp Integration (Privacy-First)**
    - [x] **Backend: Task Queue & Worker Pool**
        - [x] Implement `WhisperWorker` service with buffered channel.
        - [x] Add `POST /api/v1/ai/whisper` endpoint (Async).
        - [x] Add `GET /api/v1/ai/whisper/status/:id` for polling.
    - [x] **Security: E2EE & RAM-only Pipeline**
        - [x] Implement `StreamDecryptReader` for on-the-fly AES-256-GCM decryption.
        - [x] Build RAM-pipe to Whisper process (stdin).
        - [x] Implement manual buffer zeroing post-processing.
    - [x] **Infrastructure: Docker & Resources**
        - [x] Create `whisper-engine` Dockerfile (lean alpine).
        - [x] Configure `docker-compose.yml` with strict CPU/RAM limits.
    - [x] **Frontend: Voice HUD Integration**
        - [x] Implement `WhisperStore` with recording & client-side encryption.
        - [x] Add microphone HUD button to QuickAction bar (Global).
        - [x] Add microphone button to `Description` module (Contextual).
        - [x] Add real-time status feedback (pulsing/glowing).
    - [x] **Polishing & UX (Current)**
        - [x] Title logic (first 2 words).
        - [x] Context-aware recording (Create vs Append).
        - [x] Description hydration & Appending logic.

- [ ] **Phase 7: Oracle Cloud & Scalability**
    - [ ] Set up Oracle compute instance.
    - [ ] Deploy multi-container Docker setup.

## Session Summary (Handover)
Baseline hydrated project is stable. Documentation synchronized.
**Tomorrow's Focus**: Implement **Phase 6 (Whisper)** using the **Privacy & Stability Plan**:
- Worker Queue for requests.
- Argon2id + AES-256-GCM (client-side) encryption.
- Zero-persistence RAM-only processing.
- Whisper.cpp (small Q5_0) in an isolated container.
