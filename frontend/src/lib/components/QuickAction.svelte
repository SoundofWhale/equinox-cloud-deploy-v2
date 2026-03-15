<script lang="ts">
    import { tasks } from "$lib/stores/tasks";
    import { activeDimension } from "$lib/stores/dimension";
    import { uploadForOCR, uploadFile } from "$lib/utils/api";
    import { fade, slide } from "svelte/transition";
    import { container } from "$lib/services";

    import { onMount, tick } from "svelte";

    import { whisper } from "$lib/stores/whisper";
    
    let isExpanded = false;
    let taskTitle = "";
    let isUploading = false;
    let inputEl: HTMLInputElement;

    // Watch for whisper completion to create task
    $: if ($whisper.status === 'completed' && $whisper.result && $whisper.targetTaskId === null) {
        console.log("🌟 Whisper: Creating NEW Task from Global context");
        handleWhisperResult($whisper.result);
        whisper.reset();
    }

    function handleWhisperResult(text: string) {
        const layout = container.getService<any>('Layout');
        const pos = layout.findEmptySpot($tasks);
        
        // Smart title: Strictly First 2 words or 1 if only one exists
        const words = text.trim().split(/\s+/).filter(w => w.length > 0);
        const title = words.length >= 2 ? words.slice(0, 2).join(" ") : (words[0] || "Новая планета");

        const newTask = {
            title: title + ".",
            text: text, // Full text in description
            dimension: $activeDimension === "cloud" ? "work" : ($activeDimension as "work" | "personal"),
            x: pos.x,
            y: pos.y
        };
        tasks.addTask(newTask as any);
    }

    async function toggleExpand() {
        isExpanded = !isExpanded;
        if (isExpanded) {
            await tick();
            inputEl?.focus();
        }
    }

    function toggleRecording() {
        if ($whisper.status === 'recording') {
            whisper.stopRecording();
        } else {
            whisper.startRecording();
        }
    }

    async function handleFileUpload(event: Event) {
        const input = event.target as HTMLInputElement;
        if (!input.files?.length) return;

        isUploading = true;
        try {
            const taskId = Math.random().toString(36).substr(2, 9);
            // Parallel upload
            const [ocrResult] = await Promise.all([
                uploadForOCR(input.files[0]),
                uploadFile(taskId, input.files[0])
            ]);

            const layout = container.getService<any>('Layout');
            const pos = layout.findEmptySpot($tasks);

            const newTask = {
                title: input.files[0].name.split(".")[0] || "Новый документ",
                dimension:
                    $activeDimension === "cloud"
                        ? "work"
                        : ($activeDimension as "work" | "personal"),
                text: ocrResult.summary || ocrResult.text,
                files: [input.files[0].name],
                x: pos.x,
                y: pos.y
            };
            tasks.addTask(newTask as any);
            alert("Документ успешно обработан и добавлен на холст.");
        } catch (err) {
            console.error(err);
            alert("Ошибка при обработке файла.");
        } finally {
            isUploading = false;
            isExpanded = false;
        }
    }

    function createSimpleTask() {
        if (!taskTitle.trim()) return;

        const layout = container.getService<any>('Layout');
        const pos = layout.findEmptySpot($tasks);

        const newTask = {
            title: taskTitle,
            dimension:
                $activeDimension === "cloud"
                    ? "work"
                    : ($activeDimension as "work" | "personal"),
            x: pos.x,
            y: pos.y
        };

        tasks.addTask(newTask as any);
        taskTitle = "";
        isExpanded = false;
    }
</script>

