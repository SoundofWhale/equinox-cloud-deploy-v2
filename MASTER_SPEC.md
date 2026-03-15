# EQUINOX 2.0 — Master Technical Specification

> Structured developer reference derived from the EQUINOX MASTER SPECIFICATION V.2.0 (FINAL).

---

## 1. FOUNDATION & SECURITY (The Warden)

**Agent**: [equinox_warden.md](file:///c:/AI/agents/equinox_warden.md) | **Skills**: [equinox_warden_skills.md](file:///c:/AI/skills/equinox_warden_skills.md)

### Storage
| Component | Technology |
|-----------|-----------|
| Container | Docker (`golang:1.22-alpine` / `node:20-alpine`) |
| Network | Docker Bridge (Frontend -> `http://backend:8080`) |
| Database | SQLite + SQLCipher (AES-256) |
| Page size | 4096 bytes (`PRAGMA cipher_page_size`) |
| KDF | Argon2id (`kdf_iter = 256000`) |
| Media | `/vault/media/` — each file AES-256-GCM encrypted individually |

### Master Key Modes
| Mode | Storage | Use Case |
|------|---------|----------|
| **Paranoid** | RAM only (Argon2id derived) | Maximum security |
| **Convenience** | OS Keychain (`go-keyring`) | Daily use |

### Recovery Key
- **Type**: BIP-39 mnemonic (12 words, 128-bit entropy)
- **Display**: Once only at registration
- **Backup**: Encrypted blob → Relay server (relay sees only ciphertext)
- **Recovery**: Email → blob retrieval + 12 words → local decryption

### E2EE Blind Relay
| Scenario | Protocol |
|----------|----------|
| Personal | Local/self-hosted, free |
| Teams | Multi-sig threshold encryption, paid |

### Relay API
```
POST /api/v1/backup         — Upload encrypted blob
GET  /api/v1/backup/:id     — Retrieve by hash
POST /api/v1/sync/delta     — Push encrypted delta
GET  /api/v1/sync/delta     — Pull deltas since timestamp
```

---

## 2. VISUAL CANVAS (The Architect)

**Agent**: [equinox_architect.md](file:///c:/AI/agents/equinox_architect.md) | **Skills**: [equinox_architect_skills.md](file:///c:/AI/skills/equinox_architect_skills.md)

### Infinite Canvas
| Feature | Implementation |
|---------|---------------|
| Engine | PixiJS v8 |
| Pan | Pointer drag on `worldContainer` |
| Zoom | Mouse wheel scale |
| Performance | Object pooling, off-screen culling, 60fps target |

### Work Dimension (Cosmic)
| Element | Visual | Color |
|---------|--------|-------|
| Background | Deep space slate | `#0a0a1a` |
| Parent task | Planet (HUD Style) | `#90caf9` (Pastel Blue) |
| Subtask | Satellite orbiting | `#80cbc4` (Soft Teal) |
| Layout | HUD Center Window | Cinematic / Sci-fi |

### Personal Dimension (Organic Tree)
| Element | Visual | Color |
|---------|--------|-------|
| Background | Warm dark | `#1a1208` |
| Parent task | Branch (bezier) | `#66bb6a` (moss) |
| Subtask | Sub-branch | `#ffb300` (amber) |
| Layout | Organic, right-anchored | Warm tones |

### Cloud Mode (Without Schedule)
- Notes = translucent drifting rounded rectangles
- Random `vx`, `vy` drift; bounce off canvas edges
- No hierarchy, no grid

### Weight Dynamics
```
Planet radius  = min(24 + subtasks × 4,  72)  px
Branch thickness = min(2  + subtasks × 0.8, 12) px
```

### Smart Focus
- Click node → others fade to `alpha: 0.15` + `BlurFilter(8)`
- Click background → restore all to full alpha

### Completed Task Lifecycle
1. Completion → visual drift near parent (gentle float)
2. After 2–3 days → animate off-screen
3. Move to archive store; remove from canvas

---

## 3. LOGIC & CONFLICT ENGINE (The Arbiter)

**Agent**: [equinox_arbiter.md](file:///c:/AI/agents/equinox_arbiter.md) | **Skills**: [equinox_arbiter_skills.md](file:///c:/AI/skills/equinox_arbiter_skills.md)

### Task Detail Modularity
Tasks now support dynamic module composition. Users can add or remove modules per-task. These modules are rendered in a decoupled `TaskDetailPanel.svelte` component.

| Module | ID | Icon | Description |
|--------|----|------|-------------|
| **Description** | `description` | 📝 | Rich text body |
| **Checklist** | `checklist` | ✅ | Interactive subtasks |
| **Attachments** | `attachments` | 📎 | OCR-capable file uploads |
| **AI Advice** | `ai_advice` | 🧠 | Contextual recommendations |

### Templates (Default Modules)
Templates provide the initial set of modules, which can be further customized.
| Template | Modules | Use Case |
|----------|---------|----------|
| **Task** | Title + TimeBlock | Deadline-driven work |
| **Note** | Title + Text | Information capture |
| **Meeting** | Title + TimeBlock + Checklist | Scheduled with contacts |
| **Ritual** | Title + RitualCron | Cyclic habits / routines |

### Conflict Engine
| Slot Type | Behavior | Override? |
|-----------|----------|-----------|
| **Hard** | Blocks work task creation | ❌ Never |
| **Soft** | Warning shown | ✅ User confirms |

**Default Hard Slots (seeded at first boot)**:
| Name | Time | Recurs |
|------|------|--------|
| Sleep | 23:00–07:00 | Daily |
| Family Time | 18:00–20:00 | Daily |
| Rest | Sat 13:00–18:00 | Weekly |

### Emergency Mode (Orange)
| Phase | Action |
|-------|--------|
| Trigger | Hold button **5 seconds** |
| Active | 4-hour countdown timer starts |
| Expiry | `emergency_expired` event → Equi cough state |
| Mascot state | Ромб (Emergency) → Кашель |

### Snapshotting
| Parameter | Value |
|-----------|-------|
| Schedule | Daily midnight (cron) |
| Granularity | Per branch/planet node |
| Compression | gzip JSON |
| Retention | 30 snapshots per node |
| Restore | Safe upsert (no data loss) |

---

## 4. AI & UX LAYER (The Messenger)

**Agent**: [equinox_messenger.md](file:///c:/AI/agents/equinox_messenger.md) | **Skills**: [equinox_messenger_skills.md](file:///c:/AI/skills/equinox_messenger_skills.md)

### AI Configuration
| Parameter | Value |
|-----------|-------|
| Model | Gemma2 2b |
| Runtime | Ollama (`localhost:11434`) |
| Network | Local only — no external calls |

### AI Roles
| Role | Dimension | Persona |
|------|-----------|---------|
| **CTO** | Work | Direct, concise, action-oriented |
| **Zen Master** | Personal | Calm, mindful, nature metaphors |

### Anonymizer Pipeline
```
Raw user text → Regex anonymizer → Anonymized text → Ollama → Response
```
Patterns replaced: `[NAME]`, `[EMAIL]`, `[PHONE]`, `[AMOUNT]`, `[LOCATION]`

### OCR Pipeline
```
File upload → Tesseract (Multi-language: rus+eng+...) → AnonymizeText → Ollama summarize → task.Text
```

### Inter-Context Nudges
- Check every 25 min during Work session
- Look ahead 4 hours for personal tasks
- Default: **ON** (user-configurable)

### The Gateway (Dimension Switch)
| Element | Detail |
|---------|--------|
| Duration | 3 seconds |
| Animation | Breathing circle (scale 0.6→1.4→0.8) |
| Text | "Сделай глубокий вдох" + random quote |
| Audio | `equi_breath.mp3` |
| Mascot | Shown during transition |

### Equi Mascot States
| State | Russian | Trigger | Audio |
|-------|---------|---------|-------|
| Каска | Focus | Emergency active | — |
| Прозрачный | Flow | Idle, deep work | — |
| Молния | Warning | Soft conflict / nudge | `equi_warning.mp3` |
| Ромб | Emergency | Emergency mode | `equi_warning.mp3` |
| Пушистый | Rest | Rest slot active | `equi_breath.mp3` |
| Кашель | Cough | Emergency timer expired | `equi_warning.mp3` |

> **Mascot appears ONLY during**: Gateway transition OR critical alerts.

### Audio & Voice Layer
| Feature | Implementation |
|---------|----------------|
| **Sound FX** | `Web Audio API` (success, breath, warning) |
| **Voice STT** | `Web Speech API` (Lite mode) / `Whisper.cpp` (Total Privacy) |
| **Trigger** | QuickAction Mic button -> automatic task creation |

---

## 5. DEVELOPMENT STACK SUMMARY

| Layer | Technology | Version |
|-------|-----------|---------|
| Backend language | Go | 1.22+ |
| DB | SQLite + SQLCipher | Latest |
| Frontend framework | SvelteKit | 2.x |
| Canvas engine | PixiJS | v8 |
| AI runtime | Ollama | Latest |
| AI model | Gemma2 2b | q4_K_M |
| OCR | Tesseract | 5.x (Multi-lang) |
| Container | Docker | 24+ |
| Encryption | AES-256-GCM, Argon2id | — |
| Recovery | BIP-39 | 12 words |
| Key storage | go-keyring | Latest |

---

*This document is the structured technical reference for EQUINOX 2.0. For the narrative spec, see the original EQUINOX MASTER SPECIFICATION V.2.0.*
