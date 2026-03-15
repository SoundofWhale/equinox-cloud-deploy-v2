import { Graphics, Container, Text, TextStyle } from 'pixi.js';
import type { Task } from '../stores/tasks';
import { branchThickness } from '../utils/weightCalc';

/**
 * PersonalDimension — renders the Organic Tree.
 * Tasks = Branches (bezier curves). Subtasks = Leaf sub-branches.
 */

const ROOT_X = 0;
const ROOT_Y = 300;

/** Draw warm grain background texture. */
export function drawGrain(container: Container): void {
    const g = new Graphics();

    // Base warm tone
    g.rect(-2000, -2000, 4000, 4000);
    g.fill({ color: 0x1a1208 });

    // Subtle grain particles
    for (let i = 0; i < 150; i++) {
        const x = (Math.random() - 0.5) * 4000;
        const y = (Math.random() - 0.5) * 4000;
        const size = Math.random() * 2 + 0.5;
        g.circle(x, y, size);
        g.fill({ color: 0x3e2723, alpha: Math.random() * 0.3 + 0.05 });
    }

    container.addChildAt(g, 0);
}

/** Render a branch (task) with bezier curve from root, plus sub-branches. */
export function drawBranch(
    parent: Container,
    task: Task,
    isSelected: boolean = false
): Container {
    const nodeContainer = new Container();
    nodeContainer.label = task.id;

    const thickness = branchThickness(task.subtasks.length);
    const g = new Graphics();

    // Main branch: bezier from root to task position
    const fromX = ROOT_X;
    const fromY = ROOT_Y;
    const toX = task.x;
    const toY = task.y;

    // Control points for organic curve
    const cpX1 = fromX;
    const cpY1 = fromY - Math.abs(toY - fromY) * 0.4;
    const cpX2 = toX;
    const cpY2 = toY + Math.abs(toY - fromY) * 0.3;

    g.moveTo(fromX, fromY);
    g.bezierCurveTo(cpX1, cpY1, cpX2, cpY2, toX, toY);
    g.stroke({
        color: task.completed ? 0x5d4037 : 0x66bb6a,
        width: thickness,
        alpha: 0.9,
        cap: 'round',
    });

    // Leaf node at tip
    g.circle(toX, toY, 6 + task.subtasks.length * 1.5);
    g.fill({ color: task.completed ? 0x795548 : 0x81c784, alpha: 0.9 });

    // Inner glow
    g.circle(toX - 2, toY - 2, 3 + task.subtasks.length);
    g.fill({ color: 0xffffff, alpha: 0.1 });

    // Selection ring (new)
    if (isSelected) {
        g.circle(toX, toY, 15 + task.subtasks.length * 1.5);
        g.stroke({ color: 0x00ffff, width: 4, alpha: 0.8 });
    }

    nodeContainer.addChild(g);

    // Sub-branches (subtasks)
    task.subtasks.forEach((sub, i) => {
        const sg = new Graphics();
        const angle = -Math.PI / 4 + (i / Math.max(task.subtasks.length - 1, 1)) * (Math.PI / 2);
        const length = 40 + i * 15;

        const subX = toX + Math.cos(angle) * length;
        const subY = toY + Math.sin(angle) * length - 30;

        // Sub-branch bezier
        const subCpX = toX + Math.cos(angle) * length * 0.4;
        const subCpY = toY - 20;

        sg.moveTo(toX, toY);
        sg.bezierCurveTo(subCpX, subCpY, subX, subY + 15, subX, subY);
        sg.stroke({
            color: sub.done ? 0x795548 : 0xffb300,
            width: Math.max(thickness * 0.5, 1.5),
            alpha: 0.8,
            cap: 'round',
        });

        // Leaf dot
        sg.circle(subX, subY, sub.done ? 3 : 5);
        sg.fill({ color: sub.done ? 0x8d6e63 : 0xffca28, alpha: 0.9 });

        nodeContainer.addChild(sg);
    });

    // If task has timeBlock, draw a subtle glowing ring around the leaf node
    if (task.timeBlock) {
        g.circle(toX, toY, 15 + task.subtasks.length * 1.5);
        g.stroke({ color: 0xffa726, width: 2, alpha: 0.6 });
    }

    // Title label
    const style = new TextStyle({
        fontFamily: 'sans-serif',
        fontSize: 14,
        fill: '#ffffff',
        align: 'center',
        dropShadow: { color: 0x000000, alpha: 0.8, distance: 2, blur: 4 }
    });
    const label = new Text({ text: task.title, style });
    label.anchor.set(0.5, 0);
    label.x = toX;
    label.y = toY + (15 + task.subtasks.length * 1.5) + 15;
    nodeContainer.addChild(label);

    // Make interactive
    nodeContainer.eventMode = 'static';
    nodeContainer.cursor = 'pointer';

    parent.addChild(nodeContainer);
    return nodeContainer;
}

/** Draw the central root node (tree trunk base). */
export function drawRoot(container: Container): void {
    const g = new Graphics();

    // Trunk base
    g.circle(ROOT_X, ROOT_Y, 12);
    g.fill({ color: 0x4e342e, alpha: 0.9 });

    // Root glow
    g.circle(ROOT_X, ROOT_Y, 20);
    g.stroke({ color: 0x66bb6a, width: 2, alpha: 0.3 });

    container.addChild(g);
}
