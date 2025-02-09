<script lang="ts">
	import Error from "@/lib/Error.svelte";
	import infScroll from "@/lib/util/infScroll";
	import WatchedList from "@/lib/WatchedList.svelte";
	import type { PaginationResponse, Watched } from "@/types";
	import axios from "axios";
	import { onDestroy, onMount } from "svelte";

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
		console.debug("loadWatchedList:", listPage, listPageMax);
		if (listLoading) {
			console.debug("loadWatchedList: already running");
			return;
		}
		if (listPage >= listPageMax) {
			console.debug("loadWatchedList: max page reached");
			return;
		}
		listLoading = true;
		reqController = new AbortController();
		try {
			const pl = await axios.get<PaginationResponse<Watched>>(`/watched`, {
				params: {
					p: listPage + 1,
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

	onMount(() => {
		loadWatchedList();
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

listPage: {listPage} listPageMax: {listPageMax} listLoading: {listLoading}

<WatchedList {list} />
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
