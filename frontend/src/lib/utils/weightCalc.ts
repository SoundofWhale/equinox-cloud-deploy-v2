/** Weight Dynamics — node size based on subtask count (from MASTER_SPEC) */

/**
 * Planet radius for Work dimension.
 * Formula: min(24 + subtasks × 4, 72) px
 */
export function planetRadius(subtaskCount: number): number {
    return Math.min(24 + subtaskCount * 4, 72);
}

/**
 * Branch thickness for Personal dimension.
 * Formula: min(2 + subtasks × 0.8, 12) px
 */
export function branchThickness(subtaskCount: number): number {
    return Math.min(2 + subtaskCount * 0.8, 12);
}

/**
 * Satellite orbit radius based on index.
 * Each satellite is placed further from the parent planet.
 */
export function satelliteOrbitRadius(parentRadius: number, index: number): number {
    return parentRadius + 30 + index * 20;
}
