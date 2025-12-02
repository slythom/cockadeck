<script lang="ts">
	import { Button } from 'flowbite-svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { enhance } from '$app/forms';
	import type { PageProps } from './$types';
	let { data, form }: PageProps = $props();

	// Lire le filtre depuis l'URL
	let activeCollectionId = $derived($page.url.searchParams.get('collection'));

	// Filtrer les cartes de collection
	let displayedCards = $derived(
		activeCollectionId
			? data.collectionCards.filter((c) => c.collection_id === activeCollectionId)
			: data.cards // Toutes les cartes si pas de filtre
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
		<input name="collection_id" type="hidden" value={activeCollectionId} />
		<Button color="teal" type="submit">Export (COD)</Button>
	</form>
{/if}
{#if form?.context === 'export' && form?.xml}
	<a href={`data:text/xml;charset=utf-8,${encodeURIComponent(form.xml)}`} download="collection.cod">
		<Button color="red" type="button">Télécharger le fichier XML</Button>
	</a>
{/if}

<!-- Affichage -->
<div class="grid grid-cols-2 gap-4 p-8 md:grid-cols-4 lg:grid-cols-6">
	{#each displayedCards as card (card.id || card.card_id)}
		<div class="relative">
			<img src={card.image_uri} alt={card.card_name || card.name} />
			{card.quantity}
		</div>
	{/each}
</div>
