<script lang="ts">
	import { Button, Select, Input, Label, Gallery } from 'flowbite-svelte';
	import { enhance } from '$app/forms';
	import type { PageProps } from './$types';
	let { data, form }: PageProps = $props();

	let selected = '';
</script>

<div class="mx-auto mb-3 flex flex-wrap items-center justify-center gap-3 py-4 md:py-8">
	<Button pill size="xl" outline>All cards</Button>
	<Button pill size="xl" color="alternative">deck 1</Button>
	<Button pill size="xl" color="alternative">deck 2</Button>
</div>

<!-- boucler sur cards -->
<div class="grid grid-cols-2 gap-4 p-8 md:grid-cols-4 lg:grid-cols-6">
	{#each data.cards as card (card.id)}
		<div class="relative">
			<img src={card.image_uri} alt={card.name} class="h-auto w-full" />
			<!-- {card.name} -->
			<form action="?/addToCollection" method="POST" use:enhance>
				<div class="flex">
					<Select id="quantity-{card.id}" name="quantity" required>
						{#each Array.from({ length: card.quantity ?? 0 }, (_, i) => i + 1) as qty (qty)}
							<option value={qty}>{qty}</option>
						{/each}
					</Select>
					<Select name="collection_id" required>
						{#each data.collections as collection (collection.id)}
							<option value={collection.id}>{collection.name}</option>
						{/each}
					</Select>
				</div>
				<Input type="hidden" name="card_id" value={card.id} />

				<Button color="secondary" class="mt-2" type="submit">Add to deck</Button>
			</form>
		</div>
	{/each}
</div>

{#each data.collections as collection (collection.id)}
	<form action="?/export" method="POST" class="space-y-6" use:enhance>
		<Input type="hidden" name="collection_id" value={collection.id} />
		<Button color="alternative" type="submit">Export {collection.name} to XML</Button>
	</form>
{/each}

{#if form?.context === 'export' && form?.xml}
	<a href={`data:text/xml;charset=utf-8,${encodeURIComponent(form.xml)}`} download="collection.cod">
		<Button color="dark">Télécharger le fichier XML</Button>
	</a>
{/if}
