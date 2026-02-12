<script lang="ts">
	import favicon from '$lib/assets/favicon.svg';
	import "../app.css";
	let { children, data } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>
{#if data.userEmail}
  <p class="text-green-500 mt-4">Welcome, {data.userEmail}.</p>
{/if}
{@render children()}
