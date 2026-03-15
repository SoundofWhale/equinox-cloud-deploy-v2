<script lang="ts">
    import { onDestroy } from "svelte";
    import { isEmergencyMode } from "../stores/emergency";

    let holdTimer: ReturnType<typeof setTimeout> | null = null;
    let progress = 0;
    let holding = false;
    let activated = false;

    // Keep button synced if emergency is deactivated elsewhere
    $: {
        if (!$isEmergencyMode && activated) {
            activated = false;
        }
    }

    let progressInterval: ReturnType<typeof setInterval> | null = null;

    const HOLD_DURATION = 5000; // 5 seconds

    function onHoldStart() {
        if (activated) {
            // Quick deactivate if already on
            activated = false;
            isEmergencyMode.set(false);
            return;
        }

        holding = true;
        progress = 0;

        // Animate progress bar
        const startTime = Date.now();
        progressInterval = setInterval(() => {
            const elapsed = Date.now() - startTime;
            progress = Math.min((elapsed / HOLD_DURATION) * 100, 100);
        }, 30);

        holdTimer = setTimeout(async () => {
            try {
                await fetch("/api/v1/emergency/activate", { method: "POST" });
                activated = true;
                isEmergencyMode.set(true);

                // Play warning audio (suppressed local errors)
                try {
                    const audio = new Audio("/assets/audio/equi_warning.mp3");
                    audio.volume = 0.5;
                    audio.play().catch(() => {});
                } catch {}
            } catch (err) {
                console.error("Emergency activation failed:", err);
            }
        }, HOLD_DURATION);
    }

    function onHoldEnd() {
        holding = false;
        progress = 0;
        if (holdTimer) {
            clearTimeout(holdTimer);
            holdTimer = null;
        }
        if (progressInterval) {
            clearInterval(progressInterval);
            progressInterval = null;
        }
    }

    onDestroy(() => {
        onHoldEnd();
    });
</script>

<button
    class="emergency-btn"
    class:holding
    class:activated
    on:pointerdown={onHoldStart}
    on:pointerup={onHoldEnd}
    on:pointerleave={onHoldEnd}
>
    {#if activated}
        🚨 EMERGENCY ACTIVE
    {:else}
        ⚡ Hold 5s for Emergency
    {/if}

    {#if holding && !activated}
        <div class="progress-bar">
            <div class="progress-fill" style="width: {progress}%"></div>
        </div>
    {/if}
</button>

<style>
    .emergency-btn {
        position: fixed;
        bottom: 24px;
        right: 24px;
        z-index: 150;
        padding: 12px 20px;
        border: 2px solid rgba(255, 152, 0, 0.4);
        border-radius: 12px;
        background: rgba(30, 20, 10, 0.85);
        backdrop-filter: blur(12px);
        color: #ffb300;
        font-family: "Inter", sans-serif;
        font-size: 13px;
        font-weight: 600;
        letter-spacing: 0.5px;
        cursor: pointer;
        user-select: none;
        transition: all 200ms ease;
        overflow: hidden;
    }

    .emergency-btn:hover {
        border-color: #ff9800;
        background: rgba(50, 30, 10, 0.9);
    }

    .emergency-btn.holding {
        border-color: #ff5722;
        box-shadow: 0 0 20px rgba(255, 87, 34, 0.3);
        animation: pulse-emergency 500ms infinite;
    }

    .emergency-btn.activated {
        border-color: #f44336;
        background: rgba(60, 10, 10, 0.9);
        color: #ff5252;
        box-shadow: 0 0 30px rgba(244, 67, 54, 0.4);
        animation: pulse-active 1.5s infinite;
    }

    .progress-bar {
        position: absolute;
        bottom: 0;
        left: 0;
        right: 0;
        height: 3px;
        background: rgba(255, 255, 255, 0.1);
    }

    .progress-fill {
        height: 100%;
        background: linear-gradient(90deg, #ff9800, #ff5722);
        transition: width 30ms linear;
    }

    @keyframes pulse-emergency {
        0%,
        100% {
            transform: scale(1);
        }
        50% {
            transform: scale(1.03);
        }
    }

    @keyframes pulse-active {
        0%,
        100% {
            box-shadow: 0 0 20px rgba(244, 67, 54, 0.3);
        }
        50% {
            box-shadow: 0 0 40px rgba(244, 67, 54, 0.6);
        }
    }
</style>
