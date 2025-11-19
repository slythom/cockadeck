<script lang="ts">
	import type { PageProps } from './$types';
	import { enhance } from '$app/forms';

	let { data, form }: PageProps = $props();
</script>

<div class="flex min-h-full flex-col justify-center px-6 py-12 lg:px-8">
	<div class="sm:mx-auto sm:w-full sm:max-w-sm">
		<img
			src="https://tailwindcss.com/plus-assets/img/logos/mark.svg?color=indigo&shade=600"
			alt="Your Company"
			class="mx-auto h-10 w-auto"
		/>
		<h2 class="mt-10 text-center text-2xl/9 font-bold tracking-tight text-gray-900">
			Search your card
		</h2>
	</div>

	<div class="mt-10 sm:mx-auto sm:w-full sm:max-w-sm">
		<form action="?/search" method="POST" class="space-y-6" use:enhance>
			<div>
				<label for="setcode" class="block text-sm/6 font-medium text-gray-900">setcode</label>
				<div class="mt-2">
					<input
						id="setcode"
						type="text"
						name="setcode"
						required
						class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600 sm:text-sm/6"
					/>
				</div>
			</div>

			<div>
				<div class="flex items-center justify-between">
					<label for="collector_number" class="block text-sm/6 font-medium text-gray-900"
						>collector_number</label
					>
				</div>
				<div class="mt-2">
					<input
						id="collector_number"
						type="text"
						name="collector_number"
						required
						class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600 sm:text-sm/6"
					/>
				</div>
				<div class="flex items-center justify-between">
					<label for="quantity" class="block text-sm/6 font-medium text-gray-900">quantity</label>

					<div class="mt-2">
						<input
							id="quantity"
							type="number"
							name="quantity"
							required
							class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600 sm:text-sm/6"
						/>
					</div>
				</div>
			</div>

			<div>
				<button
					type="submit"
					class="flex w-full justify-center rounded-md bg-indigo-600 px-3 py-1.5 text-sm/6 font-semibold text-white shadow-xs hover:bg-indigo-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
					>Find!</button
				>
			</div>
		</form>
	</div>
</div>

<!-- Afficher les erreurs -->
{#if form?.context === 'search' && form?.error}
	<div class="mt-4 rounded bg-red-100 p-4 text-red-700">
		{form.error}
	</div>
{/if}

<!-- Afficher les résultats de la recherche -->
{#if form?.context === 'search' && form?.card_name}
	<div class="mt-4 rounded bg-gray-100 p-4">
		<p>
			{form.card_name} - {form.card_set_name} ({form.card_setcode}) #{form.card_collector_number}
		</p>
		<p>Quantité: {form.card_quantity}</p>
		<p>Coût: {form.card_mana_cost || 'N/A'}</p>
		<p>Couleurs: {form.card_colors?.join(', ') || 'Aucune'}</p>
		<img src={form.card_image_normal} alt={form.card_name} />
	</div>

	<form action="?/save" method="POST" class="mt-4 space-y-6" use:enhance>
		<input type="hidden" name="setcode" value={form.card_setcode} />
		<input type="hidden" name="collector_number" value={form.card_collector_number} />
		<input type="hidden" name="quantity" value={form.card_quantity} />
		<input type="hidden" name="name" value={form.card_name} />
		<input type="hidden" name="image_uri" value={form.card_image_normal} />
		<input type="hidden" name="colors" value={form.card_colors} />
		<input type="hidden" name="mana_cost" value={form.card_mana_cost} />
		<input type="hidden" name="setname" value={form.card_set_name} />

		<button
			type="submit"
			class="flex w-full justify-center rounded-md bg-indigo-600 px-3 py-1.5 text-sm/6 font-semibold text-white shadow-xs hover:bg-indigo-500"
		>
			Sauvegarder dans la DB
		</button>
	</form>
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
