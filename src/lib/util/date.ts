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
