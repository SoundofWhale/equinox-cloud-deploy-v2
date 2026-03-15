---
description: Run conflict engine tests, seed default hard slots, and test Emergency Mode timer (Arbiter agent workflow)
---

# Arbiter: Logic Engine Workflow

## Prerequisites
- Go 1.22+ installed
- Backend dependencies: `go mod download` completed

## Steps

1. **Download Go dependencies**
```powershell
cd c:\AI\EQUINOX\backend
go mod download
```

2. **Seed default hard conflict slots**
```powershell
go run ./cmd/seed_slots
```
Expected: Slots created — Sleep (23:00–07:00), Family Time (18:00–20:00), Rest (Sat 13:00–18:00)

3. **Run conflict engine tests**
```powershell
go test ./internal/arbiter/... -v -run TestConflict
```

4. **Run snapshot tests**
```powershell
go test ./internal/snapshot/... -v
```

5. **Run all Arbiter tests**
```powershell
go test ./internal/arbiter/... ./internal/snapshot/... -v
```

6. **Test Emergency Mode activation** (manual)
- Start the backend: `go run ./cmd/server`
- Call endpoint: `curl -X POST http://localhost:8080/api/v1/emergency/activate`
- Check status: `curl http://localhost:8080/api/v1/emergency/status`
- Expected: Session with `ends_at = now + 4h`

7. **Test snapshot creation** (manual)
```powershell
curl -X POST http://localhost:8080/api/v1/snapshots/trigger
```

## Read Before Starting
**Agent**: [equinox_arbiter.md](file:///c:/AI/agents/equinox_arbiter.md)  
**Skills**: [equinox_arbiter_skills.md](file:///c:/AI/skills/equinox_arbiter_skills.md)
