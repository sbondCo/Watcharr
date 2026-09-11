import type { HandbookEntryId } from "./lib/handbook/entries";

// See https://kit.svelte.dev/docs/types#app
// for information about these interfaces
declare global {
	namespace App {
		interface PageState {
			handbook?: HandbookEntryId;
		}
	}
}

export {};
