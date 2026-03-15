import { Graphics, Container, Text, TextStyle } from 'pixi.js';
import type { Task } from '../stores/tasks';
import { planetRadius, satelliteOrbitRadius } from '../utils/weightCalc';

/**
 * WorkDimension — renders the Cosmic planetary system.
 * Tasks = Planets (circles), Subtasks = Satellites (orbiting dots).
 */

/** Draw star field background particles. */
export function drawStars(container: Container, count: number = 200): void {
    const g = new Graphics();
    for (let i = 0; i < count; i++) {
        const x = (Math.random() - 0.5) * 4000;
        const y = (Math.random() - 0.5) * 4000;
        const size = Math.random() * 1.5 + 0.5;
        const alpha = Math.random() * 0.6 + 0.2;
        g.circle(x, y, size);
        g.fill({ color: 0xffffff, alpha });
    }
    container.addChildAt(g, 0);
}

/** Render a planet (task node) with glow ring. Returns its container for hit testing. */
export function drawPlanet(
    parent: Container,
    task: Task,
    isSelected: boolean = false
): Container {
    const nodeContainer = new Container();
    nodeContainer.label = task.id;
    nodeContainer.x = task.x;
    nodeContainer.y = task.y;

    const radius = planetRadius(task.subtasks.length);
    const g = new Graphics();

    // Outer glow ring
    g.circle(0, 0, radius + 8);
    g.stroke({ color: 0x00e5ff, width: 2, alpha: 0.25 });

    // Selection ring (new)
    if (isSelected) {
        g.circle(0, 0, radius + 15);
        g.stroke({ color: 0x00ffff, width: 4, alpha: 0.8 });
    }

    // Planet body
    g.circle(0, 0, radius);
    g.fill({ color: task.completed ? 0x37474f : 0x4fc3f7, alpha: 0.9 });

    // If task has a timeBlock, glow orange on top edge
    if (task.timeBlock) {
        g.arc(0, 0, radius + 12, -Math.PI, 0);
        g.stroke({ color: 0xffa726, width: 3, alpha: 0.8 });
    }

    // Inner highlight
    g.circle(-radius * 0.25, -radius * 0.25, radius * 0.35);
    g.fill({ color: 0xffffff, alpha: 0.12 });

    nodeContainer.addChild(g);

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
    label.y = radius + 15;
    nodeContainer.addChild(label);

    // Draw satellites (subtasks)
    (task.subtasks || []).forEach((sub, i) => {
        if (!sub) return;
        const orbitR = satelliteOrbitRadius(radius, i);
        const angle = (i / Math.max(task.subtasks.length, 1)) * Math.PI * 2 - Math.PI / 2;

        const sg = new Graphics();

        // Orbit path (dotted circle)
        sg.circle(0, 0, orbitR);
        sg.stroke({ color: 0x4fc3f7, width: 1, alpha: 0.15 });

        // Satellite dot
        const sx = Math.cos(angle) * orbitR;
        const sy = Math.sin(angle) * orbitR;

        // Dust ring for nested children
        if (sub.childrenCount && sub.childrenCount > 0) {
            const fullness = Math.min(sub.childrenCount, 12) / 12;
            const ringR = 12 + fullness * 4;
            sg.circle(sx, sy, ringR);
            sg.stroke({ color: 0xcddc39, width: 2, alpha: 0.3 + (fullness * 0.4) });
            
            for (let j = 0; j < 6; j++) {
                const pAngle = (j / 6) * Math.PI * 2;
                const pR = ringR + 2;
                sg.circle(sx + Math.cos(pAngle)*pR, sy + Math.sin(pAngle)*pR, 1.5);
                sg.fill({ color: 0xffffff, alpha: 0.4 });
            }
        }

        sg.circle(sx, sy, sub.done ? 5 : 8);
        sg.fill({ color: sub.done ? 0x37474f : 0x5c6bc0, alpha: 0.85 });

        nodeContainer.addChild(sg);
    });

    // Make interactive
    nodeContainer.eventMode = 'static';
    nodeContainer.cursor = 'pointer';

    parent.addChild(nodeContainer);
    return nodeContainer;
}

/** Animate satellite orbits — call on each tick. */
export function animateSatellites(
    nodeContainer: Container,
    task: Task,
    elapsed: number
): void {
    const radius = planetRadius((task.subtasks || []).length);
    // Satellite children start after glow ring g + text label
    (task.subtasks || []).forEach((sub, i) => {
        // children[0] = g (planet ring + body + highlight)
        // children[1] = text (label)
        // children[2+] = satellite graphics
        const satGraphics = nodeContainer.children[i + 2] as Graphics | undefined;
        if (!satGraphics || !satGraphics.clear) return;

        const orbitR = satelliteOrbitRadius(radius, i);
        const speed = 0.00005 + i * 0.00002;
        const angle = elapsed * speed + (i / task.subtasks.length) * Math.PI * 2;

        // Redraw satellite position (clear and redraw)
        satGraphics.clear();

        // Orbit path
        satGraphics.circle(0, 0, orbitR);
        satGraphics.stroke({ color: 0x4fc3f7, width: 1, alpha: 0.15 });

        // Satellite at new angle
        const sx = Math.cos(angle) * orbitR;
        const sy = Math.sin(angle) * orbitR;

        // Dust ring for nested children
        if (sub.childrenCount && sub.childrenCount > 0) {
            const fullness = Math.min(sub.childrenCount, 12) / 12;
            const ringR = 12 + fullness * 4;
            satGraphics.circle(sx, sy, ringR);
            satGraphics.stroke({ color: 0xcddc39, width: 2, alpha: 0.3 + (fullness * 0.4) });
            
            const dustRotation = elapsed * 0.001;
            for (let j = 0; j < 6; j++) {
                const pAngle = (j / 6) * Math.PI * 2 + dustRotation;
                const pR = ringR + 2;
                satGraphics.circle(sx + Math.cos(pAngle)*pR, sy + Math.sin(pAngle)*pR, 1.5);
                satGraphics.fill({ color: 0xffffff, alpha: 0.4 });
            }
        }

        satGraphics.circle(sx, sy, sub.done ? 5 : 8);
        satGraphics.fill({ color: sub.done ? 0x37474f : 0x5c6bc0, alpha: 0.85 });
    });
}
