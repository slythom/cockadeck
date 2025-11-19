import type { PageServerLoad } from './$types';
import type { Actions } from './$types';
import { error, fail } from '@sveltejs/kit';
import { db } from '$lib/server/db';
import { cards, user } from '$lib/server/db/schema';
import { eq, lt, gte, ne } from 'drizzle-orm';

export const load: PageServerLoad = async ({ locals }) => {
	const authuser = locals.user;
	// erreur if !authuser :
	{
		if (!authuser) throw error(401, 'You must be authenticated');
	}
	//
	// const userCards = await db.select({
	//      userId: authuser.id,
	//    })
	//    .from(cards);

	const userCards = await db.select().from(cards).where(eq(cards.userId, authuser.id));

	return {
		cards: userCards
	};
};
