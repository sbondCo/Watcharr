import adapter from "@sveltejs/adapter-node";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";
import sveltePreprocess from "svelte-preprocess";

/** @type {import('@sveltejs/kit').Config} */
const config = {
  // Consult https://kit.svelte.dev/docs/integrations#preprocessors
  // for more information about preprocessors
  preprocess: [
    sveltePreprocess({
      scss: {
        prependData: `@import "./src/norm.scss";`
      }
    }),
    vitePreprocess()
  ],

  kit: {
    adapter: adapter(),

    alias: {
      "@": "src"
    }
  }
};

export default config;
