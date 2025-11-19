import type { Actions } from './$types';
import { fail } from '@sveltejs/kit';
import { db } from '$lib/server/db';
import { cards } from '$lib/server/db/schema';

export const actions = {
	search: async ({ request, fetch }) => {
		const data = await request.formData();
		const quantity = data.get('quantity');
		const setcode = data.get('setcode');
		const collector_number = data.get('collector_number');

		const api_url = `https://api.scryfall.com/cards/${setcode}/${collector_number}`;
		const res = await fetch(api_url);

		if (!res.ok) {
			return fail(404, {
				context: 'search',
				error: 'Carte non trouvée sur Scryfall'
			});
		}

		const card = await res.json();

		return {
			context: 'search',
			card_quantity: Number(quantity ?? 0),
			card_setcode: card.set,
			card_collector_number: card.collector_number,
			card_name: card.name,
			card_set_name: card.set_name,
			card_image_normal: card.image_uris.small,
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

		await db.insert(cards).values({
			id: crypto.randomUUID(),
			userId: user.id,
			name,
			setcode,
			collector_number,
			image_uri,
			quantity
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
	}
} satisfies Actions;
