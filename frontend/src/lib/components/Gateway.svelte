<script lang="ts">
    import { onMount, createEventDispatcher } from "svelte";
    import { fade } from "svelte/transition";
    import { mascotStore, showMascot } from "$lib/stores/mascot";

    const dispatch = createEventDispatcher();

    const quotes = [
        "Каждый момент — это новое начало.",
        "Ты контролируешь своё время.",
        "Один шаг за раз.",
        "Дыши. Фокусируйся. Действуй.",
        "Путь в тысячу ли начинается с одного шага.",
        "Тишина — это тоже ответ.",
        "Сила в спокойствии.",
        "Будь водой, друг мой.",
    ];

    let quote = quotes[Math.floor(Math.random() * quotes.length)];

    onMount(() => {
        showMascot("flow", 3000);
        setTimeout(() => dispatch("complete"), 3000);
    });
</script>

<div class="gateway" transition:fade={{ duration: 400 }}>
    <div class="gateway-bg"></div>

    <div class="gateway-content">
        <div class="breath-circle"></div>
        <p class="instruction">Сделай глубокий вдох</p>
        <p class="quote">{quote}</p>

        {#if $mascotStore.visible}
            <div class="mascot-container" transition:fade={{ duration: 300 }}>
                <div class="mascot mascot-{$mascotStore.state}">
                    {#if $mascotStore.state === "flow"}
                        <svg viewBox="0 0 48 48" width="48" height="48">
                            <circle
                                cx="24"
                                cy="24"
                                r="20"
                                fill="none"
                                stroke="rgba(79,195,247,0.6)"
                                stroke-width="2"
                            />
                            <circle
                                cx="24"
                                cy="24"
                                r="12"
                                fill="rgba(79,195,247,0.3)"
                            />
                            <circle
                                cx="24"
                                cy="24"
                                r="5"
                                fill="rgba(79,195,247,0.8)"
                            />
                        </svg>
                    {/if}
                </div>
            </div>
        {/if}
    </div>
</div>

<style>
    .gateway {
        position: fixed;
        inset: 0;
        z-index: 9999;
        display: flex;
        align-items: center;
        justify-content: center;
    }

    .gateway-bg {
        position: absolute;
        inset: 0;
        background: radial-gradient(
            ellipse at center,
            rgba(10, 10, 26, 0.97) 0%,
            rgba(5, 5, 15, 0.99) 100%
        );
        backdrop-filter: blur(40px);
    }

    .gateway-content {
        position: relative;
        z-index: 1;
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 1.5rem;
    }

    .breath-circle {
        width: 100px;
        height: 100px;
        border-radius: 50%;
        background: radial-gradient(
            circle,
            rgba(79, 195, 247, 0.6) 0%,
            rgba(79, 195, 247, 0.15) 50%,
            transparent 70%
        );
        box-shadow:
            0 0 60px rgba(79, 195, 247, 0.3),
            0 0 120px rgba(79, 195, 247, 0.1);
        animation: breathe 3s ease-in-out;
    }

    @keyframes breathe {
        0% {
            transform: scale(0.6);
            opacity: 0.4;
        }
        40% {
            transform: scale(1.4);
            opacity: 1;
        }
        70% {
            transform: scale(1);
            opacity: 0.8;
        }
        100% {
            transform: scale(0.8);
            opacity: 0.6;
        }
    }

    .instruction {
        font-family: "Inter", "Segoe UI", sans-serif;
        font-size: 1.4rem;
        font-weight: 300;
        color: rgba(255, 255, 255, 0.85);
        letter-spacing: 0.05em;
        margin: 0;
    }

    .quote {
        font-family: "Inter", "Segoe UI", sans-serif;
        font-size: 0.95rem;
        font-weight: 200;
        color: rgba(255, 255, 255, 0.45);
        font-style: italic;
        max-width: 300px;
        text-align: center;
        line-height: 1.5;
        margin: 0;
    }

    .mascot-container {
        margin-top: 0.5rem;
    }

    .mascot {
        opacity: 0.7;
        animation: mascotFloat 2s ease-in-out infinite;
    }

    @keyframes mascotFloat {
        0%,
        100% {
            transform: translateY(0);
        }
        50% {
            transform: translateY(-4px);
        }
    }
</style>
