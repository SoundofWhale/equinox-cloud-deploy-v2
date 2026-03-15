import { writable } from 'svelte/store';
import type { ServiceContainer, Task } from './interfaces';

class LayoutService {
    /**
     * Finds an empty spot on the canvas to prevent overlapping.
     * Uses a simple spiral search algorithm.
     */
    findEmptySpot(existingTasks: Task[]): { x: number, y: number } {
        const minDistance = 120; // Distance between planets
        let x = 0;
        let y = 0;
        let angle = 0;
        let step = 20;

        // Spiral search
        for (let i = 0; i < 200; i++) { // Limit search to 200 attempts
            let collision = false;
            for (const task of existingTasks) {
                const dx = x - task.x;
                const dy = y - task.y;
                const dist = Math.sqrt(dx * dx + dy * dy);
                if (dist < minDistance) {
                    collision = true;
                    break;
                }
            }

            if (!collision) return { x, y };

            // Move further out in the spiral
            angle += 0.5;
            x = (step * angle) * Math.cos(angle);
            y = (step * angle) * Math.sin(angle);
        }

        return { x: (Math.random() - 0.5) * 400, y: (Math.random() - 0.5) * 400 }; // Fallback
    }
}

class Container implements ServiceContainer {
    private services: Map<string, any> = new Map();

    registerService(name: string, service: any): void {
        this.services.set(name, service);
        console.log(`[DI] Registered service: ${name}`);
    }

    getService<T>(name: string): T {
        const service = this.services.get(name);
        if (!service) {
            throw new Error(`[DI] Service not found: ${name}`);
        }
        return service;
    }
}

export const container = new Container();

// Global registration of real services (The "Pedalboard")
// In an isolated session, some of these can be replaced by Mock services.
export function initializeBaseServices() {
    container.registerService('API_URL', '/api/v1');
    container.registerService('Layout', new LayoutService());
    
    // Example: placeholder for AI service
    container.registerService('AI', {
        query: async (prompt: string) => {
            console.log("[Pedal] AI Query:", prompt);
            return "AI Response Mock";
        }
    });
}
