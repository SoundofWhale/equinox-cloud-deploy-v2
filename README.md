# EQUINOX 2.0 — Master Orchestrator

> Zero-knowledge visual task manager. Cosmic Work dimension. Organic Personal dimension. On-device AI. No compromises on privacy.

---

## Agent Network

| Agent | Role | Skills | File |
|-------|------|--------|------|
| 🛡️ **Warden** | Security, Docker, Encryption, Recovery | [equinox_warden_skills.md](file:///c:/AI/skills/equinox_warden_skills.md) | [equinox_warden.md](file:///c:/AI/agents/equinox_warden.md) |
| 🎨 **Architect** | Infinite Canvas, PixiJS, Dimensions, Animations | [equinox_architect_skills.md](file:///c:/AI/skills/equinox_architect_skills.md) | [equinox_architect.md](file:///c:/AI/agents/equinox_architect.md) |
| ⚖️ **Arbiter** | Conflict Engine, Emergency Mode, Snapshotting | [equinox_arbiter_skills.md](file:///c:/AI/skills/equinox_arbiter_skills.md) | [equinox_arbiter.md](file:///c:/AI/agents/equinox_arbiter.md) |
| 🤖 **Messenger** | AI (Mistral/Ollama), OCR, Mascot Equi, Audio | [equinox_messenger_skills.md](file:///c:/AI/skills/equinox_messenger_skills.md) | [equinox_messenger.md](file:///c:/AI/agents/equinox_messenger.md) |

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.22+ |
| Database | SQLite + SQLCipher (AES-256) |
| Frontend | SvelteKit + PixiJS |
| AI | Mistral 7B via Ollama (local) |
| OCR | Tesseract |
| Container | Docker |
| Encryption | AES-256-GCM, Argon2id, BIP-39 |

---

## Project Structure

```
EQUINOX/
├── backend/                    # Go backend
│   ├── cmd/
│   │   ├── server/             # main.go — HTTP server entrypoint
│   │   ├── seed_slots/         # CLI: seed default conflict slots
│   │   └── check_cipher/      # CLI: verify SQLCipher setup
│   └── internal/
│       ├── models/             # Task, Slot, Snapshot structs
│       ├── security/           # Encryption, key management (Warden)
│       ├── relay/              # E2EE sync relay client (Warden)
│       ├── arbiter/            # Conflict engine, emergency mode (Arbiter)
│       ├── snapshot/           # Snapshotting & restore (Arbiter)
│       ├── ai/                 # Ollama client, anonymizer (Messenger)
│       └── ocr/                # Tesseract wrapper (Messenger)
├── frontend/                   # SvelteKit app
│   ├── src/
│   │   ├── routes/
│   │   │   ├── work/           # Work Dimension (cosmic)
│   │   │   └── personal/       # Personal Dimension (organic tree)
│   │   └── lib/
│   │       ├── canvas/         # PixiJS renderers (Architect)
│   │       ├── components/     # Gateway, Mascot, EmergencyButton
│   │       ├── stores/         # tasks, dimension, mascot stores
│   │       └── utils/          # audio, smartFocus, weightCalc
│   └── static/assets/audio/   # equi_success.mp3, equi_breath.mp3, equi_warning.mp3
├── docker/
│   ├── Dockerfile              # Go backend container
│   └── docker-compose.yml      # Full stack orchestration
├── vault/
│   └── media/                  # Encrypted media files (AES-256-GCM)
└── .agent/
    └── workflows/             # Agent-specific dev workflows
```

---

## Development Phases

### Phase 1 — Warden (Foundation)
```powershell
# Run workflow
# See .agent/workflows/start_warden.md
```
Deliverable: Encrypted DB running in Docker, Recovery Key generation, Relay API stub.

### Phase 2 — Architect (Canvas)
```powershell
# See .agent/workflows/start_architect.md
```
Deliverable: Infinite Canvas with Work/Personal switching, Smart Focus, weight dynamics.

### Phase 3 — Arbiter (Logic)
```powershell
# See .agent/workflows/start_arbiter.md
```
Deliverable: Full task CRUD + conflict engine + emergency mode + snapshotting.

### Phase 4 — Messenger (Intelligence)
```powershell
# See .agent/workflows/start_messenger.md
```
Deliverable: Mistral AI query + OCR pipeline + Gateway animation + Equi mascot.

---

## Full Specification

See [MASTER_SPEC.md](file:///c:/AI/EQUINOX/MASTER_SPEC.md) for the complete technical reference.
