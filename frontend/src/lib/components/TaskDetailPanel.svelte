<script lang="ts">
    import { onMount } from "svelte";
    import { fade, fly } from "svelte/transition";
    import { tasks, type Task, aggregateSubtasks } from "../stores/tasks";
    import { uploadFile } from "../utils/api";
    import { whisper } from "../stores/whisper";

    export let task: Task;
    export let onClose: () => void;

    // ── Navigation Stack (using IDs for reactivity) ─────────────────────
    let taskIdStack: string[] = [];
    let currentTaskId: string = task?.id || "";
    let slideDirection: 'left' | 'right' = 'left';
    let panelKey = 0; // Force re-render on navigation

    let isEditingTitle = false;
    let isEditingText = false;
    let editTitle = "";
    let editText = "";
    let editingSubtaskId: string | null = null;
    let editSubtaskTitle = "";
    let newSubtaskTitle = "";
    let fileInputRef: HTMLInputElement;
    let showModuleDropdown = false;

    // Reactively find the current task from the store and aggregate its subtasks
    $: currentTaskFromStore = $tasks.find(t => t.id === currentTaskId);
    $: baseTask = currentTaskFromStore || (currentTaskId === task.id ? task : null);
    $: currentTask = baseTask ? {
        ...baseTask,
        subtasks: aggregateSubtasks($tasks, baseTask)
    } : null;

    // Reset internal state when the root selection on canvas changes
    $: {
        if (task?.id) {
            const rootIdShowing = taskIdStack.length > 0 ? taskIdStack[0] : currentTaskId;
            if (rootIdShowing !== task.id) {
                currentTaskId = task.id;
                taskIdStack = [];
                panelKey++;
                resetEditState();
            }
        }
    }

    $: breadcrumbs = [...taskIdStack, currentTaskId]
        .map(id => $tasks.find(t => t.id === id) || (id === task?.id ? task : null))
        .filter((t): t is Task => !!t && !!t.title);

    let isNavigating = false;
    async function navigateToChild(subtaskTitle: string) {
        if (isNavigating) return;
        isNavigating = true;
        try {
            // Look for existing child task in store
            let childTask: Task | null = null;
            const allTasks = await tasks.getChildren(currentTaskId);
            childTask = allTasks.find((t: Task) => (t.title || '').toLowerCase() === subtaskTitle.toLowerCase()) || null;
            
            if (!currentTask) {
                isNavigating = false;
                return;
            }
            
            // Create a new child task ONLY if it doesn't exist
            if (!childTask) {
                childTask = await tasks.createChildTask(currentTask, subtaskTitle);
            }

            if (childTask) {
                taskIdStack = [...taskIdStack, currentTaskId];
                currentTaskId = childTask.id;
                slideDirection = 'left';
                panelKey++;
                resetEditState();
            }
        } finally {
            isNavigating = false;
        }
    }

    function navigateBack() {
        if (taskIdStack.length > 0) {
            const prevId = taskIdStack[taskIdStack.length - 1];
            taskIdStack = taskIdStack.slice(0, -1);
            currentTaskId = prevId;
            slideDirection = 'right';
            panelKey++;
            resetEditState();
        }
    }

    function navigateToBreadcrumb(index: number) {
        if (index === taskIdStack.length) return; // already on this task
        if (index === 0) {
            // Go back to root
            currentTaskId = task.id;
            taskIdStack = [];
        } else {
            currentTaskId = taskIdStack[index];
            taskIdStack = taskIdStack.slice(0, index);
        }
        slideDirection = 'right';
        panelKey++;
        resetEditState();
    }

    function resetEditState() {
        isEditingTitle = false;
        isEditingText = false;
        editingSubtaskId = null;
        newSubtaskTitle = "";
        editTitle = currentTask?.title || "";
        editText = currentTask?.text || "";
    }

    let newMeetingTitle = "";
    let newMeetingStart = "";
    let newMeetingEnd = "";

    onMount(() => {
        const handleClickOutside = () => {
            showModuleDropdown = false;
        };
        window.addEventListener("click", handleClickOutside);
        return () => window.removeEventListener("click", handleClickOutside);
    });

    $: {
        if (currentTask && currentTask.title) {
            if (!isEditingTitle) editTitle = currentTask.title;
            if (!isEditingText) editText = currentTask.text || "";
        }
    }

    // Contextual whisper handler for description
    $: if ($whisper.status === 'completed' && $whisper.targetTaskId === currentTaskId && $whisper.result) {
        console.log("📝 Whisper: [CONTEXTUAL] Appending to task:", currentTaskId);
        // Append to local state AND store
        editText = (editText ? editText + "\n\n" : "") + $whisper.result;
        tasks.updateTask(currentTaskId, { text: editText });
        whisper.reset();
    }

    function saveTitle() {
        if (!currentTask || !editTitle.trim()) {
            isEditingTitle = false;
            return;
        }
        tasks.updateTask(currentTask.id, { title: editTitle });
        isEditingTitle = false;
    }

    function saveText() {
        if (!currentTask) {
            isEditingText = false;
            return;
        }
        tasks.updateTask(currentTask.id, { text: editText });
        isEditingText = false;
    }

    function startEditSubtask(sub: any) {
        editingSubtaskId = sub.id;
        editSubtaskTitle = sub.title;
    }

    function saveSubtask(subtaskId: string) {
        if (!currentTask) return;
        if (editSubtaskTitle.trim()) {
            tasks.editSubtask(currentTask.id, subtaskId, editSubtaskTitle);
        }
        editingSubtaskId = null;
    }

    function addSubtask() {
        if (!currentTask) return;
        if (newSubtaskTitle.trim()) {
            tasks.addSubtask(currentTask.id, newSubtaskTitle);
            newSubtaskTitle = "";
        }
    }

    async function handleFileUpload(e: any) {
        if (!currentTask) return;
        const file = e.target.files?.[0];
        if (file) {
            try {
                await uploadFile(currentTask.id, file);
                tasks.addFile(currentTask.id, file.name);
            } catch (err) {
                console.error("Upload failed", err);
                alert("Ошибка при загрузке файла.");
            }
            e.target.value = "";
        }
    }

    function handleDeleteTask() {
        if (!currentTask) return;
        if (confirm("Удалить эту планету?")) {
            onClose();
            tasks.removeTask(currentTask.id);
        }
    }

    function addMeeting() {
        if (!currentTask) return;
        if (!newMeetingTitle.trim() || !newMeetingStart || !newMeetingEnd) {
            alert("Заполните все поля встречи");
            return;
        }

        const now = new Date();
        const start = new Date(now.toDateString() + " " + newMeetingStart);
        const end = new Date(now.toDateString() + " " + newMeetingEnd);

        const newMeeting = {
            id: Math.random().toString(36).substr(2, 9),
            title: newMeetingTitle,
            startTime: start.toISOString(),
            endTime: end.toISOString(),
        };

        const updatedMeetings = [...(currentTask.meetings || []), newMeeting];
        tasks.updateTask(currentTask.id, { meetings: updatedMeetings });

        newMeetingTitle = "";
        newMeetingStart = "";
        newMeetingEnd = "";
    }

    function removeMeeting(meetingId: string) {
        if (!currentTask) return;
        const updatedMeetings = currentTask.meetings.filter((m: any) => m.id !== meetingId);
        tasks.updateTask(currentTask.id, { meetings: updatedMeetings });
    }

    function toggleModule(moduleId: string) {
        if (!currentTask) return;
        let newModules = [...currentTask.modules];
        if (newModules.includes(moduleId)) {
            newModules = newModules.filter((m) => m !== moduleId);
        } else {
            newModules.push(moduleId);
        }
        tasks.updateTask(currentTask.id, { modules: newModules });
    }

    let draggedModuleId: string | null = null;
    let dropTargetModuleId: string | null = null;

    function handleDragStart(e: DragEvent, moduleId: string) {
        draggedModuleId = moduleId;
        if (e.dataTransfer) {
            e.dataTransfer.effectAllowed = "move";
        }
    }

    function handleDragOver(e: DragEvent, moduleId: string) {
        e.preventDefault();
        dropTargetModuleId = moduleId;
    }

    function handleDrop(e: DragEvent, targetModuleId: string) {
        e.preventDefault();
        if (!draggedModuleId || draggedModuleId === targetModuleId || !currentTask) {
            draggedModuleId = null;
            dropTargetModuleId = null;
            return;
        }

        const currentModules = [...currentTask.modules];
        const fromIndex = currentModules.indexOf(draggedModuleId);
        const toIndex = currentModules.indexOf(targetModuleId);

        if (fromIndex !== -1 && toIndex !== -1) {
            currentModules.splice(fromIndex, 1);
            currentModules.splice(toIndex, 0, draggedModuleId);
            tasks.updateTask(currentTask.id, { modules: currentModules });
        }

        draggedModuleId = null;
        dropTargetModuleId = null;
    }

    const availableModules = [
        { id: "description", label: "Описание", icon: "📝" },
        { id: "checklist", label: "Чек-лист", icon: "✅" },
        { id: "attachments", label: "Файлы", icon: "📎" },
        { id: "meetings", label: "Встречи", icon: "📅" },
        { id: "ai_advice", label: "Совет ИИ", icon: "🧠" },
    ];

    // Depth indicator
    $: currentDepth = taskIdStack.length;
    $: depthColor = currentDepth === 0 ? '#90caf9' : 
                    currentDepth === 1 ? '#80cbc4' :
                    currentDepth === 2 ? '#a5d6a7' :
                    currentDepth <= 5 ? '#fff176' :
                    '#ffab91';
