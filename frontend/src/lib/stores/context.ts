import { writable } from 'svelte/store';
import type { ContextPacket, ActionEvent } from '../interfaces';
import { tasks } from './tasks';

/**
 * context store: Manages the current session's "Packet".
 * This is the isolation layer.
 */
function createContextStore() {
    const { subscribe, set } = writable<ContextPacket | null>(null);

    return {
        subscribe,
        
        /**
         * Initialize the session from a packet.
         * This can be called when a task-specific window opens.
         */
        hydrate: (packet: ContextPacket) => {
            set(packet);
            
            // Sync the local task store with the packet data
            // This is the "Injected" state.
            tasks.set(packet.data.tasks);
            
            console.log(`[Session] Hydrated context for task: ${packet.targetTaskId || 'Global'}`);
        },

        /**
         * Submit an action back to the Orchestrator.
         * In a modular world, we don't write to DB, we emit events.
         */
        emit: (action: Omit<ActionEvent, 'sessionId' | 'timestamp'>) => {
            const event: ActionEvent = {
                ...action,
                sessionId: 'local-session', // Should be dynamic
                timestamp: Date.now()
            };
            
            console.log(`[Session] Emitting Output action:`, event);
            
            // In a real app, this would go through a WebSocket or shared worker
            // For now, we simulate by also calling the global API
        }
    };
}

export const context = createContextStore();
