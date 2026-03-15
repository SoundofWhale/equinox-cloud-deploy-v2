<script lang="ts">
	import { onMount, onDestroy } from "svelte";
	import {
		tasks,
		workTasks,
		personalTasks,
		type Task,
	} from "../stores/tasks";
	import { activeDimension, type DimensionType } from "../stores/dimension";
	import { fade, slide } from "svelte/transition";
	import { uploadFile } from "../utils/api";
	import TaskDetailPanel from "../components/TaskDetailPanel.svelte";

	export let dimension: DimensionType = "work";

	let selectedTaskId: string | null = null;
	let selectedTask: Task | null = null;

	let draggedNode: any = null;

	$: {
		if (selectedTaskId && $tasks) {
			selectedTask = $tasks.find(t => t.id === selectedTaskId) || null;
		} else {
			selectedTask = null;
			// Auto-select the first task if it's the only one (for new users)
			if ($tasks && $tasks.length === 1 && !selectedTaskId) {
				selectedTaskId = $tasks[0].id;
			}
		}
	}


	let canvasEl: HTMLDivElement;
	let app: any;
	
	// Method 2: Persistent Layers
	let workWorld: any;
	let personalWorld: any;
	let cloudWorld: any;
	
	// Method 3: Viewport Persistence
	let viewportStates: Record<DimensionType, { x: number, y: number, scale: number }> = {
		work: { x: 0, y: 0, scale: 1 },
		personal: { x: 0, y: 0, scale: 1 },
		cloud: { x: 0, y: 0, scale: 1 }
	};
	let currentDimension: DimensionType = dimension;

	let nodes: any[] = [];
	let clouds: any[] = [];
	let tickerCallback: any = null;
	let isBackgroundClick = false;

	// Dragging individual nodes
	let lastPos = { x: 0, y: 0 };
	let dragClickPos = { x: 0, y: 0 }; // Screen pos where click started
	let dragOffset = { x: 0, y: 0 };   // Offset from node center to mouse
	let dragMovedThreshold = false;   // True if moved more than 5px

	// Lazy-loaded modules (client-side only)
	let pixi: any;
	let workDim: any;
	let personalDim: any;
	let cloudMode: any;
	let smartFocusMod: any;

	// Background colors per dimension
	const BG_COLORS: Record<DimensionType, number> = {
		work: 0x0a0a1a,
		personal: 0x1a1208,
		cloud: 0x0d0d1f,
	};

	onMount(async () => {
		// Dynamic imports — PixiJS only loads in the browser
		pixi = await import("pixi.js");
		workDim = await import("./WorkDimension");
		personalDim = await import("./PersonalDimension");
		cloudMode = await import("./CloudMode");
		smartFocusMod = await import("../utils/smartFocus");

		const { Application, Container } = pixi;

		app = new Application();
		await app.init({
			resizeTo: window,
			backgroundColor: BG_COLORS[dimension],
			antialias: true,
			resolution: window.devicePixelRatio || 1,
		});
		if (canvasEl) {
			canvasEl.appendChild(app.canvas);
		}

		// Initialize persistent layers
		workWorld = new Container();
		personalWorld = new Container();
		cloudWorld = new Container();
		
		app.stage.addChild(workWorld);
		app.stage.addChild(personalWorld);
		app.stage.addChild(cloudWorld);

		// Center the worlds initially
		[workWorld, personalWorld, cloudWorld].forEach(w => {
			w.x = window.innerWidth / 2;
			w.y = window.innerHeight / 2;
		});

		setupPanZoom();
		renderDimension(dimension);
	});

	onDestroy(() => {
		if (tickerCallback && app) {
			app.ticker.remove(tickerCallback);
		}
		if (app) {
			app.destroy(true, { children: true });
		}
	});

	function getActiveContainer() {
		if (dimension === "work") return workWorld;
		if (dimension === "personal") return personalWorld;
		return cloudWorld;
	}

	function setupPanZoom() {
		let dragging = false;
		let dragMoved = false;

		// Stage background click (PIXI events respect stopPropagation)
		app.stage.eventMode = 'static';
		app.stage.hitArea = app.screen;
		app.stage.on("pointerdown", (e: any) => {
			isBackgroundClick = true;
		});

		app.canvas.addEventListener("pointerdown", (e: PointerEvent) => {
			dragging = true;
			dragMoved = false;
			lastPos = { x: e.clientX, y: e.clientY };
		});

		app.canvas.addEventListener("pointermove", (e: PointerEvent) => {
			const activeWorld = getActiveContainer();
			if (!activeWorld) return;

			if (draggedNode) {
				const dx_screen = e.clientX - dragClickPos.x;
				const dy_screen = e.clientY - dragClickPos.y;
				const dist = Math.sqrt(dx_screen * dx_screen + dy_screen * dy_screen);
				
				if (dist > 15) {
					dragMovedThreshold = true;
				}

				if (dragMovedThreshold) {
					const mouseX_world = (e.clientX - activeWorld.x) / activeWorld.scale.x;
					const mouseY_world = (e.clientY - activeWorld.y) / activeWorld.scale.y;
					draggedNode.container.x = mouseX_world - dragOffset.x;
					draggedNode.container.y = mouseY_world - dragOffset.y;
				}
				
				lastPos = { x: e.clientX, y: e.clientY };
				return;
			}

			if (!dragging) return;
			dragMoved = true;
			activeWorld.x += e.clientX - lastPos.x;
			activeWorld.y += e.clientY - lastPos.y;
			
			// Save state
			viewportStates[dimension].x = activeWorld.x;
			viewportStates[dimension].y = activeWorld.y;
			
			lastPos = { x: e.clientX, y: e.clientY };
		});

		app.canvas.addEventListener("pointerup", (e: PointerEvent) => {
			const activeWorld = getActiveContainer();
			if (draggedNode) {
				const dx_screen = e.clientX - dragClickPos.x;
				const dy_screen = e.clientY - dragClickPos.y;
				const totalDist = Math.sqrt(dx_screen * dx_screen + dy_screen * dy_screen);

				if (totalDist > 15 || dragMovedThreshold) {
					// It was a drag
					tasks.updateTask(draggedNode.id, {
						x: draggedNode.container.x,
						y: draggedNode.container.y
					});
				} else {
					// It was a pure click!
					smartFocusMod.activateFocus(draggedNode.id, nodes, activeWorld);
					selectedTaskId = draggedNode.id;
					isBackgroundClick = false; // Prevent background-click from closing it immediately
				}
				draggedNode = null;
			}

			if (isBackgroundClick && !dragMoved && smartFocusMod.isFocusActive()) {
				smartFocusMod.deactivateFocus(nodes);
				selectedTaskId = null;
			}
			dragging = false;
			dragMovedThreshold = false;
		});

		app.canvas.addEventListener("pointerleave", () => { dragging = false; });
		app.canvas.addEventListener("pointercancel", () => { dragging = false; });

		app.canvas.addEventListener("wheel", (e: WheelEvent) => {
			const activeWorld = getActiveContainer();
			if (!activeWorld) return;
			
			e.preventDefault();
			const delta = Math.sign(e.deltaY);
			const zoomSpeed = 0.08; 
			const scaleFactor = delta > 0 ? (1 - zoomSpeed) : (1 + zoomSpeed);
			const minScale = 0.15;
			const maxScale = 4.0;
			const currentScale = activeWorld.scale.x;
			let newScale = Math.max(minScale, Math.min(maxScale, currentScale * scaleFactor));

			if (newScale !== currentScale) {
				const mouseX = e.clientX;
				const mouseY = e.clientY;
				const worldX = (mouseX - activeWorld.x) / currentScale;
				const worldY = (mouseY - activeWorld.y) / currentScale;
				
				activeWorld.scale.set(newScale);
				activeWorld.x = mouseX - worldX * newScale;
				activeWorld.y = mouseY - worldY * newScale;
				
				// Save state
				viewportStates[dimension].x = activeWorld.x;
				viewportStates[dimension].y = activeWorld.y;
				viewportStates[dimension].scale = newScale;
			}
		}, { passive: false });
	}

	// Method 1: Tracking Map
	let taskContainers = new Map<string, any>();

	function renderDimension(dim: DimensionType) {
		console.log(`[PixiCanvas] Rendering dimension: ${dim} | Tasks: ${$tasks.length} | Selected: ${selectedTaskId}`);
		if (!app || !workWorld || !personalWorld || !cloudWorld) return;

		// Method 2: Toggle Visibility
		workWorld.visible = (dim === "work");
		personalWorld.visible = (dim === "personal");
		cloudWorld.visible = (dim === "cloud");

		// Method 3: Restore Viewport State
		const activeWorld = getActiveContainer();
		const state = viewportStates[dim];
		activeWorld.x = state.x;
		activeWorld.y = state.y;
		activeWorld.scale.set(state.scale);

		// Ticker cleanup
		if (tickerCallback) {
			app.ticker.remove(tickerCallback);
			tickerCallback = null;
		}

		// Initial render or update
		if (dim === "work") {
			renderWork(selectedTaskId);
		} else if (dim === "personal") {
			renderPersonal(selectedTaskId);
		} else {
			renderCloud();
		}
	}

	function renderWork(selId: string | null = null) {
		if (!workWorld) return;

		// Draw stars only once if not already there
		if (workWorld.children.length === 0) {
			workDim.drawStars(workWorld, 300);
		}

		const currentTasks = getCurrentTasks("work");
		const currentIds = new Set(currentTasks.map(t => t.id));

		// Remove old tasks
		for (const [id, container] of taskContainers) {
			const task = currentTasks.find(t => t.id === id);
			if (!task || task.dimension !== "work") {
				if (workWorld.children.includes(container)) {
					workWorld.removeChild(container);
					taskContainers.delete(id);
				}
			}
		}

		nodes = [];
		currentTasks.forEach((task: Task) => {
			let container = taskContainers.get(task.id);
			
			if (!container) {
				// New Task: Create
				container = workDim.drawPlanet(workWorld, task, task.id === selId);
				taskContainers.set(task.id, container);
			} else {
				// Existing Task: Update stability
				container.x = task.x;
				container.y = task.y;
				
				// Redraw internals to handle selection/completion
				// Optimization: We should ideally update existing sprites, 
				// but for now we'll keep it simple but avoid full container wipe if not needed.
				if (container.children.length === 0) {
					const dummy = new pixi.Container();
					const fresh = workDim.drawPlanet(dummy, task, task.id === selId);
					while(fresh.children.length > 0) {
						container.addChild(fresh.children[0]);
					}
				} else {
					// Update basic props
					container.x = task.x;
					container.y = task.y;
				}
			}

			nodes.push({ id: task.id, container });

			// Re-bind events (cheaper than rebuilding everything)
			container.eventMode = 'static';
			container.removeAllListeners();

			container.on("pointerdown", (e: any) => {
				e.stopPropagation();
				draggedNode = { id: task.id, container };
				dragClickPos = { x: e.clientX, y: e.clientY };
				dragMovedThreshold = false;
				const activeWorld = getActiveContainer();
				const mouseX_world = (e.clientX - activeWorld.x) / activeWorld.scale.x;
				const mouseY_world = (e.clientY - activeWorld.y) / activeWorld.scale.y;
				dragOffset = { x: mouseX_world - container.x, y: mouseY_world - container.y };
			});

			container.on("pointertap", (e: any) => {
				e.stopPropagation();
				if (!dragMovedThreshold) {
					const activeWorld = getActiveContainer();
					smartFocusMod.activateFocus(task.id, nodes, activeWorld);
					selectedTaskId = task.id;
					isBackgroundClick = false;
				}
			});
		});

		tickerCallback = () => {
			const elapsed = performance.now();
			currentTasks.forEach((task: Task) => {
				const container = taskContainers.get(task.id);
				if (container) {
					workDim.animateSatellites(container, task, elapsed);
				}
			});
		};
		app.ticker.add(tickerCallback);
	}

	function renderPersonal(selId: string | null = null) {
		if (!personalWorld) return;

		// Draw background only once
		if (personalWorld.children.length === 0) {
			personalDim.drawGrain(personalWorld);
			personalDim.drawRoot(personalWorld);
		}

		const currentTasks = getCurrentTasks("personal");
		const currentIds = new Set(currentTasks.map(t => t.id));

		// Remove old tasks
		for (const [id, container] of taskContainers) {
			const task = currentTasks.find(t => t.id === id);
			if (!task || task.dimension !== "personal") {
				if (personalWorld.children.includes(container)) {
					personalWorld.removeChild(container);
					taskContainers.delete(id);
				}
			}
		}

		nodes = [];
		currentTasks.forEach((task: Task) => {
			let container = taskContainers.get(task.id);
			
			if (!container) {
				container = personalDim.drawBranch(personalWorld, task, task.id === selId);
				taskContainers.set(task.id, container);
			} else {
				container.x = 0; // Branches are relative to root or parent, but in our flat view they use toX/toY internally
				container.y = 0;
				
				container.removeChildren();
				const dummy = new pixi.Container();
				const fresh = personalDim.drawBranch(dummy, task, task.id === selId);
				while(fresh.children.length > 0) {
					container.addChild(fresh.children[0]);
				}
			}

			nodes.push({ id: task.id, container });

			container.eventMode = 'static';
			container.removeAllListeners();

			container.on("pointerdown", (e: any) => {
				e.stopPropagation();
				draggedNode = { id: task.id, container };
				dragClickPos = { x: e.clientX, y: e.clientY };
				dragMovedThreshold = false;
				const activeWorld = getActiveContainer();
				const mouseX_world = (e.clientX - activeWorld.x) / activeWorld.scale.x;
				const mouseY_world = (e.clientY - activeWorld.y) / activeWorld.scale.y;
				// Personal dimension containers are at 0,0; the 'node' is at task.x, task.y
				dragOffset = { x: mouseX_world - task.x, y: mouseY_world - task.y };
			});

			container.on("pointertap", (e: any) => {
				e.stopPropagation();
				if (!dragMovedThreshold) {
					const activeWorld = getActiveContainer();
					smartFocusMod.activateFocus(task.id, nodes, activeWorld);
					selectedTaskId = task.id;
					isBackgroundClick = false;
				}
			});
		});
	}

	function renderCloud() {
		clouds = cloudMode.createDemoClouds(cloudWorld);

		tickerCallback = () => {
			cloudMode.tickClouds(clouds, window.innerWidth, window.innerHeight);
		};
		app.ticker.add(tickerCallback);
	}

	function getCurrentTasks(dim: "work" | "personal"): Task[] {
		let result: Task[] = [];
		const unsub = (dim === "work" ? workTasks : personalTasks).subscribe(
			(v) => {
				result = v;
			},
		);
		unsub(); // Using a one-off read because we handle full reactivity with the $tasks block below
		return result;
	}

	// React to dimension OR tasks changes OR selection changes
	$: if (app && app.renderer && app.renderer.background && dimension && $tasks && selectedTaskId !== undefined) {
		app.renderer.background.color = BG_COLORS[dimension];
		renderDimension(dimension);
	}

	function handleClosePanel() {
		selectedTaskId = null;
		smartFocusMod.deactivateFocus(nodes);
	}
</script>

<div class="canvas-wrapper" bind:this={canvasEl}></div>

{#if selectedTask}
	<div class="ui-layer" transition:fade={{ duration: 200 }}>
		<TaskDetailPanel
			task={selectedTask}
			onClose={() => {
				selectedTaskId = null;
			}}
		/>
	</div>
{/if}

<style>
	.canvas-wrapper {
		position: fixed;
		inset: 0;
		z-index: 0;
	}

	.ui-layer {
		position: fixed;
		inset: 0;
		z-index: 9000;
		pointer-events: none;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	:global(.ui-layer > *) {
		pointer-events: auto;
	}
</style>
