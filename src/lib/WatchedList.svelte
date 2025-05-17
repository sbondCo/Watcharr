<script lang="ts">
	import { goto } from "$app/navigation";
	import Icon from "@/lib/Icon.svelte";
	import Poster from "@/lib/poster/Poster.svelte";
	import PosterList from "@/lib/poster/PosterList.svelte";
	import { store, clearActiveFilters } from "@/store.svelte";
	import type { Watched } from "@/types";
	import GamePoster from "./poster/GamePoster.svelte";
	import { notify } from "./util/notify";
	import Spinner from "./Spinner.svelte";

	interface Props {
		list: Watched[];
		isLoading: boolean;
		isPublicList?: boolean;
	}

	let { list, isPublicList = false, isLoading = false }: Props = $props();

	let sort = $derived(store.activeSort);
	let filters = $derived(store.activeFilters);
	let settings = $derived(store.userSettings);
	let watched: Watched[] = $state([]);

	/**
	 * Checks if content has been watched previously
	 * by analyzing the watched entrys activity.
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

	// Monsterous code for filters. Soz.
	function filt() {
		console.debug("WatchedList: filt()");
		try {
			// Set watched to list and sort it.
			watched = list
				.sort((a, b) => {
					if (sort[0] === "LASTFIN") {
						const aLastFinishActivity = a.activity
							?.sort(
								(aa, bb) =>
									Date.parse(bb.customDate ?? bb.updatedAt) -
									Date.parse(aa.customDate ?? aa.updatedAt),
							)
							?.find(
								(aa) =>
									// PORT NOTE: ALSO `IMPORTED_WATCHED` & `IMPORTED_ADDED_WATCHED`
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
					watched = watched.filter(
						(w) =>
							(filters.status.includes(w.status?.toLowerCase()) ||
								contentWatchedPreviously(w)) &&
							filters.type.includes(
								w.content ? w.content.type : w.game ? "game" : "",
							),
					);
				} else {
					watched = watched.filter(
						(w) =>
							filters.status.includes(w.status?.toLowerCase()) &&
							filters.type.includes(
								w.content ? w.content.type : w.game ? "game" : "",
							),
					);
				}
			} else if (filters.type.length > 0) {
				// Only filter type
				watched = watched.filter((w) =>
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
					watched = watched.filter(
						(w) =>
							filters.status.includes(w.status?.toLowerCase()) ||
							contentWatchedPreviously(w),
					);
				} else {
					watched = watched.filter((w) =>
						filters.status.includes(w.status?.toLowerCase()),
					);
				}
			}
		} catch (err) {
			console.error("filt: Failed to filter/sort current list!", err);
			notify({
				text: "Failed to filter/sort list!",
				type: "error",
				time: 6000,
			});
		}
	}

	/**
	 * Callback for when a watched list item is updated through poster,
	 * this allows us to run the filt() func again so the sorting is
	 * updated.
	 */
	function itemUpdated() {
		console.debug("itemUpdated");
		// filt();
	}
</script>

<PosterList>
	{#if list?.length > 0}
		{#each list as w, i (w.id)}
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
					onUpdated={itemUpdated}
				/>
			{:else if w.content}
				<Poster
					bind:watched={list[i]}
					media={{
						id: w.content.tmdbId,
						poster_path: w.content.poster_path,
						title: w.content.title,
						overview: w.content.overview,
						media_type: w.content.type,
						release_date: w.content.release_date,
						first_air_date: w.content.first_air_date,
					}}
					disableInteraction={isPublicList}
					fluidSize={true}
					pinned={w.pinned}
					onUpdated={itemUpdated}
				/>
			{/if}
		{/each}
	{:else if !isLoading}
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
</PosterList>

{#if isLoading}
	<div style="margin-bottom: 60px;">
		<Spinner />
	</div>
{/if}

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
</style>
