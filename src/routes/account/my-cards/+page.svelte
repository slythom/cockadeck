<script lang="ts">
	import { enhance } from '$app/forms';
	import type { PageProps } from './$types';
	let { data }: PageProps = $props();
</script>

<!-- boucler sur cards -->
<ul>
	{#each data.cards as card (card.id)}
		<li>{card.name} x {card.quantity}</li>
		<form action="?/addToCollection" method="POST" class="space-y-6" use:enhance>
			<select name="collection_id" required>
				{#each data.collections as collection (collection.id)}
					<option value={collection.id}>{collection.name}</option>
				{/each}
			</select>

			<input type="hidden" name="card_id" value={card.id} />
			<label for="quantity-{card.id}">Choose a quantity:</label>
			<select id="quantity-{card.id}" name="quantity" required>
				{#each Array.from({ length: card.quantity ?? 0 }, (_, i) => i + 1) as qty (qty)}
					<option value={qty}>{qty}</option>
				{/each}
			</select>
			<button type="submit">add to collec</button>
		</form>
	{/each}
</ul>
