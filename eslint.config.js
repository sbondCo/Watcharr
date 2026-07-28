import prettier from "eslint-config-prettier";
import path from "node:path";
import js from "@eslint/js";
import svelte from "eslint-plugin-svelte";
import { defineConfig, includeIgnoreFile } from "eslint/config";
import globals from "globals";
import ts from "typescript-eslint";
import svelteConfig from "./svelte.config.js";

const gitignorePath = path.resolve(import.meta.dirname, ".gitignore");

export default defineConfig(
	includeIgnoreFile(gitignorePath),
	js.configs.recommended,
	ts.configs.recommended,
	svelte.configs.recommended,
	prettier,
	svelte.configs.prettier,
	{
		languageOptions: {
			globals: { ...globals.browser /* ...globals.node */ },
			parserOptions: {
				projectService: true,
				parser: ts.parser,
				ecmaVersion: "latest",
			},
		},
		rules: {
			// typescript-eslint strongly recommend that you do not use the no-undef lint rule on TypeScript projects.
			// see: https://typescript-eslint.io/troubleshooting/faqs/eslint/#i-get-errors-from-the-no-undef-rule-about-global-variables-not-being-defined-even-though-there-are-no-typescript-errors
			"no-undef": "off",
		},
	},
	{
		files: ["**/*.svelte", "**/*.svelte.ts", "**/*.svelte.js"],
		languageOptions: {
			parserOptions: {
				extraFileExtensions: [".svelte"],
				svelteConfig,
			},
		},
	},
	{
		// Override or add rule settings here.
		rules: {
			"@typescript-eslint/no-empty-object-type": [
				"error",
				{
					// Allowing this because it's nice to create an empty
					// interface that extends a base interface for specific
					// use, even if it is currently empty, incase i add stuff
					// to it in the future. It's also less confusing to read
					// the actual type name I want instead of the base type
					// everywhere in certain scenarios.
					allowInterfaces: "with-single-extends",
				},
			],
		},
	},
);