<div class="quick-action-wrapper" class:expanded={isExpanded}>
    {#if isExpanded}
        <div class="expand-panel" transition:slide>
            <input
                type="text"
                bind:this={inputEl}
                bind:value={taskTitle}
                placeholder="Что нужно сделать?"
                on:keydown={(e) => e.key === "Enter" && createSimpleTask()}
            />

            <div class="actions">
                <label class="action-btn upload">
                    <input
                        type="file"
                        on:change={handleFileUpload}
                        disabled={isUploading}
                        hidden
                    />
                    <span>{isUploading ? "⌛" : "📄"}</span>
                </label>

                <button 
                    class="action-btn voice" 
                    class:recording={$whisper.status === 'recording'}
                    class:processing={$whisper.status === 'processing'}
                    on:click={toggleRecording}
                >
                    {#if $whisper.status === 'processing'}
                         <span class="spinner">🌀</span>
                    {:else}
                         <span>🎤</span>
                    {/if}
                </button>

                <button class="add-confirm" on:click={createSimpleTask}
                    >Done</button
                >
            </div>
        </div>
    {/if}

    <button class="main-fab" on:click={toggleExpand}>
        <span class="icon">{isExpanded ? "×" : "+"}</span>
    </button>
</div>

<style>
    .quick-action-wrapper {
        position: fixed;
        bottom: 2rem;
        left: 50%;
        transform: translateX(-50%);
        z-index: 8000;
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 1rem;
    }

    .main-fab {
        width: 60px;
        height: 60px;
        border-radius: 50%;
        border: none;
        background: linear-gradient(135deg, #4fc3f7, #2196f3);
        color: white;
        font-size: 2rem;
        cursor: pointer;
        box-shadow: 0 8px 24px rgba(33, 150, 243, 0.4);
        display: flex;
        align-items: center;
        justify-content: center;
        transition: all 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275);
    }

    .main-fab:hover {
        transform: scale(1.1) rotate(90deg);
        box-shadow: 0 12px 32px rgba(33, 150, 243, 0.6);
    }

    .expand-panel {
        background: rgba(20, 20, 35, 0.9);
        backdrop-filter: blur(20px);
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 2rem;
        padding: 1.5rem;
        width: 320px;
        display: flex;
        flex-direction: column;
        gap: 1rem;
        box-shadow: 0 20px 50px rgba(0, 0, 0, 0.5);
    }

    .expand-panel input {
        background: rgba(255, 255, 255, 0.05);
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 1rem;
        padding: 0.8rem 1rem;
        color: white;
        font-family: "Inter", sans-serif;
        outline: none;
    }

    .actions {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .action-btn {
        width: 42px;
        height: 42px;
        border-radius: 50%;
        background: rgba(255, 255, 255, 0.05);
        display: flex;
        align-items: center;
        justify-content: center;
        cursor: pointer;
        font-size: 1.2rem;
        transition: all 0.2s;
        border: 1px solid rgba(255, 255, 255, 0.1);
    }

    .action-btn:hover {
        background: rgba(255, 255, 255, 0.1);
    }

    .add-confirm {
        background: white;
        color: black;
        border: none;
        padding: 0.5rem 1.2rem;
        border-radius: 1rem;
        font-weight: 600;
        cursor: pointer;
    }

    .voice.recording {
        background: rgba(255, 0, 0, 0.2);
        border-color: #ff5252;
        box-shadow: 0 0 15px rgba(255, 82, 82, 0.4);
        animation: pulse-red 1.5s infinite;
    }

    .voice.processing {
        border-color: #4fc3f7;
        box-shadow: 0 0 15px rgba(79, 195, 247, 0.4);
    }

    .spinner {
        display: inline-block;
        animation: spin 2s linear infinite;
    }

    @keyframes pulse-red {
        0% { transform: scale(1); box-shadow: 0 0 0 0 rgba(255, 82, 82, 0.7); }
        70% { transform: scale(1.1); box-shadow: 0 0 0 10px rgba(255, 82, 82, 0); }
        100% { transform: scale(1); box-shadow: 0 0 0 0 rgba(255, 82, 82, 0); }
    }

    @keyframes spin {
        from { transform: rotate(0deg); }
        to { transform: rotate(360deg); }
    }
</style>
