---
description: Launch the SvelteKit + PixiJS dev server and verify the Infinite Canvas with Work/Personal dimension switching (Architect agent workflow)
---

# Architect: Canvas Dev Server Workflow

## Prerequisites
- Node.js 20+ installed
- `npm` available in PATH

## Steps

1. **Install frontend dependencies**
```powershell
cd c:\AI\EQUINOX\frontend
npm install
```

2. **Start the SvelteKit dev server**
```powershell
npm run dev
```
Expected: Server at `http://localhost:5173`

3. **Open in browser and verify**
- Navigate to `http://localhost:5173/work` → Cosmic dimension (dark blue, planets)
- Navigate to `http://localhost:5173/personal` → Organic tree (warm dark, branches)
- Switch dimensions → Gateway 3s breathing animation triggers

4. **Test infinite canvas pan/zoom**
- Drag canvas → viewport pans smoothly
- Scroll wheel → zooms in/out on `worldContainer`

5. **Test Smart Focus**
- Click a planet/branch → others blur + fade
- Click canvas background → all restore

6. **Type-check all Svelte files**
```powershell
npm run check
```

7. **Build for production**
```powershell
npm run build
```

## Read Before Starting
**Agent**: [equinox_architect.md](file:///c:/AI/agents/equinox_architect.md)  
**Skills**: [equinox_architect_skills.md](file:///c:/AI/skills/equinox_architect_skills.md)
