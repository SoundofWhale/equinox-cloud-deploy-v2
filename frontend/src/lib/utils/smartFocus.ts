import { Container, BlurFilter } from 'pixi.js';

export interface FocusableNode {
    id: string;
    container: Container;
}

let focusedId: string | null = null;

/**
 * Activate Smart Focus on a node — all others fade + blur.
 * Clicked node stays at full opacity with no filters.
 */
export function activateFocus(
    clickedId: string,
    allNodes: FocusableNode[],
    worldContainer: Container
): void {
    focusedId = clickedId;

    allNodes.forEach((node) => {
        if (node.id === clickedId) {
            node.container.alpha = 1.0;
            node.container.filters = [];
        } else {
            node.container.alpha = 0.15;
            node.container.filters = [new BlurFilter({ strength: 8 })];
        }
    });

    // Center viewport on clicked node using local logical coordinates
    const target = allNodes.find((n) => n.id === clickedId);
    if (target) {
        const cx = target.container.x;
        const cy = target.container.y;
        worldContainer.x = -cx * worldContainer.scale.x + window.innerWidth / 2;
        worldContainer.y = -cy * worldContainer.scale.y + window.innerHeight / 2;
    }
}

/**
 * Deactivate Smart Focus — restore all nodes to full opacity.
 */
export function deactivateFocus(allNodes: FocusableNode[]): void {
    focusedId = null;
    allNodes.forEach((node) => {
        node.container.alpha = 1.0;
        node.container.filters = [];
    });
}

/** Returns whether Smart Focus is currently active. */
export function isFocusActive(): boolean {
    return focusedId !== null;
}
