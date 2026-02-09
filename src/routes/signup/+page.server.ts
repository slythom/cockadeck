import bcrypt from 'bcryptjs';
import { db } from '$lib/server/db';
import { fail } from '@sveltejs/kit';
import * as table from '$lib/server/db/schema';
import type { PageServerLoad, Actions } from './$types';

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
        const email = data.get('email')?.toString();
        const password = data.get('password')?.toString();
        // basic checks
        if (!email) {
            return fail(400, {
                message: 'Email not provided or incorrect format.'
            });
        }
        if (!password) {
            return fail(400, {
                message: 'Incorrect or missing password.'
            });
        }

        // check if account (email) already exists
        const existingUser = await db.query.user.findFirst({
            where: (user, { eq }) => eq(user.email, email)
        });
        if (existingUser) {
            return fail(400, {
                message: 'This email address already exists.'
            });
        }

        // generate user id
        const userId = generateUserId()

        // hash password
        const hashedPassword = await bcrypt.hash(password, 10);

        // insert user into DB
        await db.insert(table.user).values({ id: userId, email: email, passwordHash: hashedPassword });

        // TODO: generate session token

        return { success: true, message: "User registered successfully." };
    },
} satisfies Actions;