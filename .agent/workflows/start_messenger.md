---
description: Pull Mistral 7B via Ollama, verify OCR with Tesseract, test the anonymizer pipeline, and verify audio assets (Messenger agent workflow)
---

# Messenger: AI & UX Setup Workflow

## Prerequisites
- Ollama installed: https://ollama.com
- Tesseract installed: `winget install UB-Mannheim.TesseractOCR`
- Go 1.22+ and Node.js 20+ installed

## Steps

1. **Pull Mistral 7B model**
```powershell
ollama pull mistral:7b-instruct
```
Note: ~4GB download. Wait for completion.

2. **Verify Ollama is running**
```powershell
ollama list
curl http://localhost:11434/api/generate -d '{"model":"mistral:7b-instruct","prompt":"Hello","stream":false}'
```
Expected: JSON response with `"response"` field

3. **Verify Tesseract installation**
```powershell
tesseract --version
tesseract --list-langs
```
Expected: `rus` and `eng` in language list

4. **Run anonymizer unit tests**
```powershell
cd c:\AI\EQUINOX\backend
go test ./internal/ai/... -v -run TestAnonymize
```

5. **Test OCR pipeline** (manual)
```powershell
go test ./internal/ocr/... -v
```

6. **Verify audio assets exist**
```powershell
ls c:\AI\EQUINOX\frontend\static\assets\audio\
```
Expected: `equi_success.mp3`, `equi_breath.mp3`, `equi_warning.mp3`

> If audio files are missing, generate or source .mp3 files and place them in the above directory.

7. **Test AI endpoint** (backend must be running)
```powershell
curl -X POST http://localhost:8080/api/v1/ai/query `
  -H "Content-Type: application/json" `
  -d '{"role":"cto","prompt":"What should I prioritize today?"}'
```

8. **Test inter-context nudge endpoint**
```powershell
curl http://localhost:8080/api/v1/ai/nudge
```

## Read Before Starting
**Agent**: [equinox_messenger.md](file:///c:/AI/agents/equinox_messenger.md)  
**Skills**: [equinox_messenger_skills.md](file:///c:/AI/skills/equinox_messenger_skills.md)
