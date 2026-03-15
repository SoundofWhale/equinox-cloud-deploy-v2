import { writable, derived, get } from 'svelte/store';
import { container } from '../services';
import { auth } from './auth';

export interface Subtask {
    id: string;
    title: string;
    done: boolean;
    childrenCount?: number;
}

export interface TimeBlock {
    start: string;
    end: string;
}

export interface Meeting {
    id: string;
    title: string;
    startTime: string;
    endTime: string;
}

export interface Task {
    id: string;
    title: string;
    dimension: 'work' | 'personal' | 'cloud';
    parentId?: string;
    subtasks: Subtask[];
    meetings: Meeting[];
    x: number;
    y: number;
    completed: boolean;
    completedAt?: number;
    text?: string;
    timeBlock?: TimeBlock;
    files?: string[];
    modules: string[];
    childrenCount: number;
}

export const DEFAULT_MODULES = ['description', 'checklist', 'attachments', 'ai_advice', 'meetings'];
export const CHILD_MODULES = ['description', 'checklist', 'ai_advice'];

function createTaskStore() {
    const { subscribe, set, update } = writable<Task[]>([]);

    const getApiUrl = () => {
        try {
            return container.getService<string>('API_URL');
        } catch (e) {
            return '/api/v1'; // Fallback
        }
    };

    const getAuthHeader = (): Record<string, string> => {
        const token = get(auth).token;
        return token ? { 'Authorization': `Bearer ${token}` } : {};
    };

    let isInitializing = false;

    return {
        subscribe,
        set,
        update,
        init: async () => {
            if (isInitializing) return;
            isInitializing = true;
            console.log("[Tasks] Starting init...");
            const apiUrl = getApiUrl();
            try {
                const res = await fetch(`${apiUrl}/tasks`, {
                    headers: { ...getAuthHeader() }
                });
                if (res.ok) {
                    const data = await res.json();
                    const mapped = (data || []).map((t: any) => ({
                        ...t,
                        parentId: (t.parent_id && t.parent_id !== "") ? t.parent_id : undefined,
                        subtasks: t.subtasks || [],
                        meetings: t.meetings || [],
                        modules: (t.modules && t.modules.length > 0) ? t.modules : DEFAULT_MODULES,
                        completed: !!t.completed_at,
                        childrenCount: t.children_count || 0
                    }));
                    set(mapped);

                    // If NO tasks exist for a new user, create a "Welcome" task automatically
                    if (mapped.length === 0) {
                        const welcomeTitle = "Добро пожаловать в EQUINOX";
                        // Call the internal addTask directly instead of creating a new store instance
                        const apiUrl = getApiUrl();
                        const taskModules = DEFAULT_MODULES;
                        try {
                            const res = await fetch(`${apiUrl}/tasks`, {
                                method: 'POST',
                                headers: { 
                                    'Content-Type': 'application/json',
                                    ...getAuthHeader()
                                },
                                body: JSON.stringify({
                                    title: welcomeTitle,
                                    template: 'task',
                                    dimension: 'work',
                                    x: 0,
                                    y: 0,
                                    text: "Это ваша первая планетарная задача. Нажмите на неё, чтобы открыть настройки.",
                                    files: [],
                                    modules: taskModules,
                                    meetings: []
                                })
                            });
                            if (res.ok) {
                                const created = await res.json();
                                created.parentId = undefined;
                                created.subtasks = [];
                                created.meetings = [];
                                created.modules = DEFAULT_MODULES;
                                created.completed = false;
                                created.childrenCount = 0;
                                update(tasks => [created, ...tasks]);
                            }
                        } catch (e) {
                            console.error("Failed to create welcome task", e);
                        }
                    }
                }
            } catch (err) {
                console.error("Failed to load tasks", err);
            } finally {
                isInitializing = false;
                console.log("[Tasks] Init complete.");
            }
        },
        addTask: async (task: Partial<Task>) => {
            const apiUrl = getApiUrl();
            const taskModules = (task.modules && task.modules.length > 0) ? task.modules : DEFAULT_MODULES;
            try {
                const res = await fetch(`${apiUrl}/tasks`, {
                    method: 'POST',
                    headers: { 
                        'Content-Type': 'application/json',
                        ...getAuthHeader()
                    },
                    body: JSON.stringify({
                        title: task.title,
                        template: task.modules ? 'custom' : 'task',
                        dimension: task.dimension || 'work',
                        parent_id: task.parentId,
                        x: task.x || 0,
                        y: task.y || 0,
                        text: task.text || '',
                        files: task.files || [],
                        modules: taskModules,
                        meetings: task.meetings || []
                    })
                });
                if (res.ok) {
                    const created = await res.json();
                    created.parentId = (created.parent_id && created.parent_id !== "") ? created.parent_id : undefined;
                    created.subtasks = created.subtasks || [];
                    created.meetings = created.meetings || [];
                    created.modules = (created.modules && created.modules.length > 0) ? created.modules : DEFAULT_MODULES;
                    created.completed = !!created.completed_at;
                    created.childrenCount = created.children_count || 0;
                    
                    update(tasks => {
                        const idx = tasks.findIndex(t => t.id === created.id);
                        if (idx !== -1) {
                            const newTasks = [...tasks];
                            newTasks[idx] = { ...newTasks[idx], ...created };
                            return newTasks;
                        }
                        return [created, ...tasks];
                    });
                }
            } catch (e) {
                console.error("Failed to create task", e);
            }
        },
        updateTask: async (id: string, updates: Partial<Task>) => {
            const apiUrl = getApiUrl();
            let fullTask: Task | undefined;
            update((tasks: Task[]) => {
                return tasks.map((t: Task) => {
                    if (t.id === id) {
                        fullTask = { ...t, ...updates };
                        return fullTask;
                    }
                    return t;
                });
            });
            if (fullTask) {
                fetch(`${apiUrl}/tasks/${id}`, {
                    method: 'PATCH',
                    headers: { 
                        'Content-Type': 'application/json',
                        ...getAuthHeader()
                    },
                    body: JSON.stringify({
                        title: fullTask.title,
                        text: fullTask.text,
                        x: fullTask.x,
                        y: fullTask.y,
                        files: fullTask.files,
                        modules: fullTask.modules,
                        parent_id: (fullTask.parentId && fullTask.parentId !== "") ? fullTask.parentId : null,
                        meetings: fullTask.meetings
                    })
                }).catch(e => console.error("Failed to update task", e));
            }
        },
        removeTask: async (id: string) => {
            const apiUrl = getApiUrl();
            update((tasks: Task[]) => tasks.filter((t: Task) => t.id !== id));
            fetch(`${apiUrl}/tasks/${id}`, { 
                method: 'DELETE',
                headers: { ...getAuthHeader() }
            });
        },
        completeTask: async (id: string) => {
            const apiUrl = getApiUrl();
            update((tasks: Task[]) =>
                tasks.map((t: Task) =>
                    t.id === id ? { ...t, completed: true, completedAt: Date.now() } : t
                )
            );
            fetch(`${apiUrl}/tasks/${id}/complete`, { 
                method: 'POST',
                headers: { ...getAuthHeader() }
            });
        },
        addSubtask: async (taskId: string, title: string) => {
            const apiUrl = getApiUrl();
            try {
                const res = await fetch(`${apiUrl}/subtasks/${taskId}`, {
                    method: 'POST',
                    headers: { 
                        'Content-Type': 'application/json',
                        ...getAuthHeader()
                    },
                    body: JSON.stringify({ title })
                });
                if (res.ok) {
                    const sub = await res.json();
                    update((tasks: Task[]) => tasks.map((t: Task) =>
                        t.id === taskId ? {
                            ...t,
                            subtasks: [...t.subtasks, sub]
                        } : t
                    ));
                }
            } catch (e) {
                console.error("Failed to add subtask", e);
            }
        },
        toggleSubtask: async (taskId: string, subtaskId: string) => {
            const apiUrl = getApiUrl();
            let isFullTask = false;
            let currentStatus = false;

            update((tasks: Task[]) => {
                const parent = tasks.find(t => t.id === taskId);
                const subtaskItem = parent?.subtasks.find(s => s.id === subtaskId);
                const branchedTask = tasks.find(t => t.id === subtaskId && t.parentId === taskId);
                
                isFullTask = !!branchedTask || !!(subtaskItem && (subtaskItem as any).childrenCount !== undefined);
                
                return tasks.map((t: Task) => {
                    // Update the subtask inside the parent
                    if (t.id === taskId) {
                        const updatedSubtasks = t.subtasks.map((s: Subtask) => {
                            if (s.id === subtaskId) {
                                currentStatus = !s.done;
                                return { ...s, done: currentStatus };
                            }
                            return s;
                        });
                        return { ...t, subtasks: updatedSubtasks };
                    }
                    // If it's a full task, update it as well
                    if (isFullTask && t.id === subtaskId) {
                        return { ...t, completed: !t.completed, completedAt: !t.completed ? Date.now() : undefined };
                    }
                    return t;
                });
            });

            if (isFullTask) {
                // If it was already completed, we might need a "reopen" or just toggle. 
                // The backend completeTask is a POST to /complete.
                if (currentStatus) {
                    fetch(`${apiUrl}/tasks/${subtaskId}/complete`, { 
                        method: 'POST',
                        headers: { ...getAuthHeader() }
                    });
                } else {
                    // Patch to clear completed_at or just update
                    fetch(`${apiUrl}/tasks/${subtaskId}`, { 
                        method: 'PATCH',
                        headers: { 
                            'Content-Type': 'application/json',
                            ...getAuthHeader()
                        },
                        body: JSON.stringify({ completed_at: null }) 
                    });
                }
            } else {
                fetch(`${apiUrl}/subtasks/${subtaskId}/toggle`, { 
                    method: 'PATCH',
                    headers: { ...getAuthHeader() }
                });
            }
        },
        editSubtask: async (taskId: string, subtaskId: string, title: string) => {
            const apiUrl = getApiUrl();
            update((tasks: Task[]) => tasks.map((t: Task) =>
                t.id === taskId ? {
                    ...t,
                    subtasks: t.subtasks.map((s: Subtask) => s.id === subtaskId ? { ...s, title } : s)
                } : t
            ));
            fetch(`${apiUrl}/subtasks/${subtaskId}`, {
                method: 'PATCH',
                headers: { 
                    'Content-Type': 'application/json',
                    ...getAuthHeader()
                },
                body: JSON.stringify({ title })
            });
        },
        removeSubtask: async (taskId: string, subtaskId: string) => {
            const apiUrl = getApiUrl();
            let isFullTask = false;
            
            update((tasks: Task[]) => {
                const parent = tasks.find(t => t.id === taskId);
                const subtaskItem = parent?.subtasks.find(s => s.id === subtaskId);
                const branchedTask = tasks.find(t => t.id === subtaskId && t.parentId === taskId);
                
                isFullTask = !!branchedTask || !!(subtaskItem && (subtaskItem as any).childrenCount !== undefined);

                const updated = tasks.map(t =>
                    t.id === taskId ? {
                        ...t,
                        subtasks: t.subtasks.filter(s => s.id !== subtaskId)
                    } : t
                );
                
                if (isFullTask) {
                    return updated.filter(t => t.id !== subtaskId);
                }
                return updated;
            });

            if (isFullTask) {
                // It's a full task acting as a subtask
                fetch(`${apiUrl}/tasks/${subtaskId}`, { 
                    method: 'DELETE',
                    headers: { ...getAuthHeader() }
                });
            } else {
                fetch(`${apiUrl}/subtasks/${subtaskId}`, { 
                    method: 'DELETE',
                    headers: { ...getAuthHeader() }
                });
            }
        },
        addFile: async (taskId: string, fileName: string) => {
            const apiUrl = getApiUrl();
            let fullTask: Task | undefined;
            update((tasks: Task[]) => tasks.map((t: Task) => {
                if (t.id === taskId) {
                    fullTask = { ...t, files: [...(t.files || []), fileName] };
                    return fullTask;
                }
                return t;
            }));
            if (fullTask) {
                fetch(`${apiUrl}/tasks/${taskId}`, {
                    method: 'PATCH',
                    headers: { 
                        'Content-Type': 'application/json',
                        ...getAuthHeader()
                    },
                    body: JSON.stringify({
                        title: fullTask.title,
                        text: fullTask.text,
                        x: fullTask.x,
                        y: fullTask.y,
                        files: fullTask.files,
                        modules: fullTask.modules,
                        parent_id: (fullTask.parentId && fullTask.parentId !== "") ? fullTask.parentId : null,
                        meetings: fullTask.meetings
                    })
                }).catch(e => console.error("Failed to add file", e));
            }
        },
        createChildTask: async (parentTask: Task, title: string): Promise<Task | null> => {
            const apiUrl = getApiUrl();
            try {
                const res = await fetch(`${apiUrl}/tasks`, {
                    method: 'POST',
                    headers: { 
                        'Content-Type': 'application/json',
                        ...getAuthHeader()
                    },
                    body: JSON.stringify({
                        title,
                        template: 'task',
                        dimension: parentTask.dimension,
                        parent_id: parentTask.id,
                        x: 0,
                        y: 0,
                        text: '',
                        files: [],
                        modules: CHILD_MODULES,
                        meetings: []
                    })
                });
                if (res.ok) {
                    const created = await res.json();
                    created.parentId = (created.parent_id && created.parent_id !== "") ? created.parent_id : undefined;
                    created.subtasks = created.subtasks || [];
                    created.meetings = created.meetings || [];
                    created.modules = (created.modules && created.modules.length > 0) ? created.modules : CHILD_MODULES;
                    created.completed = !!created.completed_at;
                    created.childrenCount = 0;
                    // Add to store and update parent's childrenCount
                    update((tasks: Task[]) => {
                        const newTasks = tasks.map((t: Task) =>
                            t.id === parentTask.id ? { ...t, childrenCount: t.childrenCount + 1 } : t
                        );
                        
                        const idx = newTasks.findIndex(t => t.id === created.id);
                        if (idx !== -1) {
                            newTasks[idx] = { ...newTasks[idx], ...created };
                            return newTasks;
                        }
                        return [created, ...newTasks];
                    });
                    return created;
                }
            } catch (e) {
                console.error("Failed to create child task", e);
            }
            return null;
        },
        getChildren: async (parentId: string): Promise<Task[]> => {
            const apiUrl = getApiUrl();
            try {
                const res = await fetch(`${apiUrl}/tasks/${parentId}/children`, {
                    headers: { ...getAuthHeader() }
                });
                if (res.ok) {
                    const data = await res.json();
                        const children = (data || []).map((t: any) => ({
                            ...t,
                            parentId: (t.parent_id && t.parent_id !== "") ? t.parent_id : undefined,
                            subtasks: t.subtasks || [],
                            meetings: t.meetings || [],
                            modules: (t.modules && t.modules.length > 0) ? t.modules : CHILD_MODULES,
                            completed: !!t.completed_at,
                            childrenCount: t.children_count || 0
                        }));

                    update((tasks: Task[]) => {
                        const newTasks = [...tasks];
                        for (const child of children) {
                            // Replace element if it exists in store already (e.g., partial syncs), 
                            // otherwise push to the store.
                            const existingIndex = newTasks.findIndex((t: Task) => t.id === child.id);
                            if (existingIndex !== -1) {
                                newTasks[existingIndex] = { ...newTasks[existingIndex], ...child };
                            } else {
                                newTasks.push(child);
                            }
                        }
                        return newTasks;
                    });

                    return children;
                }
            } catch (e) {
                console.error("Failed to get children", e);
            }
            return [];
        },
    };
}

