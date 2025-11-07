<script lang="ts">
	import { goto } from "$app/navigation";
	import Icon from "@/lib/Icon.svelte";
	import Poster from "@/lib/poster/Poster.svelte";
	import PosterList from "@/lib/poster/PosterList.svelte";
	import { store, clearActiveFilters } from "@/store.svelte";
	import type { Watched, UserSettings, Filters } from "@/types";
	import GamePoster from "./poster/GamePoster.svelte";
	import { getLatestWatchedInTv } from "./util/helpers";
	import { notify } from "./util/notify";
	import { onMount, onDestroy } from "svelte";

	interface Props {
		list: Watched[];
		isPublicList?: boolean;
	}

	let { list, isPublicList = false }: Props = $props();

	let sort = $derived(store.activeSort);
	let filters = $derived(store.activeFilters);
	let settings = $derived(store.userSettings);

	const increment = 30;
	let visibleCount = $state(increment);
	let loading = false;
	let observer: IntersectionObserver;

	/**
	 * Checks if content has been watched previously
	 * by analyzing the watched entrys activity (with
	 * the latest AI improvements added in of course.)
	 */
	function contentWatchedPreviously(w: Watched) {
		let wp = false;
		const relatedActivity = w.activity.filter(
			(a) =>
				a.type === "ADDED_WATCHED" ||
				a.type === "IMPORTED_ADDED_WATCHED" ||
				a.type === "IMPORTED_WATCHED" ||
				a.type === "STATUS_CHANGED",
		);
		for (let i = 0; i < relatedActivity.length; i++) {
			const ra = relatedActivity[i];
			if (ra.type === "IMPORTED_ADDED_WATCHED") {
				wp = true;
				break;
			} else if (
				ra.type === "ADDED_WATCHED" ||
				ra.type === "IMPORTED_WATCHED"
			) {
				const data = JSON.parse(ra.data);
				if (data?.status == "FINISHED") {
					wp = true;
					break;
				}
			} else if (ra.type === "STATUS_CHANGED") {
				if (ra.data === "FINISHED") {
					wp = true;
					break;
				}
			}
		}
		return wp;
	}

	function filt(
		list: Watched[],
		sort: string[],
		filters: Filters,
		settings: UserSettings | undefined,
	) {
		console.debug("WatchedList: filt()");
		try {
			let result = [...list]
				.sort((a, b) => {
					if (sort[0] === "DATEADDED" && sort[1] === "UP") {
						return Date.parse(a.createdAt) - Date.parse(b.createdAt);
					} else if (sort[0] === "ALPHA") {
						const atitle = a.content
							? a.content.title
							: a.game
								? a.game.name
								: "";
						const btitle = b.content
							? b.content.title
							: b.game
								? b.game.name
								: "";
						if (sort[1] === "UP") {
							return atitle.localeCompare(btitle);
						} else if (sort[1] === "DOWN") {
							return btitle.localeCompare(atitle);
						}
					} else if (sort[0] === "LASTCHANGED") {
						if (sort[1] === "UP")
							return Date.parse(a.updatedAt) - Date.parse(b.updatedAt);
						else if (sort[1] === "DOWN")
							return Date.parse(b.updatedAt) - Date.parse(a.updatedAt);
					} else if (sort[0] === "LASTFIN") {
						const aLastFinishActivity = a.activity
							?.sort(
								(aa, bb) =>
									Date.parse(bb.customDate ?? bb.updatedAt) -
									Date.parse(aa.customDate ?? aa.updatedAt),
							)
							?.find(
								(aa) =>
									(aa.type === "STATUS_CHANGED" && aa.data === "FINISHED") ||
									(aa.type === "ADDED_WATCHED" &&
										aa.data?.includes("FINISHED")),
							);
						const bLastFinishActivity = b.activity
							?.sort(
								(aa, bb) =>
									Date.parse(bb.customDate ?? bb.updatedAt) -
									Date.parse(aa.customDate ?? aa.updatedAt),
							)
							?.find(
								(aa) =>
									(aa.type === "STATUS_CHANGED" && aa.data === "FINISHED") ||
									(aa.type === "ADDED_WATCHED" &&
										aa.data?.includes("FINISHED")),
							);
						if (!aLastFinishActivity) return 1;
						if (!bLastFinishActivity) return -1;
						const alfaDate =
							aLastFinishActivity.customDate ?? aLastFinishActivity.updatedAt;
						const blfaDate =
							bLastFinishActivity.customDate ?? bLastFinishActivity.updatedAt;
						if (sort[1] === "UP")
							return Date.parse(alfaDate) - Date.parse(blfaDate);
						else if (sort[1] === "DOWN")
							return Date.parse(blfaDate) - Date.parse(alfaDate);
					} else if (sort[0] === "RATING") {
						if (sort[1] === "UP") return (a.rating ?? 0) - (b.rating ?? 0);
						else if (sort[1] === "DOWN")
							return (b.rating ?? 0) - (a.rating ?? 0);
					}
					// default DATEADDED DOWN
					return Date.parse(b.createdAt) - Date.parse(a.createdAt);
				})
				.sort((a, b) => {
					if (a.pinned && !b.pinned) return -1;
					if (!a.pinned && b.pinned) return 1;
					return 0;
				});
			// Now apply filters to watch list.
			if (filters.status.length > 0 && filters.type.length > 0) {
				// If status and type filters applied, combine both.
				if (
					settings?.includePreviouslyWatched &&
					filters.status.includes("finished")
				) {
					result = result.filter(
						(w) =>
							(filters.status.includes(w.status?.toLowerCase()) ||
								contentWatchedPreviously(w)) &&
							filters.type.includes(
								w.content ? w.content.type : w.game ? "game" : "",
							),
					);
				} else {
					result = result.filter(
						(w) =>
							filters.status.includes(w.status?.toLowerCase()) &&
							filters.type.includes(
								w.content ? w.content.type : w.game ? "game" : "",
							),
					);
				}
			} else if (filters.type.length > 0) {
				// Only filter type
				result = result.filter((w) =>
					filters.type.includes(
						w.content ? w.content.type : w.game ? "game" : "",
					),
				);
			} else if (filters.status.length > 0) {
				// Only filter status
				if (
					settings?.includePreviouslyWatched &&
					filters.status.includes("finished")
				) {
					result = result.filter(
						(w) =>
							filters.status.includes(w.status?.toLowerCase()) ||
							contentWatchedPreviously(w),
					);
				} else {
					result = result.filter((w) =>
						filters.status.includes(w.status?.toLowerCase()),
					);
				}
			}

			return result;
		} catch (err) {
			console.error("filt: Failed to filter/sort current list!", err);
			notify({
				text: "Failed to filter/sort list!",
				type: "error",
				time: 6000,
			});
			return [];
		}
	}

	// NOTE: I believe this should only recaulculate watched when necessary and similarly only visibleWatched when necessary
	let result = $derived.by(() => filt(list, sort, filters, settings));
	let visibleWatched = $derived.by(() => result.slice(0, visibleCount));

	function setupObserver() {
		const sentinel = document.querySelector("#sentinel");
		if (!sentinel) return;

		let timeout: number;
		observer = new IntersectionObserver(
			(entries) => {
				for (const entry of entries) {
					if (entry.isIntersecting && !loading) {
						loading = true;
						clearTimeout(timeout);
						timeout = window.setTimeout(() => {
							visibleCount += increment;
							loading = false;
						}, 200);
					}
				}
			},
			{ rootMargin: "100px" },
		);
		observer.observe(sentinel);
	}

	onMount(setupObserver);
	onDestroy(() => observer?.disconnect());
