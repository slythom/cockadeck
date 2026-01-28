<script lang="ts">
	import { Button } from 'flowbite-svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { enhance } from '$app/forms';
	import type { PageProps } from './$types';
	let { data, form }: PageProps = $props();

	// Lire le filtre depuis l'URL
	let activeDeckId = $derived(page.url.searchParams.get('deck'));

	// Filtrer les cartes de deck
	let filteredDeckCards = $derived(
		data.deckCards.filter((c) => c.deck_id === activeDeckId)
	);
</script>

<!-- Boutons -->
<div class="mx-auto mb-3 flex flex-wrap items-center justify-center gap-3 py-4">
	<Button onclick={() => goto('?')}>All Cards</Button>

	{#each data.decks as deck (deck.id)}
		<Button onclick={() => goto(`?deck=${deck.id}`)}>
			{deck.name}
		</Button>
	{/each}
</div>

{#if activeDeckId}
	<form action="?/export" method="POST" use:enhance>
		<div class="p-8">
			<input name="deck_id" type="hidden" value={activeDeckId} />
			<Button color="teal" type="submit">Export for Cockatrice</Button>
		</div>
	</form>
	<div class="grid grid-cols-2 gap-4 p-8 md:grid-cols-4 lg:grid-cols-6">
		{#each filteredDeckCards as card (card.card_id)}
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
	<div class="grid grid-cols-2 gap-10 p-8 md:grid-cols-4 lg:grid-cols-4">
		{#each data.cards as card (card.id)}
			<div class="relative flex">
				<img src={card.image_uri} alt={card.name} />
				<div
					class="absolute top-5 right-5 h-7 w-10 rounded-sm bg-teal-900 text-center font-black text-white"
				>
					{card.quantity}
				</div>

<div class="absolute top-20">
				<!-- Formulaire d'ajout à deck -->
                <form method="POST" action="?/addToDeck" use:enhance class="mt-2">
                    <input type="hidden" name="card_id" value={card.id} />
                    <input type="hidden" name="quantity" value="1" />
                    
                    <select name="deck_id" required class="w-full text-sm rounded">
                        <option value="" disabled selected>+ deck</option>
                        {#each data.decks as deck}
                            <option value={deck.id}>{deck.name}</option>
                        {/each}
                    </select>
                    
                    <Button size="xs" type="submit" class="w-full mt-1">Ajouter</Button>
                </form>

</div>

			</div>
		{/each}
	</div>
{/if}
{#if form?.context === 'export' && form?.xml}
	<a href={`data:text/xml;charset=utf-8,${encodeURIComponent(form.xml)}`} download="deck.cod">
		<div class="p-8">
			<Button color="blue" type="button" class="cursor-pointer">Download deck!</Button>
		</div>
	</a>
{/if}
