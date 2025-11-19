<script lang="ts">
	import type { PageProps } from './$types';
	let { data }: PageProps = $props();
</script>

<!-- boucler sur cards -->
<ul>
	{#each data.cards as card (card.id)}
		<li>{card.name} x {card.quantity}</li>
	{/each}
</ul>
