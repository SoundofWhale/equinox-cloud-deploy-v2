import { writable } from 'svelte/store';

export type DimensionType = 'work' | 'personal' | 'cloud';

export const activeDimension = writable<DimensionType>('work');
export const isGatewayActive = writable(false);

const GATEWAY_QUOTES = [
    'Баланс — это искусство, а не цель.',
    'Каждое дыхание — маленький рестарт.',
    'Между хаосом и порядком — покой.',
    'Ты больше, чем твой список задач.',
    'Глубина важнее скорости.',
    'Остановись. Почувствуй. Продолжай.',
];

export function getRandomQuote(): string {
    return GATEWAY_QUOTES[Math.floor(Math.random() * GATEWAY_QUOTES.length)];
}

/** Trigger the Gateway transition animation between dimensions. */
export function switchDimension(target: DimensionType): void {
    isGatewayActive.set(true);

    // Play breath audio
    try {
        const audio = new Audio('/assets/audio/equi_breath.mp3');
        audio.volume = 0.4;
        audio.play().catch(() => { });
    } catch { }

    // After 3 seconds, switch and close gateway
    setTimeout(() => {
        activeDimension.set(target);
        isGatewayActive.set(false);
    }, 3000);
}
