import type { PageServerLoad } from './$types';
import type { Actions } from './$types';
import { error, fail } from '@sveltejs/kit';
import { db } from '$lib/server/db';
import { cards, deck_cards, decks, user } from '$lib/server/db/schema';
import { eq, lt, gte, ne, sql, and } from 'drizzle-orm';

function generateXML(data: Array<{ cardName: string | null; quantity: number | null }>) {
	const cards = data
		.map((item) => `  <card number="${item.quantity ?? 0}" name="${item.cardName ?? ''}"/>`)
		.join('\n');

	return `<?xml version="1.0" encoding="UTF-8"?>
	<cockatrice_deck version="1">
<bannerCard providerId=""></bannerCard>
<comments></comments>
<tags/>
<zone name="main">
${cards}
</zone>
</cockatrice_deck>`;
}
export const load: PageServerLoad = async ({ locals }) => {
	const authuser = locals.user;
	if (!authuser) throw error(401, 'You must be authenticated');

	const userCards = await db.select().from(cards).where(eq(cards.userId, authuser.id));

	const userdecks = await db
		.select()
		.from(decks)
		.where(eq(decks.userId, authuser.id));

	// NOUVEAU : Charger aussi les cartes de decks avec JOIN
	const deckCardsData = await db
		.select({
			card_id: deck_cards.card_id,
			deck_id: deck_cards.deck_id,
			quantity: deck_cards.quantity,
			card_name: cards.name,
			image_uri: cards.image_uri,
			setcode: cards.setcode
			// ... autres champs dont tu as besoin
		})
		.from(deck_cards)
		.innerJoin(cards, eq(deck_cards.card_id, cards.id))
		.innerJoin(decks, eq(deck_cards.deck_id, decks.id))
		.where(eq(decks.userId, authuser.id));

	return {
		cards: userCards,
		decks: userdecks,
		deckCards: deckCardsData // Données pour filtrer côté client
	};
};

export const actions = {
	export: async ({ locals, request }) => {
		const user = locals.user;

		if (!user) {
			return fail(401, {
				context: 'export',
				error: 'Non authentifié'
			});
		}

		const data = await request.formData();
		const deck_id = data.get('deck_id') as string;

		// Récupérer toutes les cartes d'une collection avec JOIN
		const deckData = await db
			.select({
				cardName: cards.name,
				quantity: deck_cards.quantity,
				deckName: decks.name
			})
			.from(deck_cards)
			.innerJoin(cards, eq(deck_cards.card_id, cards.id))
			.innerJoin(decks, eq(deck_cards.deck_id, decks.id))
			.where(and(eq(decks.id, deck_id), eq(decks.userId, user.id)));

		// Générer le XML
		const xmlContent = generateXML(deckData);

		return {
			context: 'export',
			data: deckData,
			xml: xmlContent
		};
	},

	addToDeck: async ({ locals, request }) => {
		const user = locals.user;

		if (!user) {
			return fail(401, {
				context: 'addToDeck',
				error: 'Non authentifié'
			});
		}

		const data = await request.formData();
		const deck_id = String(data.get('deck_id'));
		const card_id = String(data.get('card_id'));
		const quantity = Number(data.get('quantity'));



		await db.insert(deck_cards).values({
            deck_id,
            card_id,
            quantity
        })
		.onConflictDoUpdate({
            target: [deck_cards.deck_id, deck_cards.card_id],
            set: {
                quantity: sql`${deck_cards.quantity} + ${quantity}`
            }
        });

		return {
			context: 'addToDeck',
			message:"Carte  ajoutée à la collec!"
		};
	}

} satisfies Actions;
