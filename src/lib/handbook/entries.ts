import type { Component } from "svelte";

type HandbookGroup = {
	/** Translation key of the title. */
	title: string;
	entries: HandbookEntry[];
};

export type HandbookEntry = {
	id: string;
	/** Translation key of the title. */
	title: string;
	comp: () => Promise<{ default: Component }>;
};

export const entries = [
	{
		title: "handbook.groups.general",
		entries: [
			{
				id: "adding-to-your-watchlist",
				title: "handbook.addingToWatchlist.title",
				comp: () =>
					import("./entries/general/AddingToYourWatchlist/AddingToYourWatchlist.svelte"),
			},
		],
	},
	{
		title: "handbook.groups.admin",
		entries: [
			{
				id: "jellyfin-webhook-setup",
				title: "handbook.jellyfinWebhook.title",
				comp: () => import("./entries/admin/JellyfinWebhookSetup.svelte"),
			},
			{
				id: "your-own-tmdb-api-key",
				title: "handbook.tmdbKey.title",
				comp: () => import("./entries/admin/YourOwnTMDBAPIKey.svelte"),
			},
		],
	},
] as const satisfies HandbookGroup[];

export type HandbookEntryId = (typeof entries)[number]["entries"][number]["id"];
