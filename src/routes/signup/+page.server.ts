import bcrypt from 'bcryptjs';
import * as db from '$lib/server/db';
import { error } from 'console';
import type { PageServerLoad, Actions } from './$types';
import * as auth from '$lib/server/auth';
import * as table from '$lib/server/db/schema';

// export const load: PageServerLoad = async ({ cookies }) => {
// 	const user = await db.getUserFromSession(cookies.get('sessionid'));
// 	return { user };
// };

function generateUserId() {
  const dateString = Date.now().toString(36);
  const randomness = Math.random().toString(36).substr(2);
  return dateString + randomness;
};

export const actions = {
	signup: async ({ cookies, request }) => {
		const data = await request.formData();
		const email = data.get('email') as string;
		const password = data.get('password') as string;
        // basic checks
        if (!email) {
            return {success: false, message: "Email not provided or incorrect format."}
        }
        if (!password) {
            return {success: false, message: "Incorrect password format."}
        }
        if (!data) {
            return {success: false, message: "No data found."}
        }

        // TODO: vérifier si user email déjà existant

        // generate user id
        const userId = generateUserId()

        // hash password
        const hashedPassword = await bcrypt.hash(password, 10);

		// insert user into DB
        await db.db.insert(table.user).values({ id: userId, email: email, passwordHash: hashedPassword });

        // TODO: generate session token

		return { success: true, message: "User registered successfully." };
	},
} satisfies Actions;