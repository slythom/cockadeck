import type { Actions } from './$types';
import { fail } from '@sveltejs/kit';
import { db } from '$lib/server/db';
import { card, deck, deckCard } from '$lib/server/db/schema';

export const actions = {
	search: async ({ request, fetch }) => {
		const data = await request.formData();
		// const quantity = data.get('quantity');
		const set_code = data.get('set_code');
		const collector_number = data.get('collector_number');

		const api_url = `https://api.scryfall.com/cards/${set_code}/${collector_number}`;
		console.log(api_url);
		const res = await fetch(api_url);

		if (!res.ok) {
			return fail(404, {
				context: 'search',
				error: 'Carte non trouvée sur Scryfall'
			});
		}

		const card = await res.json();
		console.log(card);

		const transformImage = card.image_uris?.normal ?? card.card_faces?.[0]?.image_uris?.normal;

		return {
			context: 'search',
			// card_quantity: Number(quantity ?? 0),
			card_setcode: card.set,
			card_collector_number: card.collector_number,
			card_name: card.name,
			card_set_name: card.set_name,
			card_image_normal: transformImage,
			card_colors: card.colors,
			card_mana_cost: card.mana_cost
		};
	},

	save: async ({ request, locals }) => {
		const user = locals.user;

		if (!user) {
			return fail(401, {
				context: 'save',
				error: 'Non authentifié'
			});
		}

		const data = await request.formData();
		const name = String(data.get('name') ?? '');
		const setcode = String(data.get('setcode') ?? '');
		const collector_number = String(data.get('collector_number') ?? '');
		const image_uri = String(data.get('image_uri') ?? '');
		const quantity = Number(data.get('quantity') ?? 1);
		const colors = String(data.get('colors') ?? '');
		const mana_cost = String(data.get('mana_cost') ?? '');
		const setname = String(data.get('setname') ?? '');

		await db.insert(card).values({
			id: crypto.randomUUID(),
			userId: user.id,
			name,
			setcode,
			collector_number,
			image_uri,
			quantity,
			colors,
			mana_cost,
			setname
		});

		// Retourne les données de la carte pour pouvoir continuer à l'afficher
		return {
			context: 'save',
			message: 'Carte sauvegardée avec succès !',
			card_setcode: setcode,
			card_collector_number: collector_number,
			card_name: name,
			card_set_name: data.get('setname'),
			card_image_normal: image_uri,
			card_colors: data.get('colors'),
			card_mana_cost: data.get('mana_cost'),
			card_quantity: quantity
		};
	},

	createDeck: async ({ request, locals }) => {
		const user = locals.user;

		if (!user) {
			return fail(401, {
				context: 'save',
				error: 'Non authentifié'
			});
		}

		const data = await request.formData();
		const name = String(data.get('name') ?? '');

		await db.insert(deck).values({
			id: crypto.randomUUID(),
			userId: user.id,
			name
		});

		// Retourne les données de la carte pour pouvoir continuer à l'afficher
		return {
			context: 'createDeck',
			message: 'Deck créée avec succès !',
			name: name
		};
	},
} satisfies Actions;
