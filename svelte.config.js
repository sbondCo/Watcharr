import adapter from "@sveltejs/adapter-node";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";
import { sveltePreprocess } from "svelte-preprocess";

if (
	process.env.WATCHARR_BASE &&
	(!process.env.WATCHARR_BASE.startsWith("/") ||
		process.env.WATCHARR_BASE.endsWith("/"))
) {
	throw new Error("WATCHARR_BASE must start with, but not end with '/'");
}

/** @type {import('@sveltejs/kit').Config} */
const config = {
	// Consult https://kit.svelte.dev/docs/integrations#preprocessors
	// for more information about preprocessors
	preprocess: [
		sveltePreprocess({
			scss: {
				prependData: `@import "./src/norm.scss";`,
			},
		}),
		vitePreprocess(),
	],

	kit: {
		adapter: adapter(),

		alias: {
			"@": "src",
		},

		paths: {
			base: process.env.WATCHARR_BASE,
			relative: true,
		},
	},
};

export default config;
