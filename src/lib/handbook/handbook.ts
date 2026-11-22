import { pushState } from "$app/navigation";
import type { HandbookEntryId } from "./entries";

/**
 * Opens the Handbook Modal with shallow routing.
 *
 * The actual Modal element is created/managed in the app +Layout.svelte file.
 */
export function openHandbookModal(handbook: HandbookEntryId) {
	pushState("", { handbook });
}
