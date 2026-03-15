import { Graphics, Container } from 'pixi.js';

/**
 * CloudMode — drifting translucent note rectangles.
 * No grid, no hierarchy. Each cloud floats with random velocity.
 */

export interface Cloud {
    id: string;
    text: string;
    container: Container;
    vx: number;
    vy: number;
    width: number;
    height: number;
}

const CLOUD_COLORS = [0x7c4dff, 0x448aff, 0x40c4ff, 0x64ffda, 0xffab40];

/** Create a cloud note visual. */
export function createCloud(
    parent: Container,
    id: string,
    text: string,
    startX: number,
    startY: number
): Cloud {
    const w = 120 + Math.random() * 80;
    const h = 60 + Math.random() * 30;
    const color = CLOUD_COLORS[Math.floor(Math.random() * CLOUD_COLORS.length)];

    const container = new Container();
    container.label = id;
    container.x = startX;
    container.y = startY;

    const g = new Graphics();

    // Rounded rectangle — translucent
    g.roundRect(-w / 2, -h / 2, w, h, 16);
    g.fill({ color, alpha: 0.18 });
    g.stroke({ color, width: 1, alpha: 0.3 });

    container.addChild(g);
    container.eventMode = 'static';
    container.cursor = 'pointer';
    parent.addChild(container);

    return {
        id,
        text,
        container,
        vx: (Math.random() - 0.5) * 0.6,
        vy: (Math.random() - 0.5) * 0.3,
        width: w,
        height: h,
    };
}

/** Tick update — move clouds and bounce off edges. */
export function tickClouds(clouds: Cloud[], canvasWidth: number, canvasHeight: number): void {
    clouds.forEach((cloud) => {
        cloud.container.x += cloud.vx;
        cloud.container.y += cloud.vy;

        // Bounce off edges
        const halfW = cloud.width / 2;
        const halfH = cloud.height / 2;

        if (cloud.container.x - halfW < -canvasWidth / 2 || cloud.container.x + halfW > canvasWidth / 2) {
            cloud.vx *= -1;
        }
        if (cloud.container.y - halfH < -canvasHeight / 2 || cloud.container.y + halfH > canvasHeight / 2) {
            cloud.vy *= -1;
        }
    });
}

/** Create a set of demo clouds. */
export function createDemoClouds(parent: Container): Cloud[] {
    const notes = [
        'Идея: новый формат встреч',
        'Прочитать книгу о фокусе',
        'Рецепт пасты с песто',
        'Цитата дня: «Начни сейчас»',
        'Позвонить маме',
        'Подкаст про продуктивность',
    ];

    return notes.map((text, i) => {
        const x = (Math.random() - 0.5) * 600;
        const y = (Math.random() - 0.5) * 400;
        return createCloud(parent, `cloud-${i}`, text, x, y);
    });
}
