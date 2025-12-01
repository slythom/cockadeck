<script lang="ts">
	import { Badge, Button, Input, Label, Helper, Card, Toast } from 'flowbite-svelte';
	import {
		CheckCircleSolid,
		ExclamationCircleSolid,
		FireOutline,
		CloseCircleSolid
	} from 'flowbite-svelte-icons';
	import type { PageProps } from './$types';
	import { enhance } from '$app/forms';

	let { data, form }: PageProps = $props();

	let searchResults = $state<any[]>([]);
</script>

<div class="flex w-full gap-4 p-6">
	<Card class="max-w-none flex-1 p-4">
		<h2 class="mb-4 text-xl font-bold text-slate-300">Create a deck</h2>
		<form action="?/createCollection" method="POST" use:enhance>
			<div class="mb-6 grid grid-cols-[1fr_auto] gap-4">
				<div>
					<Label for="name" class="mb-2">Choose a name:</Label>
					<Input id="name" type="text" name="name" required />
				</div>
				<div>
					<Button type="submit" class="mt-7 cursor-pointer">Create</Button>
				</div>
			</div>
		</form>
	</Card>

	<Card class="max-w-none flex-1 p-4">
		<h2 class="mb-4 text-xl font-bold text-slate-300">Search your cards</h2>
		<div class="mx-auto">
			<form
				action="?/search"
				method="POST"
				class="mx-auto"
				use:enhance={() => {
					return async ({ result, update }) => {
						await update();
						if (
							result.type === 'success' &&
							result.data?.context === 'search' &&
							result.data?.card_name
						) {
							searchResults = [
								{
									name: result.data.card_name,
									setcode: result.data.card_setcode,
									collector_number: result.data.card_collector_number,
									set_name: result.data.card_set_name,
									image: result.data.card_image_normal,
									colors: result.data.card_colors,
									mana_cost: result.data.card_mana_cost,
									quantity: result.data.card_quantity
								},
								...searchResults
							];
						}
					};
				}}
			>
				<div class="mb-6 grid grid-cols-[1fr_1fr_1fr_auto] gap-4">
					<div>
						<Label for="setcode" class="mb-2">Set Code</Label>
						<Input id="setcode" type="text" name="setcode" required />
					</div>

					<div>
						<Label for="collector_number" class="mb-2">Collector Number</Label>
						<Input id="collector_number" type="text" name="collector_number" required />
					</div>
					<div>
						<Label for="quantity" class="mb-2">Quantity</Label>
						<Input id="quantity" type="number" name="quantity" required />
					</div>
					<div>
						<Button type="submit" class="mt-7 cursor-pointer">Find!</Button>
					</div>
				</div>
			</form>
		</div>
	</Card>
	<Card>how it works?</Card>
</div>

<!-- Afficher les erreurs -->
{#if form?.context === 'search' && form?.error}
	<div class="mt-4 rounded bg-red-100 p-4 text-red-700">
		{form.error}
	</div>
{/if}

<!-- Afficher TOUS les résultats accumulés -->
{#if searchResults.length > 0}
	<div class="mt-6 space-y-4">
		<div class="grid grid-cols-5">
			{#each searchResults as result (result.name)}
				<Card class="border-none p-4 shadow-xl">
					<div class="m-6">
						<img src={result.image} alt={result.name} class="max-w-2/3" />
						<h3 class="mt-2 mb-2 font-bold tracking-tight text-gray-900 dark:text-white">
							{result.name}
						</h3>
						<p class="mb-2 leading-tight font-normal text-gray-700 dark:text-gray-400">
							{result.set_name} #{result.collector_number}
						</p>
						<p class="mb-2 leading-tight font-normal text-gray-700 dark:text-gray-400">
							Mana cost: {result.mana_cost || 'N/A'}
						</p>
						<p class="mb-2 leading-tight font-normal text-gray-700 dark:text-gray-400">
							Quantity: {result.quantity}
						</p>
						<form action="?/save" method="POST" class="mt-4" use:enhance>
							<input type="hidden" name="setcode" value={result.setcode} />
							<input type="hidden" name="collector_number" value={result.collector_number} />
							<input type="hidden" name="quantity" value={result.quantity} />
							<input type="hidden" name="name" value={result.name} />
							<input type="hidden" name="image_uri" value={result.image} />
							<input type="hidden" name="colors" value={result.colors} />
							<input type="hidden" name="mana_cost" value={result.mana_cost} />
							<input type="hidden" name="setname" value={result.set_name} />

							<Button type="submit" class="cursor-pointer">Save to database!</Button>
						</form>
					</div>
				</Card>
			{/each}
		</div>
	</div>
{/if}

<!-- Message de confirmation après sauvegarde -->
{#if form?.context === 'save' && form?.message}
	<Toast color="green" position="top-right">
		{#snippet icon()}
			<CheckCircleSolid class="h-5 w-5" />
			<span class="sr-only">Check icon</span>
		{/snippet}
		<p class="font-bold">{form.card_name}</p>
		has been saved!
	</Toast>
{/if}

{#if form?.context === 'save' && form?.error}
	<div class="mt-4 rounded bg-red-100 p-4 text-red-700">
		{form.error}
	</div>
{/if}
