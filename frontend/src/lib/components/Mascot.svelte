<script lang="ts">
    import { fade } from "svelte/transition";
    import { mascotStore } from "$lib/stores/mascot";
    import type { MascotState } from "$lib/stores/mascot";

    const stateConfig: Record<
        MascotState,
        { label: string; emoji: string; color: string }
    > = {
        focus: { label: "Каска", emoji: "🪖", color: "#ff9800" },
        flow: { label: "Прозрачный", emoji: "👻", color: "#4fc3f7" },
        warning: { label: "Молния", emoji: "⚡", color: "#fdd835" },
        emergency: { label: "Ромб", emoji: "🔶", color: "#f44336" },
        rest: { label: "Пушистый", emoji: "☁️", color: "#66bb6a" },
        cough: { label: "Кашель", emoji: "🤧", color: "#ef5350" },
    };

    $: config = stateConfig[$mascotStore.state] || stateConfig.flow;
</script>

{#if $mascotStore.visible}
    <div class="mascot-overlay" transition:fade={{ duration: 300 }}>
        <div class="mascot-bubble" style="--mascot-color: {config.color}">
            <span class="mascot-emoji">{config.emoji}</span>
            <span class="mascot-label">{config.label}</span>
        </div>
    </div>
{/if}

<style>
    .mascot-overlay {
        position: fixed;
        bottom: 2rem;
        right: 2rem;
        z-index: 9000;
        pointer-events: none;
    }

    .mascot-bubble {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.6rem 1rem;
        background: rgba(0, 0, 0, 0.7);
        border: 1px solid var(--mascot-color, #4fc3f7);
        border-radius: 1.5rem;
        box-shadow:
            0 0 20px color-mix(in srgb, var(--mascot-color) 30%, transparent),
            0 8px 32px rgba(0, 0, 0, 0.4);
        backdrop-filter: blur(12px);
        animation: mascotPulse 2s ease-in-out infinite;
    }

    .mascot-emoji {
        font-size: 1.5rem;
        line-height: 1;
    }

    .mascot-label {
        font-family: "Inter", "Segoe UI", sans-serif;
        font-size: 0.8rem;
        font-weight: 500;
        color: rgba(255, 255, 255, 0.8);
        letter-spacing: 0.04em;
    }

    @keyframes mascotPulse {
        0%,
        100% {
            transform: scale(1);
            opacity: 0.9;
        }
        50% {
            transform: scale(1.05);
            opacity: 1;
        }
    }
</style>
