import { getFormattingLocale, t } from "@/lib/i18n/index.svelte";

/**
 * Date helpers.
 */

/**
 * Check if `d` is a **valid** `Date`.
 *
 * **NOTE:** Returns `false` even if `d` is an instance of `Date`,
 * if it doesn't contain a valid date.
 */
export function dateValid(d: unknown): d is Date {
	return d instanceof Date && !isNaN(d.getTime());
}

/**
 * Format a date (and/or time) in the users current language.
 * With no `opts`, formats like `Date.toLocaleDateString()`.
 */
export function formatDate(d: Date, opts?: Intl.DateTimeFormatOptions) {
	return new Intl.DateTimeFormat(getFormattingLocale(), opts).format(d);
}

/**
 * Format the time of a date in the users current language.
 * Formats like `Date.toLocaleTimeString()`.
 */
export function formatTime(d: Date) {
	return formatDate(d, {
		hour: "numeric",
		minute: "2-digit",
		second: "2-digit",
	});
}

/**
 * Get the day of the month as an ordinal (eg `3rd` in english).
 */
export function ordinalDay(d: Date) {
	return t("date.ordinalDay", { count: d.getDate(), ordinal: true });
}

/**
 * Get a short, readable date.
 * eg (in english): `3rd Jan` or `3rd Jan 2024`.
 *
 * @param withYear Include the year. If `"short"`, a two digit year is used.
 */
export function formatShortDate(d: Date, withYear: boolean | "short" = true) {
	const month = formatDate(d, { month: "short" });
	if (!withYear) {
		return t("date.dayMonth", { day: ordinalDay(d), month });
	}
	const year = formatDate(d, {
		year: withYear === "short" ? "2-digit" : "numeric",
	});
	return t(
		withYear === "short" ? "date.dayMonthShortYear" : "date.dayMonthYear",
		{
			day: ordinalDay(d),
			month,
			year,
		},
	);
}
