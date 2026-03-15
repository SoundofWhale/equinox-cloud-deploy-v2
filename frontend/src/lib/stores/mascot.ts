import { writable } from 'svelte/store';

export type MascotState = 'focus' | 'flow' | 'warning' | 'emergency' | 'rest' | 'cough';

export const mascotStore = writable<{ state: MascotState; visible: boolean }>({
    state: 'flow',
    visible: false
});

let hideTimeout: ReturnType<typeof setTimeout> | null = null;

/**
 * Show the mascot in a specific state for a given duration.
 * Auto-hides after the duration expires.
 */
export function showMascot(state: MascotState, duration = 4000) {
    if (hideTimeout) clearTimeout(hideTimeout);
    mascotStore.set({ state, visible: true });
    hideTimeout = setTimeout(() => {
        mascotStore.update(s => ({ ...s, visible: false }));
        hideTimeout = null;
    }, duration);
}

/**
 * Immediately hide the mascot.
 */
export function hideMascot() {
    if (hideTimeout) {
        clearTimeout(hideTimeout);
        hideTimeout = null;
    }
    mascotStore.update(s => ({ ...s, visible: false }));
}
