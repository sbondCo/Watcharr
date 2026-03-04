<script lang="ts">
	import { page } from "$app/state";
	import { afterNavigate, goto } from "$app/navigation";
	import axios, { type GenericAbortSignal } from "axios";
	import Poster from "@/lib/poster/Poster.svelte";
	import PosterList from "@/lib/poster/PosterList.svelte";
	import { store } from "@/store.svelte.js";
	import Spinner from "@/lib/Spinner.svelte";
	import PersonPoster from "@/lib/poster/PersonPoster.svelte";
	import {
		MediaTypeE,
		SearchType,
		type Media,
		type PublicUser,
		type SearchRequest,
		type SearchResponseMeta,
	} from "@/types";
	import UsersList from "@/lib/UsersList.svelte";
	import { onDestroy, onMount } from "svelte";
	import Error from "@/lib/Error.svelte";
	import infScroll from "@/lib/util/infScroll.js";
	import paginatedLoader, {
		PaginatedLoaderRunFnAction,
	} from "@/lib/util/paginatedLoader.svelte.js";
	import PageTitle from "@/lib/generic/PageTitle.svelte";
	import MediaTypeFilter from "@/lib/search/MediaTypeFilter.svelte";

	let { data } = $props();

	const scroll = infScroll({ callback: onScrollToBottom });
	const dataLoader = paginatedLoader<Media, SearchResponseMeta>(load);

	let searchType: SearchType | undefined = $derived.by(() => {
		const t = page.url.searchParams.get("type");
		if (t) {
			return t as SearchType;
		}
		return SearchType.multi;
	});

	let preferMyList: boolean = $derived.by(() => {
		const t = page.url.searchParams.get("preferMyList");
		return Boolean(t);
	});
	let showingResultsFromMyList: boolean = $derived(
		Boolean(
			dataLoader.state.data?.length > 0 && dataLoader.state.meta?.fromMyList,
		),
	);

	let nextLoadParams: SearchRequest = $derived({
		page: dataLoader.state.page + 1,
		query: store.searchQuery,
		type: searchType,
		preferMyList: preferMyList,
	});

	async function load(signal: GenericAbortSignal) {
		console.debug("load: loadParams:", nextLoadParams);
		if (nextLoadParams.page === dataLoader.state.page) {
			console.warn("load: Already on this page, not loading it again!");
			return;
		}
		if (!nextLoadParams.query) {
			console.warn("load: There is no search query!");
			return;
		}
		const r = await axios.get(`/search`, {
			params: nextLoadParams,
			signal,
		});
		scroll.dataLoaded();
		return r;
	}

	async function onScrollToBottom() {
		// If an error is being shown, no more infinite scroll.
		if (dataLoader.state.reqLoadError) {
			return;
		}
		dataLoader.runFn();
	}

	function setActiveSearchFilter(to: SearchType | undefined) {
		console.debug("setActiveSearchFilter: to:", to);
		const curLocation = new URL(page.url);
		if (!to || searchType === to) {
			curLocation.searchParams.delete("type");
		} else {
			curLocation.searchParams.set("type", to);
		}
		// Running the goto will cause afterNavigate hook to be called,
		// which will run a fresh search, so nothing else to do here.
		goto(`?${curLocation.searchParams.toString()}`);
	}

	async function searchUsers(query: string) {
		return (await axios.get(`/user/search`, { params: { q: query } }))
			.data as PublicUser[];
	}

	onMount(() => {
		if (!store.searchQuery && data?.query) {
			store.searchQuery = decodeURIComponent(data?.query);
		}
		dataLoader.runFn(PaginatedLoaderRunFnAction.Reset);
	});

	afterNavigate((e) => {
		if (!e.from?.route?.id?.toLowerCase()?.includes("/search")) {
			// AfterNavigate will also be called when this page is mounted,
			// but that won't work for us since the OnMount hook also runs
			// a clean search, which can cause errors when both ran at same
			// time. We can't remove the OnMount hook since it's the only
			// hook to be ran if watcharr is first loaded at a search url.
			// `e.type` is always `goto` (that's how we search) so we can't
			// use that. The only alternative to only run this hook after a
			// navigation on the search page (query change), seems to be
			// checking the `from` property an making sure it's from the
			// `/search` route already.
			return;
		}
		console.log(
			"afterNavigate: Query changed, performing search.",
			"searchParams:",
			page.url.searchParams,
		);
		// Sync state (so back button updates search correctly)
		store.searchQuery = data?.query ? decodeURIComponent(data?.query) : "";
		dataLoader.abortReq("navigated away");
		dataLoader.runFn(PaginatedLoaderRunFnAction.Reset);
	});

	onDestroy(() => {
		console.debug("SEARCH PAGE DESTROYED");
		store.searchQuery = "";
		scroll.destroy();
		dataLoader.abortReq("page destroyed");
	});

	function dontPreferMyListClicked() {
		const curLocation = new URL(page.url);
		curLocation.searchParams.delete("preferMyList");
		// Running the goto will cause afterNavigate hook to be called,
		// which will run a fresh search, so nothing else to do here.
		goto(`?${curLocation.searchParams.toString()}`);
	}
