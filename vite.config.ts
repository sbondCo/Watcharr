import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vite";
import { readFileSync } from "fs";
import {SvelteKitPWA} from '@vite-pwa/sveltekit'
const pkg = JSON.parse(readFileSync("package.json", "utf8"));

export default defineConfig({
  plugins: [
    sveltekit(), 
    SvelteKitPWA({
      manifest:{
        start_url: '/',
        scope: '/',
        display: 'standalone',
        icons: [
          {
            src: '/logo-col.png',
            sizes: '192x192',
            type: 'image/png',
          },
          {
            src: '/logo-col.png',
            sizes: '512x512',
            type: 'image/png',
          },
          {
            src: '/logo-col.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'any maskable',
          },
        ],
      }
    })
  ],
  define: {
    __WATCHARR_VERSION__: JSON.stringify(pkg.version)
  }
});
