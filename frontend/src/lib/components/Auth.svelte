<script lang="ts">
    import { auth } from '../stores/auth';
    import { tasks } from '../stores/tasks';
    import { onMount } from 'svelte';
    import { fade, fly, scale } from 'svelte/transition';
    import { quintOut } from 'svelte/easing';

    let isLogin = true;
    let email = '';
    let password = '';
    let error = '';
    let loading = false;

    $: error = $auth.error || '';
    $: loading = $auth.loading;

    async function handleSubmit() {
        if (!email || !password) {
            error = 'Please fill in all fields';
            return;
        }

        let success = false;
        if (isLogin) {
            success = await auth.login(email, password);
        } else {
            const registrationSuccess = await auth.register(email, password);
			if (registrationSuccess) {
                isLogin = true; // Set mode to 'login'
                error = 'Registration successful! Please login.';
                return;
            }
            success = registrationSuccess; // Assign to the outer 'success' for consistency if needed later
        }

        if (success && isLogin) {
            tasks.init();
        }
    }

    function toggleMode() {
        isLogin = !isLogin;
        error = '';
    }
</script>

<div class="auth-wrapper" in:fade={{ duration: 800 }}>
    <div class="background-elements">
        <div class="glow-orb orb-1"></div>
        <div class="glow-orb orb-2"></div>
    </div>

    <div class="auth-container" in:fly={{ y: 20, duration: 1000, easing: quintOut, delay: 200 }}>
        <div class="auth-card">
            <div class="logo-section" in:scale={{ start: 0.9, duration: 800, delay: 400 }}>
                <div class="logo">
                    <span class="equi">EQUI</span><span class="nox">NOX</span>
                </div>
                <div class="logo-glow"></div>
            </div>

            <div class="form-header">
                <h2>{isLogin ? 'Welcome Back' : 'Join the Orbit'}</h2>
                <p class="subtitle">{isLogin ? 'Enter your credentials to access your dimensions' : 'Create an account to start organizing your universe'}</p>
            </div>

            <form on:submit|preventDefault={handleSubmit} class="auth-form">
                <div class="input-group">
                    <label for="email">Email</label>
                    <div class="input-wrapper">
                        <input 
                            type="email" 
                            id="email" 
                            bind:value={email} 
                            placeholder="voyager@equinox.com"
                            autocomplete="email"
                        />
                        <div class="input-focus-border"></div>
                    </div>
                </div>

                <div class="input-group">
                    <label for="password">Password</label>
                    <div class="input-wrapper">
                        <input 
                            type="password" 
                            id="password" 
                            bind:value={password} 
                            placeholder="Secure key"
                            autocomplete="current-password"
                        />
                        <div class="input-focus-border"></div>
                    </div>
                </div>

                {#if error}
                    <div class="error-box" transition:fly={{ y: -10, duration: 300 }}>
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
                        <span>{error}</span>
                    </div>
                {/if}

                <button type="submit" class="auth-submit" disabled={loading}>
                    {#if loading}
                        <div class="loading-state">
                            <span class="spinner"></span>
                            <span>Authenticating...</span>
                        </div>
                    {:else}
                        <span class="btn-text">{isLogin ? 'Initialize Session' : 'Create Identity'}</span>
                        <div class="btn-glow"></div>
                    {/if}
                </button>
            </form>

            <div class="auth-footer">
                <p>
                    {isLogin ? "New to Equinox?" : "Already a voyager?"}
                    <button class="toggle-btn" on:click={toggleMode}>
                        {isLogin ? 'Sign Up' : 'Sign In'}
                    </button>
                </p>
            </div>
        </div>
    </div>
</div>

<style>
    :global(body) {
        margin: 0;
        overflow: hidden;
    }

    .auth-wrapper {
        position: fixed;
        inset: 0;
        display: flex;
        align-items: center;
        justify-content: center;
        background: #050508;
        z-index: 10000;
        font-family: 'Inter', sans-serif;
    }

    .background-elements {
        position: absolute;
        inset: 0;
        overflow: hidden;
        pointer-events: none;
    }

    .glow-orb {
        position: absolute;
        border-radius: 50%;
        filter: blur(80px);
        opacity: 0.15;
        animation: orb-float 20s infinite alternate ease-in-out;
    }

    .orb-1 {
        width: 400px;
        height: 400px;
        background: #4f46e5;
        top: -100px;
        right: -100px;
        pointer-events: none;
    }

    .orb-2 {
        width: 300px;
        height: 300px;
        background: #7c3aed;
        bottom: -50px;
        left: -50px;
        animation-delay: -5s;
        pointer-events: none;
    }

    @keyframes orb-float {
        from { transform: translate(0, 0) scale(1); }
        to { transform: translate(30px, 40px) scale(1.1); }
    }

    .auth-container {
        width: 100%;
        max-width: 440px;
        padding: 20px;
        perspective: 1000px;
    }

    .auth-card {
        background: rgba(15, 15, 25, 0.7);
        backdrop-filter: blur(20px) saturate(180%);
        border: 1px solid rgba(255, 255, 255, 0.08);
        border-radius: 28px;
        padding: 3rem;
        box-shadow: 
            0 25px 50px -12px rgba(0, 0, 0, 0.5),
            inset 0 1px 1px rgba(255, 255, 255, 0.1);
        position: relative;
        overflow: hidden;
    }

    .logo-section {
        margin-bottom: 2.5rem;
        position: relative;
    }

    .logo {
        font-size: 2.5rem;
        font-weight: 850;
        letter-spacing: -2px;
        position: relative;
        z-index: 1;
    }

    .equi { color: #fff; }
    .nox { 
        color: #6366f1; 
        background: linear-gradient(135deg, #6366f1 0%, #a855f7 100%);
        -webkit-background-clip: text;
        background-clip: text;
        -webkit-text-fill-color: transparent;
    }

    .logo-glow {
        position: absolute;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        width: 100px;
        height: 40px;
        background: #6366f1;
        filter: blur(40px);
        opacity: 0.3;
        pointer-events: none;
    }

    .form-header {
        text-align: left;
        margin-bottom: 2rem;
    }

    h2 {
        color: #fff;
        font-size: 1.75rem;
        font-weight: 700;
        margin: 0 0 0.5rem 0;
        letter-spacing: -0.5px;
    }

    .subtitle {
        color: rgba(255, 255, 255, 0.4);
        font-size: 0.95rem;
        line-height: 1.5;
        margin: 0;
    }

    .input-group {
        margin-bottom: 1.5rem;
    }

    label {
        display: block;
        color: rgba(255, 255, 255, 0.6);
        font-size: 0.8rem;
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 1px;
        margin-bottom: 0.75rem;
        padding-left: 4px;
    }

    .input-wrapper {
        position: relative;
    }

    input {
        width: 100%;
        padding: 1rem 1.25rem;
        background: rgba(255, 255, 255, 0.03);
        border: 1px solid rgba(255, 255, 255, 0.05);
        border-radius: 14px;
        color: #fff;
        font-size: 1rem;
        transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
        box-sizing: border-box;
    }

    input::placeholder {
        color: rgba(255, 255, 255, 0.15);
    }

    input:focus {
        outline: none;
        background: rgba(255, 255, 255, 0.06);
        border-color: rgba(99, 102, 241, 0.5);
    }

    .input-focus-border {
        position: absolute;
        inset: -1px;
        border-radius: 14px;
        border: 2px solid #6366f1;
        opacity: 0;
        pointer-events: none;
        transition: opacity 0.3s, transform 0.3s;
        transform: scale(0.98);
    }

    input:focus + .input-focus-border {
        opacity: 1;
        transform: scale(1);
    }

    .error-box {
        display: flex;
        align-items: center;
        gap: 0.75rem;
        background: rgba(239, 68, 68, 0.08);
        border: 1px solid rgba(239, 68, 68, 0.2);
        color: #fca5a5;
        padding: 0.875rem 1rem;
        border-radius: 12px;
        font-size: 0.9rem;
        margin-bottom: 1.5rem;
    }

    .error-box svg {
        width: 18px;
        height: 18px;
        flex-shrink: 0;
    }

    .auth-submit {
        width: 100%;
        padding: 1.1rem;
        background: #6366f1;
        color: #fff;
        border: none;
        border-radius: 14px;
        font-size: 1rem;
        font-weight: 700;
        cursor: pointer;
        transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
        position: relative;
        overflow: hidden;
        margin-top: 0.5rem;
    }

    .auth-submit:hover:not(:disabled) {
        transform: translateY(-2px);
        box-shadow: 0 10px 20px -5px rgba(99, 102, 241, 0.5);
    }

    .auth-submit:active {
        transform: translateY(0);
    }

    .auth-submit:disabled {
        opacity: 0.6;
        cursor: not-allowed;
        filter: grayscale(0.5);
    }

    .btn-text {
        position: relative;
        z-index: 2;
    }

    .btn-glow {
        position: absolute;
        inset: 0;
        background: linear-gradient(90deg, transparent, rgba(255,255,255,0.2), transparent);
        transform: translateX(-100%);
        transition: transform 0.6s;
    }

    .auth-submit:hover .btn-glow {
        transform: translateX(100%);
    }

    .loading-state {
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 0.75rem;
    }

    .spinner {
        width: 1.25rem;
        height: 1.25rem;
        border: 2px solid rgba(255, 255, 255, 0.3);
        border-radius: 50%;
        border-top-color: #fff;
        animation: spin 0.8s linear infinite;
    }

    @keyframes spin {
        to { transform: rotate(360deg); }
    }

    .auth-footer {
        margin-top: 2.5rem;
        color: rgba(255, 255, 255, 0.4);
        font-size: 0.95rem;
    }

    .auth-footer p {
        margin: 0;
    }

    .toggle-btn {
        background: none;
        border: none;
        color: #818cf8;
        font-weight: 700;
        cursor: pointer;
        padding: 0.25rem 0.5rem;
        transition: all 0.2s;
        border-radius: 6px;
    }

    .toggle-btn:hover {
        color: #fff;
        background: rgba(129, 140, 248, 0.1);
    }

    @media (max-width: 480px) {
        .auth-card {
            padding: 2rem;
            border-radius: 20px;
        }

        h2 {
            font-size: 1.5rem;
        }
    }
</style>