export const tasks = createTaskStore();

export function aggregateSubtasks($tasks: Task[], root: Task) {
    if (!root || !root.id) return [];
    
    // Collect children from the top-level tasks table that point to this root
    const childrenFromTasks = ($tasks || [])
        .filter(t => t && t.parentId === root.id);

    const finalSubtasks: any[] = [];
    const seenTitles = new Set<string>();
    const seenIds = new Set<string>();

    // 1. Real task children first (they take precedence)
    childrenFromTasks.forEach(child => {
        if (!child) return;
        const title = (child.title || '').toLowerCase().trim();
        if (!seenIds.has(child.id) && !seenTitles.has(title)) {
            finalSubtasks.push({ 
                id: child.id, 
                title: child.title, 
                done: child.completed, 
                childrenCount: child.childrenCount || 0 
            });
            seenIds.add(child.id);
            seenTitles.add(title);
        }
    });

    // 2. JSON checklist items
    (root.subtasks || []).forEach(s => {
        if (!s) return;
        const subId = s.id;
        const title = (s.title || '').toLowerCase().trim();
        if (!seenIds.has(subId) && !seenTitles.has(title)) {
            finalSubtasks.push({
                id: subId,
                title: s.title,
                done: s.done,
                childrenCount: s.childrenCount || 0
            });
            seenIds.add(subId);
            seenTitles.add(title);
        }
    });

    return finalSubtasks.filter(s => s.id !== root.id && (s.title || '').toLowerCase().trim() !== (root.title || '').toLowerCase().trim());
}

export const workTasks = derived(tasks, $tasks => {
    const roots = $tasks.filter(t => t.dimension === 'work' && !t.parentId);
    return roots.map(root => ({
        ...root,
        subtasks: aggregateSubtasks($tasks, root)
    }));
});

export const personalTasks = derived(tasks, $tasks => {
    const roots = $tasks.filter(t => t.dimension === 'personal' && !t.parentId);
    return roots.map(root => ({
        ...root,
        subtasks: aggregateSubtasks($tasks, root)
    }));
});
