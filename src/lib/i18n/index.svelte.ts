import { browser } from "$app/environment";
import i18next, { type TOptions } from "i18next";
import en from "./locales/en.json";

/**
 * Language used as the base of all translations and as a fallback
 * for any missing keys in other languages.
 */
export const defaultLanguage = "en";

/**
 * localStorage key where the users language preference is cached,
 * so we can show the right language before we have loaded the
 * users settings from the server (and on pages like login).
 */
const storageKey = "language";

type Translations = { [key: string]: string | Translations };

/**
 * Every json file in `./locales` (except english which is
 * always bundled) is a language that can be loaded on demand.
 * Adding a new language only requires adding its file there.
 */
const loaders = import.meta.glob<Translations>(
	["./locales/*.json", "!./locales/en.json"],
	{ import: "default" },
);

const languageLoaders: Record<string, () => Promise<Translations>> = {};
for (const [path, loader] of Object.entries(loaders)) {
	const code = path
		.replace(/^.*\/([^/]+)\.json$/, "$1")
		// Weblate may name files with an underscore (pt_BR).
		.replaceAll("_", "-");
	languageLoaders[code] = loader;
}

/**
 * Get the name of a language, in that language.
 */
function getNativeName(code: string) {
	try {
		const name = new Intl.DisplayNames([code], { type: "language" }).of(code);
		if (name) {
			return name.charAt(0).toLocaleUpperCase(code) + name.slice(1);
		}
	} catch (err) {
		console.warn("i18n: Failed to get native name for", code, err);
	}
	return code;
}

/**
 * All languages that the ui can be shown in.
 */
export const availableLanguages = [
	defaultLanguage,
	...Object.keys(languageLoaders),
]
	.map((code) => ({ code, name: getNativeName(code) }))
	.sort((a, b) => a.name.localeCompare(b.name));

let currentLanguage = $state(defaultLanguage);

i18next.init({
	lng: defaultLanguage,
	fallbackLng: defaultLanguage,
	resources: { [defaultLanguage]: { translation: en } },
	// Resources are bundled (or loaded by us before changing
	// language), so init can be synchronous.
	initAsync: false,
	// Svelte already escapes everything we render.
	interpolation: { escapeValue: false },
	returnNull: false,
});

// Allow translations to lowercase a value (eg `{{status, lowercase}}`),
// for languages that don't capitalize words in the middle of a sentence.
i18next.services.formatter?.add("lowercase", (value, lng) =>
	String(value).toLocaleLowerCase(lng),
);

/**
 * Translate a key into the current language.
 *
 * Reading the current language here means anything using `t`
 * reactively (eg in markup or `$derived`) will update when
 * the language changes.
 */
export function t(key: string, options?: TOptions): string {
	const lng = currentLanguage;
	return i18next.t(key, { ...options, lng }) as string;
}

/**
 * Get the current ui language (reactive).
 */
export function getLanguage() {
	return currentLanguage;
}

/**
 * Get the locale to use when formatting dates and numbers.
 * Prefers the browsers locale when it is a variant of the
 * current language (eg `en-GB` when the ui is in `en`), so
 * we keep the users regional formatting preferences.
 */
export function getFormattingLocale() {
	const lang = currentLanguage;
	if (browser) {
		const base = lang.split("-")[0].toLowerCase();
		const match = navigator.languages?.find(
			(l) => l.split("-")[0].toLowerCase() === base,
		);
		if (match) return match;
	}
	return lang;
}

/**
 * Find the best supported language from a list of
 * preferred languages (eg: `navigator.languages`).
 */
function matchLanguage(preferred: readonly string[]) {
	const codes = availableLanguages.map((l) => l.code);
	for (const p of preferred) {
		const lower = p.toLowerCase();
		const exact = codes.find((c) => c.toLowerCase() === lower);
		if (exact) return exact;
		const base = lower.split("-")[0];
		const partial = codes.find((c) => c.toLowerCase().split("-")[0] === base);
		if (partial) return partial;
	}
	return undefined;
}

/**
 * Get the users language preference cached in localStorage.
 * An empty string means detect from the browser.
 */
export function getStoredLanguagePreference() {
	try {
		return localStorage.getItem(storageKey) ?? "";
	} catch {
		return "";
	}
}

function storeLanguagePreference(pref: string) {
	try {
		if (pref) {
			localStorage.setItem(storageKey, pref);
		} else {
			localStorage.removeItem(storageKey);
		}
	} catch (err) {
		console.warn("i18n: Failed to save language preference", err);
	}
}

/**
 * Apply a language preference.
 *
 * @param pref Language code the user chose. Empty or undefined
 * means use the browsers language (falling back to english).
 */
export async function setLanguage(pref?: string) {
	pref = pref ?? "";
	storeLanguagePreference(pref);
	const lang =
		matchLanguage(pref ? [pref] : []) ??
		(browser ? matchLanguage(navigator.languages ?? []) : undefined) ??
		defaultLanguage;
	if (!i18next.hasResourceBundle(lang, "translation")) {
		const load = languageLoaders[lang];
		if (load) {
			try {
				i18next.addResourceBundle(lang, "translation", await load());
			} catch (err) {
				console.error("i18n: Failed to load language", lang, err);
				return;
			}
		}
	}
	await i18next.changeLanguage(lang);
	currentLanguage = lang;
	if (browser) document.documentElement.lang = lang;
	console.debug("i18n: Language set to", lang);
}
