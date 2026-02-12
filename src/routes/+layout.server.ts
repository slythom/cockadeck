export const load = async (event) => {
    if (event.locals.user) {
        const userEmail = event.locals.user.email
        console.log(userEmail)
        return {userEmail};
    }
    return {};
};