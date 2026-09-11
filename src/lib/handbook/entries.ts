import type { Component } from "svelte";

type HandbookGroup = {
	title: string;
	entries: HandbookEntry[];
};

export type HandbookEntry = {
	id: string;
	title: string;
	comp: () => Promise<{ default: Component }>;
};

export const entries = [
	{
		title: "General",
		entries: [
			{
				id: "adding-to-your-watchlist",
				title: "Adding to your watchlist",
				comp: () =>
					import("./entries/general/AddingToYourWatchlist/AddingToYourWatchlist.svelte"),
			},
		],
	},
	{
		title: "Admin",
		entries: [
			{
				id: "jellyfin-webhook-setup",
				title: "Jellyfin Webhook Setup",
				comp: () => import("./entries/admin/JellyfinWebhookSetup.svelte"),
			},
		],
	},
] as const satisfies HandbookGroup[];

export type HandbookEntryId = (typeof entries)[number]["entries"][number]["id"];
