<script lang="ts">
	import Icon from "@/lib/Icon.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import tooltip from "@/lib/actions/tooltip.js";
	import UserAvatar from "@/lib/img/UserAvatar.svelte";
	import { followUser, unfollowUser } from "@/lib/util/api.js";
	import { clearActiveFilters, store } from "@/store.svelte.js";
	import type { Media, PublicUser } from "@/types.js";
	import axios, { type GenericAbortSignal } from "axios";
	import { onDestroy, onMount, untrack } from "svelte";
	import paginatedLoader, {
		PaginatedLoaderRunFnAction,
	} from "@/lib/util/paginatedLoader.svelte.js";
	import infScroll from "@/lib/util/infScroll.js";
	import { page } from "$app/state";
	import PosterList from "@/lib/poster/PosterList.svelte";
	import Poster from "@/lib/poster/Poster.svelte";
	import Error from "@/lib/Error.svelte";
	import { afterNavigate } from "$app/navigation";

	let meta = $derived.by(() => {
		return {
			id: page.params.id,
			username: page.params.username,
		};
	});

	let followBtnDisabled = $state(false);
	let user: PublicUser | undefined = $state();

	let isFollowing = $derived(
		!!store.follows?.find((f) => f.followedUser.id === Number(meta.id)),
	);

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
		if (!meta.id || !meta.username) {
			console.warn("load: Missing id or username!");
			return;
		}
		const r = await axios.get(`/watched/${meta.id}/${meta.username}`, {
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

	async function getPublicUser() {
		return (await axios.get(`/user/public/${meta.id}/${meta.username}`))
			.data as PublicUser;
	}

	async function follow() {
		followBtnDisabled = true;
		console.log(isFollowing);
		if (isFollowing) {
			await unfollowUser(Number(meta.id));
		} else {
			await followUser(Number(meta.id));
		}
		followBtnDisabled = false;
	}

	$effect(() => {
		user = undefined;
		if (meta?.id && meta?.username) {
			getPublicUser()
				.then((u) => {
					user = u;
				})
				.catch((err) => {
					console.error("getPublicUser failed!", err);
				});
		}
	});

	afterNavigate((e) => {
		if (!e.from?.route?.id?.toLowerCase()?.includes("/lists")) {
			// Ensure afterNavigate can only runFn when we are coming
			// from another list page.
			// OnMount of a list page we don't want to have this also run
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
	<title>{meta.username}'s Watched List</title>
</svelte:head>

<div class="content">
	<div class="inner">
		<UserAvatar img={user?.avatar} />
		<div class="basic-ctr">
			<div class="name-row">
				<h2 title={user?.username}>
					{meta.username}
				</h2>
				<button
					class="plain"
					disabled={followBtnDisabled}
					onclick={follow}
					use:tooltip={{ text: isFollowing ? "Unfollow" : "Follow" }}
				>
					<Icon i={isFollowing ? "person-minus" : "person-add"} />
				</button>
			</div>
			{#if user?.bio}
				<span title={user?.bio}>{user?.bio}</span>
			{/if}
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
					disableInteraction={true}
				/>
			{/if}
		{/each}
	{:else if !dataLoader.state.reqLoading && !dataLoader.state.reqLoadError}
		<div class="empty-list">
			<Icon i={store.hasActiveFilters ? "filter-circle" : "reel"} wh={80} />
			<h2 class="norm">This list is empty!</h2>
			<h4 class="norm">Come back later to see if they have added anything.</h4>
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

<style lang="scss">
	.content {
		display: flex;
		width: 100%;
		justify-content: center;

		.inner {
			display: flex;
			flex-flow: row;
			gap: 15px;
			justify-content: center;
			align-items: center;
			width: 100%;
			max-width: 1200px;
			margin: 20px 30px;
			margin-top: 0;
		}
	}

	button {
		width: max-content;
	}

	textarea {
		border: 0;
		padding: 0;
		resize: none;
		text-overflow: ellipsis;
	}

	.basic-ctr {
		min-width: 200px;
		max-width: 300px;
		overflow: hidden;

		.name-row {
			display: flex;
			flex-flow: row;
			gap: 15px;

			h2 {
				overflow: hidden;
				text-overflow: ellipsis;
			}

			button {
				margin-left: auto;
				fill: $text-color;
			}
		}

		span {
			font-family: monospace;
			overflow: hidden;
			text-overflow: ellipsis;
			display: -webkit-box;
			-webkit-line-clamp: 2;
			-webkit-box-orient: vertical;
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
</style>
