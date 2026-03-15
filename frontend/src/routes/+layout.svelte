<script lang="ts">
	import "../app.css";
	import {
		activeDimension,
		isGatewayActive,
		switchDimension,
		getRandomQuote,
		type DimensionType,
	} from "$lib/stores/dimension";
	import Gateway from "$lib/components/Gateway.svelte";
	import Mascot from "$lib/components/Mascot.svelte";
	import QuickAction from "$lib/components/QuickAction.svelte";
	import EmergencyButton from "$lib/components/EmergencyButton.svelte";
	import { isEmergencyMode } from "$lib/stores/emergency";
	import { page } from "$app/stores";
	import { goto } from "$app/navigation";
	import { onMount } from "svelte";
	import { auth } from "$lib/stores/auth";
	import { tasks } from "$lib/stores/tasks";
	import { initializeBaseServices } from "$lib/services";
	import Auth from "$lib/components/Auth.svelte";

	onMount(() => {
		console.log("[Layout] onMount triggered");
		initializeBaseServices();
		auth.init();
		if ($auth.isAuthenticated) {
			console.log("[Layout] Authenticated -> calling tasks.init()");
			tasks.init();
		}
	});

	function handleSwitch(target: DimensionType) {
		// Don't switch if already on this dimension
		if ($page.url.pathname === `/${target}`) return;

		switchDimension(target);

		// Navigate after gateway closes
		setTimeout(() => {
			goto(`/${target}`);
		}, 3000);
	}

	// Determine which nav link is active
	$: currentPath = $page.url.pathname;

	function handleLogout() {
		auth.logout();
		goto("/");
	}
</script>

<!-- Navigation Bar -->
<nav class="nav-bar">
	{#if $auth.isAuthenticated}
		<button class="logout-btn" on:click={handleLogout} title="Logout">
			<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path><polyline points="16 17 21 12 16 7"></polyline><line x1="21" y1="12" x2="9" y2="12"></line></svg>
		</button>
	{/if}
	<div class="nav-links">
		{#if $auth.isAuthenticated}
			<a
				href="/work"
				class:active={currentPath === "/work" || currentPath === "/"}
				on:click|preventDefault={() => handleSwitch("work")}
			>
				🪐 Work
			</a>
			<a
				href="/personal"
				class:active-personal={currentPath === "/personal"}
				on:click|preventDefault={() => handleSwitch("personal")}
			>
				🌿 Personal
			</a>
			<a
				href="/cloud"
				class:active={currentPath === "/cloud"}
				on:click|preventDefault={() => handleSwitch("cloud")}
			>
				☁️ Cloud
			</a>
		{/if}
	</div>
</nav>

<!-- Gateway Transition Overlay -->
{#if $isGatewayActive}
	<Gateway on:complete={() => ($isGatewayActive = false)} />
{/if}

{#if !$auth.isAuthenticated}
	<Auth />
{/if}

<!-- Mascot Overlay -->
<Mascot />

<!-- Quick Action FAB -->
<QuickAction />

<!-- Emergency Mode Button -->
<EmergencyButton />

<!-- Emergency Visual Overlay -->
{#if $isEmergencyMode}
	<div class="emergency-overlay">
		<div class="emergency-text">
			<h1>EMERGENCY MODE ACTIVATED</h1>
			<p>
				Only show ultra-critical tasks. Non-essential processes paused.
			</p>
			<button
				class="disable-emergency-btn"
				on:click={() => ($isEmergencyMode = false)}>Deactivate</button
			>
		</div>
	</div>
{/if}

<!-- Page Content -->
<slot />

<style>
	.emergency-overlay {
		position: fixed;
		inset: 0;
		background: rgba(40, 0, 0, 0.9);
		border: 10px solid #ff2a2a;
		z-index: 100;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		animation: pulse-border 2s infinite alternate;
		pointer-events: auto; /* blocks clicks to background tasks */
	}

	.emergency-text {
		text-align: center;
		color: #ff5252;
		background: rgba(0, 0, 0, 0.7);
		padding: 3rem;
		border-radius: 2rem;
		border: 1px solid #ff5252;
	}

	.emergency-text h1 {
		font-size: 3rem;
		margin: 0 0 1rem 0;
		text-transform: uppercase;
		letter-spacing: 2px;
	}

	.emergency-text p {
		font-size: 1.2rem;
		color: #ff8a80;
	}

	.disable-emergency-btn {
		margin-top: 2rem;
		padding: 1rem 2rem;
		font-size: 1.2rem;
		font-weight: bold;
		background: transparent;
		color: #ff5252;
		border: 2px solid #ff5252;
		border-radius: 1rem;
		cursor: pointer;
		transition: all 0.2s;
	}

	.disable-emergency-btn:hover {
		background: #ff5252;
		color: black;
	}

	@keyframes pulse-border {
		0% {
			border-color: #ff2a2a;
			background: rgba(50, 0, 0, 0.9);
		}
		100% {
			border-color: #ff8a80;
			background: rgba(30, 0, 0, 0.95);
		}
	}
</style>
