<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import Navbar from '$lib/custom-components/Navbar.svelte';

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>
<Navbar />
{@render children()}

<!-- <style lang="postcss">
	@reference "tailwindcss";

	:global(html) {
		background-color: theme(--color-gray-100);
	}
</style> -->
