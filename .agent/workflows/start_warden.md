---
description: Set up Docker container with SQLCipher encrypted database and test Recovery Key generation (Warden agent workflow)
---

# Warden: Foundation Setup Workflow

## Prerequisites
- Docker Desktop installed and running
- Go 1.22+ installed

## Steps

1. **Verify Docker is running**
```powershell
docker info
```

2. **Build the Warden container**
```powershell
cd c:\AI\EQUINOX
docker build -f docker/Dockerfile -t equinox-backend ./backend
```

3. **Start the full stack**
```powershell
docker-compose -f docker/docker-compose.yml up -d
```

4. **Verify backend health**
```powershell
curl http://localhost:8080/health
```
Expected: `{"status":"ok","version":"2.0.0"}`

5. **Verify SQLCipher is active** (run inside container)
```powershell
docker exec -it equinox-backend ./equinox-server --check-cipher
```

6. **Test Recovery Key generation**
```powershell
cd c:\AI\EQUINOX\backend
go run ./cmd/server -generate-recovery-key
```
Expected: 12-word mnemonic printed ONCE to terminal.

7. **Run security tests**
```powershell
cd c:\AI\EQUINOX\backend
go test ./internal/security/... -v
```

## Read Before Starting
**Agent**: [equinox_warden.md](file:///c:/AI/agents/equinox_warden.md)  
**Skills**: [equinox_warden_skills.md](file:///c:/AI/skills/equinox_warden_skills.md)
