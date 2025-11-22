import type { PageServerLoad } from './$types';
import type { Actions } from './$types';
import { error, fail } from '@sveltejs/kit';
import { db } from '$lib/server/db';
import { cards, collection_cards, collections, user } from '$lib/server/db/schema';
import { eq, lt, gte, ne, sql } from 'drizzle-orm';

export const load: PageServerLoad = async ({ locals }) => {
	const authuser = locals.user;
	{
		if (!authuser) throw error(401, 'You must be authenticated');
	}

	const userCards = await db.select().from(cards).where(eq(cards.userId, authuser.id));
	const userCollections = await db
		.select()
		.from(collections)
		.where(eq(collections.userId, authuser.id));

	return {
		cards: userCards,
		collections: userCollections
	};
};

export const actions = {
	addToCollection: async ({ locals, request }) => {
		const user = locals.user;

		if (!user) {
			return fail(401, {
				context: 'save',
				error: 'Non authentifié'
			});
		}

		const data = await request.formData();
		const collection_id = data.get('collection_id');
		const card_id = data.get('card_id');
		const quantity = data.get('quantity');

		await db
			.insert(collection_cards)
			.values({
				collection_id,
				card_id,
				quantity: parseInt(quantity) // Convertir en nombre
			})
			.onConflictDoUpdate({
				target: [collection_cards.collection_id, collection_cards.card_id],
				set: {
					quantity: sql`${collection_cards.quantity} + ${parseInt(quantity)}`
				}
			});

		return {
			context: 'addToCollection',
			collection_id: collection_id,
			card_id: card_id,
			quantity: quantity
		};
	}
} satisfies Actions;
