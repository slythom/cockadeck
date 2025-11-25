<script lang="ts">
	import { Button, Input, Label, Helper, Card } from 'flowbite-svelte';
	import type { PageProps } from './$types';
	import { enhance } from '$app/forms';

	let { data, form }: PageProps = $props();

	let searchResults = $state<any[]>([]);
</script>

<!-- <div class="flex min-h-full flex-col justify-center px-6 py-12 lg:px-8"> -->
<div class="sm:mx-auto sm:w-full sm:max-w-sm">
	<img
		src="https://tailwindcss.com/plus-assets/img/logos/mark.svg?color=indigo&shade=600"
		alt="Your Company"
		class="mx-auto h-10 w-auto"
	/>
	<h2 class="mt-10 text-center text-2xl/9 font-bold tracking-tight text-gray-900">
		Give a name to your collection/deck
		<form action="?/createCollection" method="POST" class="space-y-6" use:enhance>
			<input
				id="name"
				type="text"
				name="name"
				required
				class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600 sm:text-sm/6"
			/>
			<button type="submit">confirm</button>
		</form>
	</h2>

	<h2 class="mt-10 text-center text-2xl/9 font-bold tracking-tight text-gray-900">
		Search your cards
	</h2>
</div>

<div class="mx-auto mt-10">
	<form
		action="?/search"
		method="POST"
		class="mx-auto"
		use:enhance={() => {
			return async ({ result, update }) => {
				// Mettre à jour form normalement
				await update();

				// Ajouter à la liste si succès
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
		<div class="mb-6 grid grid-cols-4 gap-4 p-10">
			<div>
				<div>
					<Label for="setcode" class="mb-2">Set Code</Label>
					<Input id="setcode" type="text" name="setcode" required />
				</div>
			</div>

			<div>
				<Label for="collector_number" class="mb-2">Collector Number</Label>
				<Input id="collector_number" type="text" name="collector_number" required />
			</div>
			<div>
				<div>
					<Label for="quantity" class="mb-2">Quantity</Label>
					<Input id="quantity" type="number" name="quantity" required />
				</div>
			</div>

			<div>
				<Button type="submit" class="mt-7 cursor-pointer">Find!</Button>
			</div>
		</div>
	</form>
</div>
<!-- </div> -->

<!-- Afficher les erreurs -->
{#if form?.context === 'search' && form?.error}
	<div class="mt-4 rounded bg-red-100 p-4 text-red-700">
		{form.error}
	</div>
{/if}

<!-- Afficher les résultats de la recherche -->
<!-- Afficher TOUS les résultats accumulés -->
{#if searchResults.length > 0}
	<div class="mt-6 space-y-4">
		<!-- <h3 class="text-xl font-bold">Résultats de recherche ({searchResults.length})</h3> -->
		<div class="grid grid-cols-4">
			{#each searchResults as result (result.name)}
				<Card class="border-none p-4 shadow-xl">
					<div class="m-6">
						<img src={result.image} alt={result.name} />
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
						<Button class="w-40">Read more</Button>
						<form action="?/save" method="POST" class="mt-4" use:enhance>
							<input type="hidden" name="setcode" value={result.setcode} />
							<input type="hidden" name="collector_number" value={result.collector_number} />
							<input type="hidden" name="quantity" value={result.quantity} />
							<input type="hidden" name="name" value={result.name} />
							<input type="hidden" name="image_uri" value={result.image} />
							<input type="hidden" name="colors" value={result.colors} />
							<input type="hidden" name="mana_cost" value={result.mana_cost} />
							<input type="hidden" name="setname" value={result.set_name} />

							<Button type="submit">Save</Button>
						</form>
					</div>
				</Card>

				<!-- <div class="mx-auto mt-4 rounded p-6 text-sm shadow-xl"> -->
				<!-- <p class="font-bold">
						{result.name}
					</p>
					<p>{result.set_name} #{result.collector_number}</p>
					<p>Mana cost: {result.mana_cost || 'N/A'}</p>
					<p>Quantity: {result.quantity}</p> -->
				<!-- <p>Couleur(s): {result.colors?.join(', ') || 'Aucune'}</p> -->
				<!-- <img src={result.image} alt={result.name} /> -->

				<!-- </div> -->
			{/each}
		</div>
	</div>
{/if}

<!-- Message de confirmation après sauvegarde -->
{#if form?.context === 'save' && form?.message}
	<div class="mt-4 rounded bg-green-100 p-4 text-green-700">
		{form.message}
	</div>

	<!-- Afficher encore la carte sauvegardée -->
	<div class="mt-4 rounded bg-gray-100 p-4">
		<p>{form.card_name} a été ajoutée à ta collection !</p>
	</div>
{/if}

{#if form?.context === 'save' && form?.error}
	<div class="mt-4 rounded bg-red-100 p-4 text-red-700">
		{form.error}
	</div>
{/if}
