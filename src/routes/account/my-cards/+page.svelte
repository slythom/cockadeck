<script lang="ts">
	import { Button } from 'flowbite-svelte';
	import { Select } from 'flowbite-svelte';
	import { enhance } from '$app/forms';
	import type { PageProps } from './$types';
	let { data, form }: PageProps = $props();

	let selected = '';
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

			<Button color="alternative" type="submit">add to collec</Button>
		</form>
	{/each}
	{#each data.collections as collection (collection.id)}
		<form action="?/export" method="POST" class="space-y-6" use:enhance>
			<input type="hidden" name="collection_id" value={collection.id} />
			<Button color="alternative" type="submit">Export {collection.name} to XML</Button>
		</form>
	{/each}
</ul>

{#if form?.context === 'export' && form?.xml}
	<a href={`data:text/xml;charset=utf-8,${encodeURIComponent(form.xml)}`} download="collection.cod">
		<Button color="dark">Télécharger le fichier XML</Button>
	</a>
{/if}
