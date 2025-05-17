<script lang="ts">
	import Error from "@/lib/Error.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import infScroll from "@/lib/util/infScroll";
	import WatchedList from "@/lib/WatchedList.svelte";
	import { store } from "@/store.svelte";
	import type { PaginationResponse, Watched } from "@/types";
	import axios from "axios";
	import { onDestroy, onMount, untrack } from "svelte";

	const scroll = infScroll({ callback: onScrollToBottom });

	let reqController = new AbortController();
	let list: Watched[] = $state([]);
	// Current list page loaded
	let listPage = $state(0);
	// Max amount of pages for list
	let listPageMax = $state(1);
	let listLoading = $state(false);
	let listLoadError: any = $state();

	/**
	 * Fetches paginated watched list.
	 */
	async function loadWatchedList() {
		const logStyle = "font-weight: bold; font-size: 18px;";
		if (listLoading) {
			console.warn("%cloadWatchedList: already running", logStyle);
			return;
		}
		if (listPage >= listPageMax) {
			console.warn("%cloadWatchedList: max page reached", logStyle);
			return;
		}
		console.debug(
			`%cloadWatchedList: Page=${listPage} Max=${listPageMax}`,
			logStyle,
		);
		listLoading = true;
		reqController = new AbortController();
		try {
			const pl = await axios.get<PaginationResponse<Watched>>(`/watched`, {
				params: {
					p: listPage + 1,
					...store.sortAndFiltersForQueryParams,
				},
				signal: reqController.signal,
			});
			if (pl.data.results.length <= 0) {
				console.log("loadWatchedList: No results.");
				return;
			}
			listPage = pl.data.page;
			listPageMax = pl.data.totalPages;
			list.push(...pl.data.results);
			list = list;
			scroll.dataLoaded();
		} catch (err: any) {
			if (err?.code === "ERR_CANCELED") {
				console.warn("loadWatchedList: Cancelled, not showing error.");
			} else {
				console.error("loadWatchedList: failed!", err);
				listLoadError = err;
			}
		}
		listLoading = false;
	}

	async function onScrollToBottom() {
		// If an error is being shown, no more infinite scroll.
		if (listLoadError) {
			return;
		}
		loadWatchedList();
	}

	/**
	 * Resets our loaded watched list data, page data, etc.
	 */
	function resetWatchedList() {
		list = [];
		listPage = 0;
		listPageMax = 1;
		listLoading = false;
		listLoadError = undefined;
	}

	$effect(() => {
		// When our sort/filter query params change,
		// load our list again.
		// Since it exists at load, this performs our
		// initial load of data too.
		if (store.sortAndFiltersForQueryParams) {
			untrack(() => {
				// We don't want to trigger another re-run of this
				// effect when state inside these funcs changes.
				resetWatchedList();
				loadWatchedList();
			});
		}
	});

	onDestroy(() => {
		console.log("MAIN PAGE DESTROYED");
		scroll.destroy();
		reqController.abort("page destroyed");
	});
</script>

<svelte:head>
	<title>Watched List</title>
</svelte:head>

<span style="position: fixed; top: 80px; background-color: white; z-index: 60;"
	>listPage: {listPage} listPageMax: {listPageMax} listLoading: {listLoading}<br
	/>
	sort: {JSON.stringify(store.activeSort)} filter: {JSON.stringify(
		store.activeFilters,
	)}<br />
	{JSON.stringify(store.sortAndFiltersForQueryParams)}</span
>

<WatchedList {list} isLoading={listLoading} />
{#if listLoadError}
	<div style="margin-bottom: 60px;">
		<Error
			pretty="Failed to load results!"
			error={listLoadError}
			onRetry={() => {
				listLoadError = undefined;
				loadWatchedList();
			}}
		/>
	</div>
{/if}

<!-- TODO: A 'Thats It! All {watcheds_count} of your records.' message
 or similar message to indicate that we are now at the bottom of the
 users list (only show when listPage === listPageMax & we have any items.. etc
 probs best adding this to WatchedList component.. but message probs needs to be
 a bit different depending on viewing own list or someone elses). -->