</script>

<svelte:head>
	<title
		>Search Results{store.searchQuery
			? ` for '${store.searchQuery}'`
			: ""}</title
	>
</svelte:head>

<!-- <span style="position: sticky;top: 70px;">{curPage} / {maxContentPage}</span> -->
<div class="content">
	<div class="inner">
		{#if data?.query}
			<!-- Uses data?.query instead of store.searchQuery,
			 	so that the debounce of search is respected. -->
			{#await searchUsers(data?.query) then results}
				{#if results?.length > 0}
					<UsersList users={results} />
				{/if}
			{:catch err}
				<Error pretty="Failed to load users!" error={err} />
			{/await}

			<PageTitle title="Results">
				<MediaTypeFilter
					active={searchType}
					disabled={dataLoader.state.reqLoading || showingResultsFromMyList}
					onChange={(nowActive) => {
						setActiveSearchFilter(nowActive as SearchType | undefined);
					}}
				/>
			</PageTitle>

			{#if showingResultsFromMyList}
				<button
					class="from-my-list-msg plain"
					onclick={dontPreferMyListClicked}
				>
					<b>Showing results from your list.</b>
					<span>Do a full search instead?</span>
				</button>
			{/if}

			<PosterList>
				{#if dataLoader.state.data?.length > 0}
					{#each dataLoader.state.data as w, i (`${i}-${w.type}`)}
						{#if w.type === MediaTypeE.tmdbPerson}
							<PersonPoster
								id={w.ids.tmdb}
								name={w.name}
								path={w.extPosterPath}
							/>
						{:else if w.type === MediaTypeE.tmdbMovie || w.type === MediaTypeE.tmdbShow || w.type === MediaTypeE.igdbGame}
							<Poster
								media={w}
								bind:watched={dataLoader.state.data[i].watched}
								fluidSize
							/>
						{/if}
					{/each}
				{:else if !dataLoader.state.reqLoading && !dataLoader.state.reqLoadError}
					<!-- If search is running or we have an error, no point in showing 'no results' message. -->
					<h2 class="norm" title="Hovering over me doesn't change the facts ;(">
						No Results!
					</h2>
				{/if}
			</PosterList>

			{#if dataLoader.state.reqLoading}
				<div style="margin-bottom: 60px;">
					<Spinner />
				</div>
			{/if}

			{#if dataLoader.state.reqLoadError}
				<div style="margin-bottom: 60px;">
					<Error
						pretty="Failed to load results!"
						error={dataLoader.state.reqLoadError}
						onRetry={() => {
							dataLoader.state.reqLoadError = undefined;
							dataLoader.runFn(
								PaginatedLoaderRunFnAction.ResetIfOnFirstOrNoPage,
							);
						}}
					/>
				</div>
			{/if}
		{:else}
			<h2>No Search Query!</h2>
		{/if}
	</div>
</div>

<style lang="scss">
	.content {
		display: flex;
		width: 100%;
		justify-content: center;

		.inner {
			width: 100%;
			max-width: 1200px;
		}
	}

	button.from-my-list-msg {
		display: flex;
		flex-flow: column;
		align-items: flex-start;
		margin: 10px auto 0 auto;
		padding: 12px 20px;
		border-radius: 10px;
		color: $text-color;
		background-color: $accent-color;
		font-size: 16px;
		user-select: none;
		transition:
			color 100ms ease,
			background-color 100ms ease;

		&:hover {
			color: $bg-color;
			background-color: $accent-color-hover;
		}
	}
</style>
