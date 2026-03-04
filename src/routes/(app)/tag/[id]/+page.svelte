<script lang="ts">
	import { afterNavigate } from "$app/navigation";
	import { page } from "$app/state";
	import Error from "@/lib/Error.svelte";
	import Icon from "@/lib/Icon.svelte";
	import Poster from "@/lib/poster/Poster.svelte";
	import PosterList from "@/lib/poster/PosterList.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import CreateTagModal from "@/lib/tag/CreateTagModal.svelte";
	import Tag from "@/lib/tag/Tag.svelte";
	import infScroll from "@/lib/util/infScroll";
	import paginatedLoader, {
		PaginatedLoaderRunFnAction,
	} from "@/lib/util/paginatedLoader.svelte";
	import { clearActiveFilters, store } from "@/store.svelte.js";
	import type { Media } from "@/types";
	import axios, { type GenericAbortSignal } from "axios";
	import { onDestroy, untrack } from "svelte";

	let meta = $derived.by(() => {
		return {
			tagId: Number(page.params.id),
		};
	});
	let tag = $derived(store.tags.find((t) => t.id === meta.tagId));
	let tagEditModalShown = $state(false);

	const scroll = infScroll({ callback: onScrollToBottom });
	const dataLoader = paginatedLoader<Media, undefined>(load);

	let nextLoadParams: {
		page: number;
		[x: string]: any;
	} = $derived({
		page: dataLoader.state.page + 1,
		...store.sortAndFiltersForQueryParams,
	});

	async function load(signal: GenericAbortSignal) {
		console.debug("load: loadParams:", nextLoadParams);
		if (nextLoadParams.page === dataLoader.state.page) {
			console.warn("load: Already on this page, not loading it again!");
			return;
		}
		if (!meta.tagId) {
			console.warn("load: Missing tag id!");
			return;
		}
		const r = await axios.get(`/tag/${meta.tagId}/watched`, {
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
		console.debug("onScrollToBottom");
		dataLoader.runFn();
	}

	// NOTE: This effect also handles initial load of data.
	$effect(() => {
		// When our sort/filter query params change,
		// load our list again.
		// Since it exists at load, this performs our
		// initial load of data too.
		if (store.sortAndFiltersForQueryParams) {
			untrack(() => {
				// We don't want to trigger another re-run of this
				// effect when state inside these funcs changes.
				dataLoader.reset();
				dataLoader.runFn();
			});
		}
	});

	afterNavigate((e) => {
		if (!e.from?.route?.id?.toLowerCase()?.includes("/tag")) {
			// Ensure afterNavigate can only runFn when we are coming
			// from another tag page.
			// OnMount of a tag page we don't want to have this also run
			// because that breaks our loader and our effect will handle loading
			// in that case already.
			return;
		}
		console.log("afterNavigate.");
		dataLoader.abortReq("navigated away");
		dataLoader.runFn(PaginatedLoaderRunFnAction.Reset);
	});

	onDestroy(() => {
		console.debug("PAGE DESTROYED");
		scroll.destroy();
		dataLoader.abortReq("page destroyed");
	});
</script>

<svelte:head>
	<title>{tag ? `${tag.name} - Tag` : "Tag"}</title>
</svelte:head>

{#if tag}
	<div class="content">
		<div class="inner">
			<div class="basic-ctr">
				<Icon i="tag" wh={20} />
				<Tag
					{tag}
					onClick={() => {
						tagEditModalShown = !tagEditModalShown;
					}}
				/>
			</div>
		</div>
	</div>

	<PosterList>
		{#if dataLoader.state.data?.length > 0}
			{#each dataLoader.state.data as w, i (`${i}-${w.type}`)}
				{#if w}
					<Poster
						watched={dataLoader.state.data[i].watched}
						media={w}
						fluidSize={true}
					/>
				{/if}
			{/each}
		{:else if !dataLoader.state.reqLoading && !dataLoader.state.reqLoadError}
			<div class="empty-list">
				<Icon i="ticket" wh={80} />
				<h2 class="norm">This tag is empty!</h2>
				<h4 class="norm">
					{`${store.hasActiveFilters ? "Try removing your active filters or a" : "A"}`}dd
					entries to this tag via its content page.
				</h4>
				{#if store.hasActiveFilters}
					<button onclick={() => clearActiveFilters()}>Clear Filters</button>
				{/if}
			</div>
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
					dataLoader.runFn();
				}}
			/>
		</div>
	{/if}
{:else}
	<div class="content">
		<div class="inner">
			<strong>Tag does not exist!</strong>
		</div>
	</div>
{/if}

{#if tagEditModalShown}
	<CreateTagModal
		existingTag={tag}
		onClose={() => (tagEditModalShown = false)}
	/>
{/if}

<style lang="scss">
	.content {
		display: flex;
		width: 100%;
		justify-content: center;

		.inner {
			display: flex;
			flex-flow: column;
			gap: 5px;
			justify-content: center;
			align-items: center;
			width: 100%;
			max-width: 1200px;
			margin: 20px 30px;
			margin-top: 0;
		}
	}

	.empty-list {
		display: flex;
		flex-flow: column;
		gap: 5px;
		align-items: center;
		max-width: 400px;

		h2 {
			margin-top: 10px;
		}

		h4 {
			font-weight: normal;
			text-align: center;
		}

		button {
			width: max-content;
			padding-left: 20px;
			padding-right: 20px;
			margin-top: 15px;
		}
	}

	.basic-ctr {
		display: flex;
		align-items: center;
		justify-content: center;
		flex-wrap: wrap;
		gap: 10px;
		max-width: 300px;
		width: 100%;
		fill: $text-color;
	}
</style>
