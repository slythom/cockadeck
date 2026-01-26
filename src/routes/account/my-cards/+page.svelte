<script lang="ts">
	import { Button } from 'flowbite-svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { enhance } from '$app/forms';
	import type { PageProps } from './$types';
	let { data, form }: PageProps = $props();

	// Lire le filtre depuis l'URL
	let activeCollectionId = $derived(page.url.searchParams.get('collection'));

	// Filtrer les cartes de collection
	let filteredCollectionCards = $derived(
		data.collectionCards.filter((c) => c.collection_id === activeCollectionId)
	);
</script>

<!-- Boutons -->
<div class="mx-auto mb-3 flex flex-wrap items-center justify-center gap-3 py-4">
	<Button onclick={() => goto('?')}>All Cards</Button>

	{#each data.collections as collection (collection.id)}
		<Button onclick={() => goto(`?collection=${collection.id}`)}>
			{collection.name}
		</Button>
	{/each}
</div>

{#if activeCollectionId}
	<form action="?/export" method="POST" use:enhance>
		<div class="p-8">
			<input name="collection_id" type="hidden" value={activeCollectionId} />
			<Button color="teal" type="submit">Export for Cockatrice</Button>
		</div>
	</form>
	<div class="grid grid-cols-2 gap-4 p-8 md:grid-cols-4 lg:grid-cols-6">
		{#each filteredCollectionCards as card (card.card_id)}
			<div class="relative flex">
				<img src={card.image_uri} alt={card.card_name} />
				<div
					class="absolute top-5 right-5 h-7 w-10 rounded-sm bg-teal-900 text-center font-black text-white"
				>
					{card.quantity}
				</div>
			</div>
		{/each}
	</div>
{:else}
	<div class="grid grid-cols-2 gap-4 p-8 md:grid-cols-4 lg:grid-cols-6">
		{#each data.cards as card (card.id)}
			<div class="relative flex">
				<img src={card.image_uri} alt={card.name} />
				<div
					class="absolute top-5 right-5 h-7 w-10 rounded-sm bg-teal-900 text-center font-black text-white"
				>
					{card.quantity}
				</div>
			</div>
		{/each}
	</div>
{/if}
{#if form?.context === 'export' && form?.xml}
	<a href={`data:text/xml;charset=utf-8,${encodeURIComponent(form.xml)}`} download="collection.cod">
		<div class="p-8">
			<Button color="blue" type="button" class="cursor-pointer">Download deck!</Button>
		</div>
	</a>
{/if}