</script>

<PosterList>
	{#if visibleWatched?.length > 0}
		{#each visibleWatched as w (w.id)}
			{#if w.game}
				<GamePoster
					id={w.id}
					rating={w.rating}
					status={w.status}
					media={{
						id: w.game.igdbId,
						coverId: w.game.coverId,
						name: w.game.name,
						summary: w.game.summary,
						firstReleaseDate: w.game.releaseDate,
						poster: w.game.poster,
					}}
					disableInteraction={isPublicList}
					extraDetails={{
						dateAdded: w.createdAt,
						dateModified: w.updatedAt,
					}}
					fluidSize={true}
					pinned={w.pinned}
				/>
			{:else if w.content}
				<Poster
					id={w.id}
					media={{
						id: w.content.tmdbId,
						poster_path: w.content.poster_path,
						title: w.content.title,
						overview: w.content.overview,
						media_type: w.content.type,
						release_date: w.content.release_date,
						first_air_date: w.content.first_air_date,
					}}
					rating={w.rating}
					status={w.status}
					disableInteraction={isPublicList}
					extraDetails={{
						dateAdded: w.createdAt,
						dateModified: w.updatedAt,
						lastWatched: getLatestWatchedInTv(
							w.watchedSeasons,
							w.watchedEpisodes,
						),
					}}
					fluidSize={true}
					pinned={w.pinned}
				/>
			{/if}
		{/each}
	{:else}
		<div class="empty-list">
			{#if list?.length > 0}
				<!-- `watched` (filtered list) is empty, but `list` (unfiltered) isn't,
          so we should let the user know why there is nothing to show. -->
				<Icon i="filter-circle" wh={80} />
				<h2 class="norm">Filters are hiding all results!</h2>
				<h4 class="norm">Try changing or removing your active filters.</h4>
				<button onclick={() => clearActiveFilters()}>Clear Filters</button>
			{:else}
				<Icon i="reel" wh={80} />
				{#if isPublicList}
					<h2 class="norm">This watched list is empty!</h2>
					<h4 class="norm">
						Come back later to see if they have added anything.
					</h4>
				{:else}
					<h2 class="norm">Your watched list is empty!</h2>
					<h4 class="norm">
						Try searching for something you would like to add.
					</h4>
					<button onclick={() => goto("/import")}>Import</button>
				{/if}
			{/if}
		</div>
	{/if}

	<!-- NOTE: For some reason button is the only el that doesn't cause flickering issues -->
	<button id="sentinel" aria-label="sentinel" class="sentinel"></button>
</PosterList>

<style lang="scss">
	.empty-list {
		display: flex;
		flex-flow: column;
		gap: 5px;
		align-items: center;

		h2 {
			margin-top: 10px;
		}

		h4 {
			font-weight: normal;
		}

		button {
			width: max-content;
			padding-left: 20px;
			padding-right: 20px;
			margin-top: 15px;
		}
	}

	.sentinel {
		visibility: hidden;
	}
</style>
