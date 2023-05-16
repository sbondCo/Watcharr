import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vite";
import { readFileSync } from "fs";

const pkg = JSON.parse(readFileSync("package.json", "utf8"));

export default defineConfig({
  plugins: [sveltekit()],
  define: {
    __WATCHARR_VERSION__: JSON.stringify(pkg.version)
  }
});
