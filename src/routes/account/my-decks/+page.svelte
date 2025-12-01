<script lang="ts">
	import { Button, Select, Input, Label, Gallery } from 'flowbite-svelte';
	import { enhance } from '$app/forms';
	import type { PageProps } from './$types';
	let { data, form }: PageProps = $props();

	let selected = '';
</script>

<!-- Filter query: -->
<div class="mx-auto mb-3 flex flex-wrap items-center justify-center gap-3 py-4 md:py-8">
	{#each data.collections as collection (collection.id)}
		<div class="relative">
			<form action="?/getCollectionCards" method="POST" use:enhance>
				<input type="hidden" name="collection_id" value={collection.id} />
				<Button type="submit" pill size="xl" outline>{collection.name}</Button>
			</form>
		</div>
	{/each}

	<Button pill size="xl" outline>All my collections</Button>
</div>

<!-- Filter query results: -->
{#if form?.context === 'getCollectionCards' && form?.result}
	{#each form.result as collectioncard (collectioncard.card_id)}
		<p>{collectioncard.card_id}</p>
		<p>{collectioncard.card_name}</p>
		<p>{collectioncard.card_quantity}</p>
	{/each}
{/if}

<!-- Export collection (deck): -->
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
