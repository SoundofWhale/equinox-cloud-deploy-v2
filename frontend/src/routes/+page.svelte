<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	// Redirect root to /work
	onMount(() => {
		goto('/work', { replaceState: true });
	});
</script>

<svelte:head>
	<title>EQUINOX 2.0</title>
</svelte:head>
