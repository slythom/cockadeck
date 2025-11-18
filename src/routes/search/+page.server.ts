import type { Actions, PageServerLoad } from './$types';
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
			return { success: false, status: res.status };
		}

		const card = await res.json();

		if (!card) {
			return { success: false, context: 'search', error: 'Scryfall 404' };
		}

		return {
			success: true,
			card,
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
			return {
				success: false,
				context: 'save',
				error: 'Not authenticated.'
			};
		}

		const data = await request.formData();

		const name = String(data.get('name') ?? '');
		const setcode = String(data.get('setcode') ?? '');
		const collector_number = String(data.get('collector_number') ?? '');
		const image_uri = String(data.get('image_uri') ?? '');
		const quantity = Number(data.get('quantity') ?? 1);

		await db.insert(cards).values({
			id: crypto.randomUUID(),
			userId: user.id, // 👈 lie la carte à l'utilisateur
			name,
			setcode,
			collector_number,
			image_uri,
			quantity
		});

		return {
			success: true,
			context: 'save'
		};
	}
} satisfies Actions;
