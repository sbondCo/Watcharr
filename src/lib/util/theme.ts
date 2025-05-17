// App Theme logic.

import { browser } from "$app/environment";
import { store } from "@/store.svelte";
import type { Theme } from "@/types";

/**
 * Utility function to handle media queries for theme preferences.
 */

const prefersDarkThemeQuery: MediaQueryList | undefined = browser
	? window.matchMedia("(prefers-color-scheme: dark)")
	: undefined;

/**
 * prefersDarkThemeQuery onChange handler.
 */
function prefersDarkThemeChanged(e: MediaQueryListEvent) {
	console.info(
		"prefersDarkThemeChanged: User preferred theme has changed. prefersDark:",
		e.matches,
	);
	document.documentElement.classList.toggle("theme-dark", e.matches);
}

/**
 * Toggle site wide theme.
 * @param theme The theme to switch to.
 * @param updateStore Should the store be updated to new theme?
 * **If set to `false`, state should be manually updated.**
 */
export function toggleTheme(theme: Theme, updateStore = true) {
	if (updateStore) {
		store.appTheme = theme;
	}

	switch (theme) {
		case "dark":
			document.documentElement.classList.add("theme-dark");
			break;
		case "light":
			document.documentElement.classList.remove("theme-dark");
			break;
		case "system":
			document.documentElement.classList.toggle(
				"theme-dark",
				prefersDarkThemeQuery?.matches,
			);
			break;
	}

	if (prefersDarkThemeQuery) {
		// Always remove first before adding to avoid cases
		// where we add multiple events at same time.
		prefersDarkThemeQuery.removeEventListener(
			"change",
			prefersDarkThemeChanged,
		);
		// If using system theme, add change listener (since
		// system theme supports live updating when user
		// preference changes on os or wherever).
		if (theme === "system") {
			prefersDarkThemeQuery.addEventListener("change", prefersDarkThemeChanged);
		}
	}
}
