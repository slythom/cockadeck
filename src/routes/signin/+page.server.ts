import bcrypt from 'bcryptjs';
import { db } from '$lib/server/db';
import { fail } from '@sveltejs/kit';
import * as table from '$lib/server/db/schema';
import { eq, lt, gte, ne } from 'drizzle-orm';
import type { PageServerLoad, Actions } from './$types';

// export const load: PageServerLoad = async ({ cookies }) => {
// 	const user = await db.getUserFromSession(cookies.get('sessionid'));
// 	return { user };
// };

export const actions = {
    signin: async ({ cookies, request }) => {
        const data = await request.formData();
        const email = data.get('email')?.toString();
        const password = data.get('password')?.toString();
        // basic checks
        if (!email) {
            return fail(400, {
                success: false,
                message: 'Email not provided or incorrect format.'
            });
        }
        if (!password) {
            return fail(400, {
                success: false,
                message: 'Incorrect or missing password.'
            });
        }

        // compare email and password with db
        const findUser = await db.query.user.findFirst({
            where: (user, { eq }) => eq(user.email, email)
        });
        if (!findUser) {
            return fail(400, {
                success: false,
                message: 'User not found.'
            });
        }
        const isPasswordOk = await bcrypt.compare(password, findUser.passwordHash)
        if (!isPasswordOk) {
            return fail(400, { success: false, message: 'Incorrect password.' })
        }

        // TODO: generate session token

        return { success: true, message: "User successfully signed in." };
    },
} satisfies Actions;