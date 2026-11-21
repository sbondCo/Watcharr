# Contributing 🎉

First of all, thank you for looking into contributing! 🫡

Feel free to contribute in any way: 🐞 bug reports, 💡 ideas, 🥖 new features etc - everything is welcome!

Big pull requests (anything more than a few lines) should be discussed first in an issue (dont want to waste anyones time!).

## Project Board

We have a [board to organize new features and bug fixes](https://github.com/orgs/sbondCo/projects/9/views/2) to be worked on. If you see something you'd like to work on, just ask in the issue and it can be assigned to you. Even if it isn't set for the next milestone already, it can be.

## AI

If you use AI, please disclose the usage in your pull request.

Fully understanding any code that is submitted is a must because someone's gotta understand it!

## Translations

The UI is translated with [i18next](https://www.i18next.com/). Translations live in [`src/lib/i18n/locales`](src/lib/i18n/locales), one JSON file per language ([i18next JSON v4 format](https://www.i18next.com/misc/json-format)). `en.json` is the base language and the fallback for anything not translated yet.

Once a translation platform (eg Weblate) is set up for the project, that will be the easiest way to contribute translations. Until then, translations can be contributed with a pull request.

### Adding a language

1. Copy `en.json` to a new file named after the language code (eg `de.json`, `pt-BR.json`).
2. Translate the values (never the keys).
3. That's it, the language is detected automatically and will show up in the language setting (with its name in that language).

Users can pick their language in their settings. If they haven't picked one, their browser language is used (falling back to English).

### Tips for translators

- Keep `{{placeholders}}` as they are, they are replaced with values (names, numbers, etc).
- Keys ending in `_one`, `_other` (and `_zero`, `_two`, `_few`, `_many`) are [plural forms](https://www.i18next.com/translation-function/plurals). Add the forms your language needs (see [plural rules](https://www.unicode.org/cldr/charts/latest/supplemental/language_plural_rules.html)), eg Italian also needs `_many`. `date.ordinalDay_ordinal_*` keys work the same for ordinal numbers (1st, 2nd..).
- Some strings contain simple tags like `<b>`, `<i>`, `<code>` or `<link>`. Keep the tags, but move them around the right words for your language.
- Use `{{value, lowercase}}` if a value (eg a watched status like "On Hold") should be lowercase in the middle of a sentence in your language.

### Adding new strings (for developers)

Don't hardcode user facing text. Add a key to `en.json` and use the `t` function:

```svelte
<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
</script>

<h2>{t("settings.pageTitle")}</h2>
<span>{t("season.episodes", { count: 3 })}</span>
```

For text with formatting or links, use the `Trans` component (`src/lib/i18n/Trans.svelte`) instead of `{@html}`.
