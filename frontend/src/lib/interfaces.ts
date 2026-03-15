/**
 * Equinox 2.0 - Pedal Architecture Interfaces
 * Defines the standard "Inputs" (Packets) and "Outputs" (Actions/Events)
 * for isolated components and sessions.
 */

import type { Task } from './stores/tasks';
export type { Task };

/**
 * ContextPacket: The "Input" provided to a module or session.
 * It contains only the data needed for the current scope.
 */
export interface ContextPacket {
    sessionId: string;
    version: string;
    targetTaskId?: string; // If null, the entire dimension is the context
    data: {
        tasks: Task[];
        dimension: 'work' | 'personal' | 'cloud';
    };
    constraints?: {
        readOnly: boolean;
        aiAssistance: boolean;
    };
}

/**
 * ActionEvent: The "Output" emitted by a module.
 * Instead of direct DB writes, modules emit these events to the Orchestrator.
 */
export interface ActionEvent {
    type: 'TASK_UPDATE' | 'TASK_CREATE' | 'TASK_DELETE' | 'SUBTASK_TOGGLE' | 'DIMENSION_SWITCH';
    sessionId: string;
    timestamp: number;
    payload: any;
}

/**
 * ServiceContainer: Standard interface for Dependency Injection.
 * Allows components to request services (AI, Audio, API) without knowing implementation.
 */
export interface ServiceContainer {
    getService<T>(name: string): T;
    registerService(name: string, service: any): void;
}
