# Equinox 2.0 - Global Context Map 🗺️
*(Use this document to initialize context in new development sessions).*

## 1. Project Overview & Architecture
Equinox 2.0 is a next-generation task manager utilizing an "Infinite Canvas" UI instead of traditional lists.
*   **Frontend**: SvelteKit + PixiJS (v8). Svelte manages the DOM UI overlays (Modals, Action Buttons), while PixiJS controls the entire interactive 3D/2D backdrop.
*   **Backend**: Go (v1.22+). Uses SQLite (wrapped with SQLCipher for E2EE).
*   **AI/Local Processing**: Gemma2 2b via Ollama (for task advice/breakdown), Tesseract (for multi-language OCR on uploaded images).
*   **Deployment**: Docker-based with two modes: `build-full.sh` (includes Ollama) and `build-lite.sh` (lightweight, no local LLM).

## 2. Critical UI/UX Decisions (Do Not Revert)
*   **PixiJS vs DOM Integration**: The canvas (`PixiCanvas.svelte`) is the root. We do **not** render HTML elements inside the canvas natively. All UI (Side Panels, FABs, Emergency Overlays) are built in Svelte and absolutely positioned `z-index` overlays on top of the `<canvas>` element.
*   **Tasks Representation**:
    *   `Work` dimension = Planets with orbiting satellites (subtasks).
    *   `Personal` dimension = Growing trees with branches (tasks) and leaves.
*   **Smart Focus (`smartFocus.ts`)**: When a task is clicked, the Pixi camera zooms in and centers entirely on the *local coordinates* of the clicked object. Background objects receive a blur filter.
*   **Modular Task Details**: The task detail panel is composed of dynamic modules (Description, Checklist, Files, AI Advice). Each task can have a different set of active modules, allowing for flexibility and future expansion. Editing is via two-way binding with auto-save.

## 3. Current State (End of Phase 5 Baseline)
*   **UI/Visuals:** 100% complete for MVP. Focus works, dimensions switch, emergency mode triggers a visual red overlay.
*   **Frontend Stores (`tasks.ts`, `dimension.ts`)**: Fully hydrated. Connected to Go Backend via JWT-authorized `fetch()` calls. Real-time CRUD is operational.

## 4. Immediate Next Steps (Phase 6)
*See `implementation_plan.md` and `task.md` for granular breakdown.*
1.  **Whisper.cpp Integration**: Implement the "Privacy & Stability" queue-based transcription engine.
2.  **Voice UI**: Connect the microphone buttons to the new backend STT worker.
3.  **Local LLM Refinement**: Fine-tune prompts for the `gemma2:2b` model.

## 5. Known Issues / Gotchas to Watch Out For
*   **Pointer Events Conflict**: PixiJS and DOM modals can fight for `pointerdown`/`pointerup`. Ensure modals have `.stopPropagation()` so clicks don't bleed through to the canvas and trigger dragging unexpectedly.
*   **Audio Assets**: Local `.mp3` files (success, warning) currently face `Range Not Satisfiable` HTTP errors in SvelteKit dev mode. This needs resolving in production build or via Go static file serving.
правило для агента и тебя. перед тем как что то менять в коде или в структуре проекта нужно ознакомится с документацией проекта в папке docs. 
второе если ты не уверен в своих действиях лучше спроси меня. 
третье если ты хочешь провести тестирование проекта в браузере, то лучше спроси меня. что тебе надо знать из тестов. 