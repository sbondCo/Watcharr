// The ui language depends on the users preference (cached in
// localStorage) and browser language, which are only known on the
// client, so render everything client side.
export const ssr = false;
