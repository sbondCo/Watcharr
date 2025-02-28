import { base } from "$app/paths";

/**
 * Takes in a `path` and prefixes it with `base`.
 * This enables support for using base paths to access
 * the frontend by fixing all absolute paths passed through.
 */
export function toBaseUrl(path: string) {
	return base + path;
}