</script>

<!-- Backdrop to darken the rest of the canvas slightly -->
<div
    class="task-modal-backdrop"
    transition:fade={{ duration: 200 }}
    on:click={onClose}
    on:keydown={(e) => e.key === "Escape" && onClose()}
    role="button"
    aria-label="Close task details"
    tabindex="-1"
></div>

<div class="task-detail-panel hud-window" transition:fade={{ duration: 300 }} style="--depth-color: {depthColor}">
    {#if !currentTask}
        <div class="loading-overlay">Загрузка...</div>
    {/if}
    <div class="hud-border-glow"></div>
    <div class="hud-scanlines"></div>

    <!-- Breadcrumbs -->
    {#if taskIdStack.length > 0}
        <nav class="breadcrumbs-bar" transition:fade={{ duration: 150 }}>
            {#each breadcrumbs as crumb, i}
                {#if i < breadcrumbs.length - 1}
                    <button
                        class="breadcrumb-item"
                        on:click={() => navigateToBreadcrumb(i)}
                    >
                        {#if i === 0}🪐{:else}📌{/if}
                        {(crumb?.title || '').length > 20 ? crumb.title.slice(0, 20) + '…' : (crumb?.title || 'Без названия')}
                    </button>
                    <span class="breadcrumb-sep">›</span>
                {:else}
                    <span class="breadcrumb-current">
                        📌 {(crumb?.title || '').length > 25 ? crumb.title.slice(0, 25) + '…' : (crumb?.title || 'Без названия')}
                    </span>
                {/if}
            {/each}
            <span class="depth-badge">LVL {currentDepth + 1}</span>
        </nav>
    {/if}

    {#if currentTask}
    <header class="panel-header hud-header">
        <div class="header-content">
            {#if taskIdStack.length > 0}
                <button class="back-btn" on:click={navigateBack} title="Назад">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M19 12H5M12 19l-7-7 7-7" />
                    </svg>
                    Назад
                </button>
            {/if}
            <span class="dimension-badge {currentTask.dimension}">
                {currentTask.dimension === "work"
                    ? "SYSTEM::WORK"
                    : "SYSTEM::PERSONAL"}
            </span>

            {#if isEditingTitle}
                <input
                    type="text"
                    class="edit-input-title hud-input"
                    bind:value={editTitle}
                    on:blur={saveTitle}
                    on:keydown={(e) => e.key === "Enter" && saveTitle()}
                    autofocus
                />
            {:else}
                <div
                    class="editable-element title-editable"
                    on:click={() => (isEditingTitle = true)}
                    on:keydown={(e) =>
                        e.key === "Enter" && (isEditingTitle = true)}
                    role="button"
                    tabindex="0"
                >
                    <h2 class="hud-glitch-text">
                        {currentTask.title} <span class="edit-icon">✎</span>
                    </h2>
                </div>
            {/if}
        </div>
        <div class="header-actions">
            <div class="module-manager">
                <button
                    class="hud-btn settings-btn"
                    title="MODULES"
                    on:click|stopPropagation={() =>
                        (showModuleDropdown = !showModuleDropdown)}
                >
                    ⚙️
                </button>
                {#if showModuleDropdown}
                    <div
                        class="module-dropdown"
                        on:click|stopPropagation
                        transition:fade={{ duration: 100 }}
                    >
                        {#each availableModules as mod}
                            <label class="module-toggle">
                                <input
                                    type="checkbox"
                                    checked={currentTask.modules.includes(mod.id)}
                                    on:change={() => toggleModule(mod.id)}
                                />
                                <span class="toggle-label"
                                    >{mod.icon} {mod.label}</span
                                >
                            </label>
                        {/each}
                    </div>
                {/if}
            </div>
            <button
                class="hud-btn delete-btn"
                on:click={handleDeleteTask}
                title="PURGE DATA"
            >
                🗑️
            </button>
            <button class="hud-btn close-btn" on:click={onClose}>✕</button>
        </div>
    </header>

    {#key panelKey}
    <div class="panel-body" in:fly={{ x: slideDirection === 'left' ? 300 : -300, duration: 250 }} out:fly={{ x: slideDirection === 'left' ? -300 : 300, duration: 200 }}>
        {#if currentTask.timeBlock}
            <section class="info-group">
                <h3>⏳ Время выполнения</h3>
                <div class="time-block">
                    {new Date(currentTask.timeBlock.start).toLocaleTimeString([], {
                        hour: "2-digit",
                        minute: "2-digit",
                    })}
                    — {new Date(currentTask.timeBlock.end).toLocaleTimeString([], {
                        hour: "2-digit",
                        minute: "2-digit",
                    })}
                </div>
            </section>
        {/if}

        {#each currentTask.modules as moduleId (moduleId)}
            <section
                class="info-group module-item"
                data-module={moduleId}
                class:is-dragging={draggedModuleId === moduleId}
                class:is-drop-target={dropTargetModuleId === moduleId}
                on:dragover={(e) => handleDragOver(e, moduleId)}
                on:drop={(e) => handleDrop(e, moduleId)}
                transition:fade
            >
                <div
                    class="module-drag-handle"
                    draggable="true"
                    on:dragstart={(e) => handleDragStart(e, moduleId)}
                    title="Move module"
                >
                    <svg
                        width="14"
                        height="14"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2.5"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                    >
                        <path d="M7 10l5-5 5 5M7 14l5 5 5-5" />
                        <path d="M12 5v14" />
                    </svg>
                </div>

                {#if moduleId === "description"}
                    <div class="module-header-row">
                        <h3>📝 Описание / Текст</h3>
                        <button 
                            class="voice-append-btn" 
                            class:recording={$whisper.status === 'recording'}
                            on:click={() => {
                                isEditingText = true;
                                if ($whisper.status === 'recording') {
                                    whisper.stopRecording();
                                } else {
                                    console.log("🎤 Whisper: Starting contextual recording for task:", currentTaskId);
                                    whisper.startRecording(currentTaskId);
                                }
                            }}
                        >
                            {$whisper.status === 'recording' ? '🔴' : '🎤'}
                        </button>
                    </div>
                    {#if isEditingText}
                        <textarea
                            class="edit-textarea"
                            bind:value={editText}
                            on:blur={saveText}
                            autofocus
                        ></textarea>
                    {:else}
                        <div
                            class="task-text editable-element"
                            on:click={() => {
                                isEditingText = true;
                                editText = currentTask?.text || "";
                            }}
                            on:keydown={(e) =>
                                e.key === "Enter" && (isEditingText = true)}
                            role="button"
                            tabindex="0"
                        >
                            {#if currentTask.text}
                                {currentTask.text}
                            {:else}
                                <span class="placeholder">Добавить описание...</span
                                >
                            {/if}
                        </div>
                    {/if}
                {:else if moduleId === "checklist"}
                    <h3>✅ Чек-лист</h3>
                    {#if currentTask.subtasks && currentTask.subtasks.length > 0}
                        <ul class="subtask-list">
                            {#each currentTask.subtasks as sub}
                                <li>
                                    <span
                                        class="check"
                                        on:click={() =>
                                            tasks.toggleSubtask(currentTask.id, sub.id)}
                                    >
                                        {sub.done ? "✓" : "○"}
                                    </span>
                                    {#if editingSubtaskId === sub.id}
                                        <input
                                            type="text"
                                            class="edit-subtask-input"
                                            bind:value={editSubtaskTitle}
                                            on:blur={() => saveSubtask(sub.id)}
                                            on:keydown={(e) =>
                                                e.key === "Enter" &&
                                                saveSubtask(sub.id)}
                                            autofocus
                                        />
                                    {:else}
                                        <span
                                            class="subtask-text editable-element"
                                            class:done={sub.done}
                                            on:click={() => navigateToChild(sub.title)}
                                            on:keydown={(e) =>
                                                e.key === "Enter" &&
                                                navigateToChild(sub.title)}
                                            role="button"
                                            tabindex="0"
                                        >
                                            {sub.title}
                                            <span class="edit-icon">
                                                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                                                    <path d="M9 18l6-6-6-6" />
                                                </svg>
                                            </span>
                                        </span>
                                        <button
                                            class="expand-subtask-btn"
                                            on:click={() => startEditSubtask(sub)}
                                            title="Переименовать"
                                        >
                                            <span class="edit-icon">✎</span>
                                        </button>
                                        <button
                                            class="remove-subtask-btn"
                                            on:click|stopPropagation={() =>
                                                tasks.removeSubtask(
                                                    currentTask.id,
                                                    sub.id,
                                                )}>✕</button
                                        >
                                    {/if}
                                </li>
                            {/each}
                        </ul>
                    {/if}
                    <div class="add-subtask-row">
                        <input
                            type="text"
                            class="add-subtask-input"
                            placeholder="Добавить пункт..."
                            bind:value={newSubtaskTitle}
                            on:keydown={(e) => e.key === "Enter" && addSubtask()}
                        />
                        <button class="add-subtask-btn" on:click={addSubtask}
                            >+</button
                        >
                    </div>
                {:else if moduleId === "meetings"}
                    <h3>📅 Связанные встречи</h3>
                    <div class="meetings-list">
                        {#if currentTask.meetings && currentTask.meetings.length > 0}
                            {#each currentTask.meetings as meeting}
                                <div class="meeting-item">
                                    <div class="meeting-info">
                                        <span class="meeting-title"
                                            >{meeting.title}</span
                                        >
                                        <span class="meeting-time">
                                            {new Date(
                                                meeting.startTime,
                                            ).toLocaleTimeString([], {
                                                hour: "2-digit",
                                                minute: "2-digit",
                                            })} -
                                            {new Date(
                                                meeting.endTime,
                                            ).toLocaleTimeString([], {
                                                hour: "2-digit",
                                                minute: "2-digit",
                                            })}
                                        </span>
                                    </div>
                                    <button
                                        class="delete-meeting-btn"
                                        on:click={() =>
                                            removeMeeting(meeting.id)}>✕</button
                                    >
                                </div>
                            {/each}
                        {:else}
                            <p class="empty-hint">Нет запланированных встреч</p>
                        {/if}
                    </div>

                    <div class="add-meeting-form">
                        <input
                            type="text"
                            placeholder="Название встречи..."
                            bind:value={newMeetingTitle}
                            class="hud-input"
                        />
                        <div class="time-inputs">
                            <input
                                type="time"
                                bind:value={newMeetingStart}
                                class="hud-input"
                            />
                            <span class="time-sep">по</span>
                            <input
                                type="time"
                                bind:value={newMeetingEnd}
                                class="hud-input"
                            />
                        </div>
                        <button
                            class="add-meeting-btn hud-action-btn"
                            on:click={addMeeting}>Запланировать</button
                        >
                    </div>
                {:else if moduleId === "attachments"}
                    <h3>📎 Прикрепленные файлы</h3>
                    {#if currentTask.files && currentTask.files.length > 0}
                        <ul class="files-list">
                            {#each currentTask.files as f}
                                <li>
                                    <span class="file-icon">📄</span>
                                    <a
                                        href="/api/v1/files/{currentTask.id}/{f}"
                                        target="_blank"
                                        class="file-link">{f}</a
                                    >
                                </li>
                            {/each}
                        </ul>
                    {/if}
                    <div
                        class="files-stub"
                        on:click={() => fileInputRef.click()}
                        on:keydown={(e) =>
                            e.key === "Enter" && fileInputRef.click()}
                        role="button"
                        tabindex="0"
                    >
                        <span class="file-icon">📎</span> Прикрепить файл
                    </div>
                    <input
                        type="file"
                        bind:this={fileInputRef}
                        style="display: none;"
                        on:change={handleFileUpload}
                    />
                {:else if moduleId === "ai_advice"}
                    <h3>🧠 Совет ИИ (CTO / Zen Master)</h3>
                    <div class="ai-box">
                        <p>
                            <em
                                >Совет генерируется на основе текста задачи.
                                Если вы подключили Mistral 7B, тут будет краткая
                                выжимка или рекомендация по выполнению.</em
                            >
                        </p>
                        <button
                            class="ai-btn"
                            on:click={() =>
                                alert(
                                    "Отличная задача! Я рекомендую разбить ее еще на пару мелких чек-листов для простоты.",
                                )}>Сгенерировать совет</button
                        >
                    </div>
                {/if}
            </section>
        {/each}
    </div>
    {/key}

    <footer class="panel-footer hud-footer">
        <button
            class="hud-action-btn complete-btn"
            on:click={() => {
                tasks.completeTask(currentTask.id);
                if (taskIdStack.length > 0) {
                    navigateBack();
                } else {
                    onClose();
                }
            }}>Завершить задачу</button
        >
    </footer>
    {/if}
</div>

<style>
    .task-modal-backdrop {
        position: fixed;
        inset: 0;
        background: rgba(0, 0, 0, 0.6);
        backdrop-filter: blur(4px);
        z-index: 999;
    }

    .hud-window {
        position: fixed;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        width: 80vw;
        height: 80vh;
        background: rgba(15, 23, 42, 0.98);
        border: 1px solid #90caf9;
        box-shadow:
            0 0 35px rgba(144, 202, 249, 0.2),
            inset 0 0 15px rgba(144, 202, 249, 0.05);
        border-radius: 8px;
        display: flex;
        flex-direction: column;
        z-index: 1000;
        overflow: hidden;
    }

    .hud-border-glow {
        position: absolute;
        inset: 0;
        pointer-events: none;
        border: 1px solid rgba(144, 202, 249, 0.2);
        box-shadow: inset 0 0 15px rgba(144, 202, 249, 0.05);
        border-radius: 8px;
    }

    .hud-scanlines {
        position: absolute;
        inset: 0;
        pointer-events: none;
        background: linear-gradient(
            rgba(15, 23, 42, 0) 50%,
            rgba(0, 0, 0, 0.05) 50%
        );
        background-size: 100% 4px;
        z-index: 1;
        opacity: 0.15;
    }

    .hud-header {
        padding: 1.5rem 2rem;
        background: rgba(144, 202, 249, 0.05);
        border-bottom: 1px solid rgba(144, 202, 249, 0.15);
        position: relative;
        z-index: 10;
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .header-content {
        display: flex;
        flex-direction: column;
    }

    .hud-glitch-text {
        color: #e2e8f0;
        text-shadow: 0 0 8px rgba(144, 202, 249, 0.3);
        font-family: "Inter", system-ui, sans-serif;
        letter-spacing: 1px;
        margin: 0;
        font-weight: 600;
    }

    .hud-input {
        background: rgba(144, 202, 249, 0.05) !important;
        border: 1px solid rgba(144, 202, 249, 0.3) !important;
        font-family: "Inter", system-ui, sans-serif !important;
        color: #e2e8f0 !important;
        width: 100%;
        padding: 0.5rem;
        border-radius: 6px;
        outline: none;
        font-weight: 600;
    }

    .header-actions {
        display: flex;
        gap: 0.5rem;
        align-items: center;
    }

    .hud-btn {
        background: rgba(144, 202, 249, 0.1);
        border: 1px solid rgba(144, 202, 249, 0.4);
        color: #90caf9;
        width: 40px;
        height: 40px;
        display: flex;
        align-items: center;
        justify-content: center;
        cursor: pointer;
        transition: all 0.2s ease;
        border-radius: 6px;
        font-family: "Inter", system-ui, sans-serif;
    }

    .hud-btn:hover {
        background: rgba(144, 202, 249, 0.2);
        color: #ffffff;
        box-shadow: 0 0 10px rgba(144, 202, 249, 0.2);
    }

    .module-manager {
        position: relative;
    }

    .module-dropdown {
        position: absolute;
        top: 100%;
        right: 0;
        background: #1a1a2e;
        border: 1px solid #90caf9;
        border-radius: 6px;
        padding: 0.5rem;
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
        z-index: 2000; /* Ensure it stays on top of everything inside and outside the panel */
        min-width: 180px;
        box-shadow:
            0 8px 32px rgba(0, 0, 0, 0.8),
            0 0 15px rgba(144, 202, 249, 0.2);
    }

    .module-toggle {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        color: #e2e8f0;
        cursor: pointer;
        padding: 0.3rem 0.5rem;
        border-radius: 4px;
        transition: background 0.2s;
        font-size: 0.9rem;
    }

    .module-toggle:hover {
        background: rgba(144, 202, 249, 0.1);
    }

    .hud-btn.delete-btn {
        border-color: rgba(239, 83, 80, 0.4);
        color: #ef5350;
    }

    .hud-btn.delete-btn:hover {
        background: rgba(239, 83, 80, 0.15);
        color: #ffcdd2;
        box-shadow: 0 0 10px rgba(239, 83, 80, 0.2);
    }

    .hud-footer {
        padding: 1.5rem 2rem;
        background: rgba(144, 202, 249, 0.02);
        border-top: 1px solid rgba(144, 202, 249, 0.1);
        display: flex;
        justify-content: center;
        z-index: 2;
    }

    .hud-action-btn {
        background: rgba(144, 202, 249, 0.05);
        border: 1px solid rgba(144, 202, 249, 0.4);
        color: #90caf9;
        padding: 0.8rem 2rem;
        font-family: "Inter", system-ui, sans-serif;
        font-weight: 600;
        letter-spacing: 2px;
        cursor: pointer;
        transition: all 0.2s;
        text-transform: uppercase;
        border-radius: 6px;
    }

    .hud-action-btn:hover {
        background: rgba(144, 202, 249, 0.15);
        box-shadow: 0 0 15px rgba(144, 202, 249, 0.15);
        color: #ffffff;
    }

    .panel-body {
        flex: 1;
        overflow-y: auto;
        padding: 2rem;
        position: relative;
        z-index: 2;
        display: grid;
        grid-template-columns: repeat(2, 1fr);
        gap: 1.5rem;
        align-content: start;
    }

    @media (max-width: 800px) {
        .panel-body {
            grid-template-columns: 1fr;
        }
    }

    .module-item {
        position: relative;
        transition: all 0.2s ease;
        border: 1px solid rgba(144, 202, 249, 0.05);
        padding: 1.5rem;
        padding-top: 2rem;
        background: rgba(15, 23, 42, 0.2);
        border-radius: 8px;
        display: flex;
        flex-direction: column;
    }

    /* Description and Checklist usually need more width */
    .module-item[data-module="description"],
    .module-item[data-module="checklist"] {
        grid-column: span 2;
    }

    @media (max-width: 800px) {
        .module-item[data-module="description"],
        .module-item[data-module="checklist"] {
            grid-column: span 1;
        }
    }

    .module-item.is-dragging {
        opacity: 0.4;
        background: rgba(144, 202, 249, 0.05);
        border: 1px dashed rgba(144, 202, 249, 0.5);
    }

    .module-item.is-drop-target {
        border-top: 2px solid #90caf9;
        padding-top: 3rem;
    }

    .module-drag-handle {
        position: absolute;
        top: 1rem;
        right: 1.5rem;
        width: 32px;
        height: 32px;
        display: flex;
        align-items: center;
        justify-content: center;
        cursor: grab;
        color: rgba(144, 202, 249, 0.4);
        background: rgba(144, 202, 249, 0.03);
        border: 1px solid rgba(144, 202, 249, 0.1);
        border-radius: 6px;
        transition: all 0.2s;
        z-index: 10;
    }

    .module-drag-handle:hover {
        color: #90caf9;
        background: rgba(144, 202, 249, 0.1);
        border-color: rgba(144, 202, 249, 0.3);
        box-shadow: 0 0 10px rgba(144, 202, 249, 0.1);
    }

    .module-drag-handle:active {
        cursor: grabbing;
        transform: scale(0.95);
    }
    .panel-body::-webkit-scrollbar {
        width: 6px;
    }
    .panel-body::-webkit-scrollbar-thumb {
        background: rgba(144, 202, 249, 0.3);
        border-radius: 3px;
    }
    .panel-body::-webkit-scrollbar-thumb:hover {
        background: rgba(144, 202, 249, 0.5);
    }

    .dimension-badge {
        font-family: "Inter", system-ui, sans-serif;
        font-size: 0.7rem;
        font-weight: 700;
        margin-bottom: 0.5rem;
        letter-spacing: 1px;
    }

    .dimension-badge.work {
        color: #90caf9;
    }
    .dimension-badge.personal {
        color: #81c784;
    }

    .module-header-row {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 0.5rem;
    }

    .voice-append-btn {
        background: rgba(144, 202, 249, 0.1);
        border: 1px solid rgba(144, 202, 249, 0.3);
        color: white;
        padding: 0.2rem 0.5rem;
        border-radius: 4px;
        cursor: pointer;
        font-size: 0.9rem;
        transition: all 0.2s;
    }

    .voice-append-btn:hover {
        background: rgba(144, 202, 249, 0.2);
    }

    .voice-append-btn.recording {
        border-color: #ff5252;
        color: #ff5252;
        animation: pulse-red-mini 1.5s infinite;
    }

    @keyframes pulse-red-mini {
        0% { transform: scale(1); }
        50% { transform: scale(1.1); }
        100% { transform: scale(1); }
    }

    .info-group h3 {
        font-size: 0.8rem;
        color: #94a3b8;
        text-transform: uppercase;
        letter-spacing: 1px;
        margin-bottom: 1rem;
        font-weight: 600;
    }

    .time-block {
        background: rgba(144, 202, 249, 0.05);
        border: 1px solid rgba(144, 202, 249, 0.2);
        padding: 0.75rem 1rem;
        border-radius: 6px;
        color: #e2e8f0;
        font-family: monospace;
    }

    .task-text {
        color: #cbd5e1;
        line-height: 1.6;
        min-height: 80px;
        padding: 1rem;
        background: rgba(15, 23, 42, 0.3);
        border-radius: 6px;
        border: 1px solid transparent;
        transition: all 0.2s;
        white-space: pre-wrap;
    }

    .editable-element:hover {
        border-color: rgba(144, 202, 249, 0.3);
        background: rgba(144, 202, 249, 0.02);
    }

    .edit-textarea {
        width: 100%;
        height: 150px;
        background: rgba(144, 202, 249, 0.05);
        border: 1px solid #90caf9;
        color: white;
        padding: 1rem;
        border-radius: 6px;
        outline: none;
        resize: vertical;
        font-family: "Inter", system-ui, sans-serif;
    }

    .subtask-list {
        list-style: none;
        padding: 0;
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }

    .subtask-list li {
        display: flex;
        align-items: stretch;
        background: rgba(15, 23, 42, 0.3);
        border-radius: 6px;
        transition: all 0.2s;
        overflow: hidden;
        border: 1px solid rgba(144, 202, 249, 0.05);
    }

    .check {
        cursor: pointer;
        color: #90caf9;
        font-family: monospace;
        font-size: 1.2rem;
        width: 44px;
        display: flex;
        align-items: center;
        justify-content: center;
        flex-shrink: 0;
        transition: transform 0.1s;
    }

    .check:hover {
        transform: scale(1.1);
        color: #ffffff;
    }

    .subtask-text {
        flex: 1;
        color: #e2e8f0;
        display: flex;
        align-items: center;
        gap: 0.75rem;
        padding: 0.75rem 1rem;
        cursor: pointer;
        border-left: 1px solid rgba(144, 202, 249, 0.05);
        border-right: 1px solid rgba(144, 202, 249, 0.05);
        transition: all 0.2s;
    }

    .subtask-text:hover {
        background: rgba(144, 202, 249, 0.08);
        color: #ffffff;
    }

    .subtask-text.done {
        text-decoration: line-through;
        opacity: 0.5;
    }

    .edit-subtask-input,
    .add-subtask-input {
        background: rgba(144, 202, 249, 0.05);
        border: 1px solid rgba(144, 202, 249, 0.2);
        color: white;
        padding: 0.4rem 0.8rem;
        border-radius: 4px;
        outline: none;
        flex: 1;
    }

    .add-subtask-row {
        margin-top: 1rem;
        display: flex;
        gap: 1rem;
    }

    .add-subtask-btn,
    .remove-subtask-btn {
        background: rgba(144, 202, 249, 0.1);
        border: 1px solid rgba(144, 202, 249, 0.3);
        color: #90caf9;
        cursor: pointer;
        border-radius: 4px;
        padding: 0.2rem 0.6rem;
        margin: 0.5rem;
    }

    .files-list {
        list-style: none;
        padding: 0;
        display: flex;
        flex-wrap: wrap;
        gap: 1rem;
    }

    .files-list li {
        background: rgba(144, 202, 249, 0.05);
        border: 1px solid rgba(144, 202, 249, 0.2);
        padding: 0.5rem 1rem;
        border-radius: 20px;
        display: flex;
        align-items: center;
        gap: 0.5rem;
    }

    .file-link {
        color: #90caf9;
        text-decoration: none;
        font-size: 0.9rem;
    }

    .files-stub {
        margin-top: 1rem;
        border: 1px dashed rgba(144, 202, 249, 0.3);
        padding: 1rem;
        text-align: center;
        border-radius: 6px;
        color: #94a3b8;
        cursor: pointer;
        transition: all 0.2s;
    }

    .files-stub:hover {
        background: rgba(144, 202, 249, 0.05);
        border-color: #90caf9;
        color: #e2e8f0;
    }

    .ai-box {
        background: linear-gradient(
            135deg,
            rgba(144, 202, 249, 0.1),
            rgba(129, 199, 132, 0.05)
        );
        border: 1px solid rgba(144, 202, 249, 0.2);
        padding: 1.5rem;
        border-radius: 8px;
    }

    .ai-box p {
        color: #e2e8f0;
        font-size: 0.95rem;
        margin-bottom: 1rem;
    }

    .ai-btn {
        background: #90caf9;
        color: #0f172a;
        border: none;
        padding: 0.6rem 1.2rem;
        border-radius: 4px;
        font-weight: 700;
        cursor: pointer;
        transition: transform 0.2s;
    }

    .ai-btn:hover {
        transform: translateY(-2px);
        background: #bbdefb;
    }
    .meetings-section {
        border-top: 1px solid rgba(144, 202, 249, 0.1);
        padding-top: 1rem;
    }

    .meeting-item {
        display: flex;
        justify-content: space-between;
        align-items: center;
        background: rgba(144, 202, 249, 0.05);
        padding: 0.75rem;
        border-radius: 6px;
        margin-bottom: 0.5rem;
        border: 1px solid rgba(144, 202, 249, 0.1);
    }

    .meeting-info {
        display: flex;
        flex-direction: column;
    }

    .meeting-title {
        font-weight: 600;
        color: #90caf9;
    }

    .meeting-time {
        font-size: 0.85rem;
        color: #94a3b8;
    }

    .delete-meeting-btn {
        background: none;
        border: none;
        color: #64748b;
        font-size: 1.2rem;
        cursor: pointer;
        padding: 0 0.5rem;
    }

    .delete-meeting-btn:hover {
        color: #f43f5e;
    }

    .add-meeting-form {
        margin-top: 1.5rem;
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
        padding: 1rem;
        background: rgba(144, 202, 249, 0.03);
        border-radius: 8px;
        border: 1px dashed rgba(144, 202, 249, 0.2);
    }

    .time-inputs {
        display: flex;
        align-items: center;
        gap: 0.5rem;
    }

    .time-sep {
        color: #94a3b8;
        font-size: 0.9rem;
    }

    .add-meeting-btn {
        width: 100%;
        margin-top: 0.5rem;
    }

    /* ── Breadcrumbs ───────────────────────────────────── */
    .breadcrumbs-bar {
        display: flex;
        align-items: center;
        gap: 0.3rem;
        padding: 0.6rem 2rem;
        background: rgba(144, 202, 249, 0.03);
        border-bottom: 1px solid rgba(144, 202, 249, 0.1);
        overflow-x: auto;
        z-index: 5;
        position: relative;
        flex-wrap: nowrap;
        scrollbar-width: none;
    }
    .breadcrumbs-bar::-webkit-scrollbar { display: none; }

    .breadcrumb-item {
        background: rgba(144, 202, 249, 0.08);
        border: 1px solid rgba(144, 202, 249, 0.2);
        color: #90caf9;
        padding: 0.25rem 0.6rem;
        border-radius: 12px;
        font-size: 0.75rem;
        cursor: pointer;
        transition: all 0.2s;
        white-space: nowrap;
        font-family: "Inter", system-ui, sans-serif;
    }
    .breadcrumb-item:hover {
        background: rgba(144, 202, 249, 0.15);
        color: #ffffff;
    }

    .breadcrumb-sep {
        color: rgba(144, 202, 249, 0.4);
        font-size: 1rem;
        user-select: none;
    }

    .breadcrumb-current {
        color: var(--depth-color, #90caf9);
        font-size: 0.8rem;
        font-weight: 600;
        white-space: nowrap;
        letter-spacing: 0.5px;
    }

    .depth-badge {
        margin-left: auto;
        background: rgba(144, 202, 249, 0.1);
        border: 1px solid var(--depth-color, #90caf9);
        color: var(--depth-color, #90caf9);
        padding: 0.15rem 0.5rem;
        border-radius: 8px;
        font-size: 0.65rem;
        font-weight: 700;
        letter-spacing: 1px;
        font-family: monospace;
        white-space: nowrap;
    }

    /* ── Back Button ───────────────────────────────────── */
    .back-btn {
        display: flex;
        align-items: center;
        gap: 0.3rem;
        background: rgba(144, 202, 249, 0.08);
        border: 1px solid rgba(144, 202, 249, 0.25);
        color: #90caf9;
        padding: 0.3rem 0.7rem;
        border-radius: 6px;
        font-size: 0.75rem;
        font-weight: 600;
        cursor: pointer;
        transition: all 0.2s;
        margin-bottom: 0.4rem;
        font-family: "Inter", system-ui, sans-serif;
    }
    .back-btn:hover {
        background: rgba(144, 202, 249, 0.15);
        color: #ffffff;
        box-shadow: 0 0 8px rgba(144, 202, 249, 0.15);
    }

    /* ── Expand Subtask Button (→ chevron) ─────────────── */
    .expand-subtask-btn {
        background: rgba(144, 202, 249, 0.08);
        border: 1px solid rgba(144, 202, 249, 0.2);
        color: #90caf9;
        width: 32px;
        height: 32px;
        display: flex;
        align-items: center;
        justify-content: center;
        cursor: pointer;
        border-radius: 6px;
        transition: all 0.2s;
        flex-shrink: 0;
        margin: 0.5rem 0 0.5rem 0.5rem;
    }
    .expand-subtask-btn:hover {
        background: rgba(144, 202, 249, 0.2);
        color: #ffffff;
        box-shadow: 0 0 8px rgba(144, 202, 249, 0.2);
        transform: translateX(2px);
    }

    /* Panel body overflow for slide animations */
    .hud-window {
        overflow: hidden;
    }

    /* Depth-colored border glow */
    .hud-border-glow {
        border-color: var(--depth-color, rgba(144, 202, 249, 0.2));
    }

    .description-footer {
        display: flex;
        justify-content: flex-end;
        padding-top: 0.5rem;
    }
</style>
