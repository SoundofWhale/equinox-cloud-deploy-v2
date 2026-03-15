<script lang="ts">
    import { onMount } from 'svelte';
    import { page } from '$app/stores';
    import { context } from '$lib/stores/context';
    import { tasks } from '$lib/stores/tasks';
    import { container, initializeBaseServices } from '$lib/services';

    let loading = true;
    let error: string | null = null;

    onMount(async () => {
        initializeBaseServices();
        const taskId = $page.params.id;
        const apiUrl = container.getService<string>('API_URL');

        try {
            // Fetch the Context Packet for this specific task
            const res = await fetch(`${apiUrl}/context?task_id=${taskId}`);
            if (res.ok) {
                const packet = await res.json();
                context.hydrate(packet);
            } else {
                error = "Failed to load task context.";
            }
        } catch (e) {
            error = "Connection error.";
        } finally {
            loading = false;
        }
    });

    function saveTask(task) {
        tasks.updateTask(task.id, { text: task.text, title: task.title });
        context.emit({
            type: 'TASK_UPDATE',
            payload: { id: task.id, title: task.title }
        });
    }
</script>

<div class="session-container">
    {#if loading}
        <div class="status">🔌 Подключение к разъему сессии...</div>
    {:else if error}
        <div class="status error">❌ Ошибка: {error}</div>
    {:else if $tasks.length > 0}
        <div class="pedal-view">
            <header>
                <span class="icon">🎸</span>
                <h1>Изолированная сессия: {$tasks[0].title}</h1>
                <span class="session-id">SID: {$context?.sessionId?.slice(0,8)}</span>
            </header>

            <main>
                <div class="field">
                    <label>Заголовок задачи (Вход)</label>
                    <input type="text" bind:value={$tasks[0].title} on:blur={() => saveTask($tasks[0])} />
                </div>

                <div class="field">
                    <label>Описание задачи</label>
                    <textarea bind:value={$tasks[0].text} on:blur={() => saveTask($tasks[0])}></textarea>
                </div>

                <div class="status-box">
                    <p>Этот компонент работает как «Педаль» эффектов. Он видит только одну задачу и отправляет сигналы (Выход) при изменениях.</p>
                </div>
            </main>
        </div>
    {/if}
</div>

<style>
    .session-container {
        display: flex;
        justify-content: center;
        align-items: center;
        min-height: 100vh;
        background: #0a0a1a;
        color: #e0e0e0;
        font-family: 'Inter', sans-serif;
    }

    .pedal-view {
        width: 100%;
        max-width: 600px;
        background: #1a1a2e;
        border: 2px solid #4fc3f7;
        border-radius: 12px;
        padding: 2rem;
        box-shadow: 0 0 30px rgba(79, 195, 247, 0.2);
    }

    header {
        display: flex;
        align-items: center;
        gap: 1rem;
        margin-bottom: 2rem;
        border-bottom: 1px solid #333;
        padding-bottom: 1rem;
    }

    h1 {
        font-size: 1.2rem;
        margin: 0;
        color: #4fc3f7;
    }

    .session-id {
        margin-left: auto;
        font-family: monospace;
        color: #666;
        font-size: 0.8rem;
    }

    .field {
        margin-bottom: 1.5rem;
    }

    label {
        display: block;
        font-size: 0.8rem;
        color: #888;
        margin-bottom: 0.5rem;
    }

    input, textarea {
        width: 100%;
        background: #0f0f1f;
        border: 1px solid #333;
        border-radius: 6px;
        color: white;
        padding: 0.8rem;
        outline: none;
    }

    input:focus, textarea:focus {
        border-color: #4fc3f7;
    }

    textarea {
        height: 150px;
        resize: vertical;
    }

    .status-box {
        margin-top: 2rem;
        padding: 1rem;
        background: rgba(79, 195, 247, 0.1);
        border-radius: 6px;
        font-size: 0.9rem;
        line-height: 1.4;
        color: #4fc3f7;
    }

    .status {
        font-size: 1.2rem;
        color: #4fc3f7;
    }

    .error {
        color: #ff5252;
    }
</style>
