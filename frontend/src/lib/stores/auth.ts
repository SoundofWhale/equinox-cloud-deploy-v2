import { writable, get } from 'svelte/store';
import { container } from '../services';

export interface User {
    id: string;
    email: string;
}

export interface AuthState {
    user: User | null;
    token: string | null;
    isAuthenticated: boolean;
    loading: boolean;
    error: string | null;
}

const initialState: AuthState = {
    user: null,
    token: null,
    isAuthenticated: false,
    loading: false,
    error: null
};

function createAuthStore() {
    const { subscribe, set, update } = writable<AuthState>(initialState);

    const getApiUrl = () => {
        try {
            return container.getService<string>('API_URL');
        } catch (e) {
            return '/api/v1';
        }
    };

    return {
        subscribe,
        init: () => {
            const token = localStorage.getItem('equinox_token');
            const userStr = localStorage.getItem('equinox_user');
            if (token && userStr) {
                try {
                    const user = JSON.parse(userStr);
                    set({
                        user,
                        token,
                        isAuthenticated: true,
                        loading: false,
                        error: null
                    });
                } catch (e) {
                    localStorage.removeItem('equinox_token');
                    localStorage.removeItem('equinox_user');
                }
            }
        },
        login: async (email: string, password: string) => {
            const apiUrl = getApiUrl();
            update(s => ({ ...s, loading: true, error: null }));
            try {
                const res = await fetch(`${apiUrl}/auth/login`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ email, password })
                });

                const bodyText = await res.text();
                let data: any = {};
                try {
                    data = JSON.parse(bodyText);
                } catch (e) {
                    // Not JSON
                }

                if (!res.ok) {
                    const contentType = res.headers.get('content-type');
                    if (contentType && contentType.includes('application/json')) {
                        throw new Error(data.error || bodyText || 'Login failed');
                    } else {
                        throw new Error('Server error: Unable to process login request');
                    }
                }

                const { token, user_id } = data;
                const user = { id: user_id, email };

                localStorage.setItem('equinox_token', token);
                localStorage.setItem('equinox_user', JSON.stringify(user));

                set({
                    user,
                    token,
                    isAuthenticated: true,
                    loading: false,
                    error: null
                });
                return true;
            } catch (e: any) {
                update(s => ({ ...s, loading: false, error: e.message }));
                return false;
            }
        },
        register: async (email: string, password: string) => {
            const apiUrl = getApiUrl();
            update(s => ({ ...s, loading: true, error: null }));
            try {
                const res = await fetch(`${apiUrl}/auth/register`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ email, password })
                });

                const bodyText = await res.text();
                let data: any = {};
                try {
                    data = JSON.parse(bodyText);
                } catch (e) {
                    // Not JSON
                }

                if (!res.ok) {
                    const contentType = res.headers.get('content-type');
                    if (contentType && contentType.includes('application/json')) {
                        throw new Error(data.error || bodyText || 'Registration failed');
                    } else {
                        throw new Error('Server error: Unable to process registration request');
                    }
                }

                update(s => ({ ...s, loading: false, error: null }));
                return true;
            } catch (e: any) {
                update(s => ({ ...s, loading: false, error: e.message }));
                return false;
            }
        },
        logout: () => {
            localStorage.removeItem('equinox_token');
            localStorage.removeItem('equinox_user');
            set(initialState);
        }
    };
}

export const auth = createAuthStore();
