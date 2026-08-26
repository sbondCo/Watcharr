<script lang="ts">
	import { goto } from "$app/navigation";
	import { resolve } from "$app/paths";
	import Icon from "@/lib/Icon.svelte";
	import Modal from "@/lib/Modal.svelte";
	import Poster from "@/lib/poster/Poster.svelte";
	import PosterList from "@/lib/poster/PosterList.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import SpinnerTiny from "@/lib/SpinnerTiny.svelte";
	import { req } from "@/lib/util/api";
	import { notify } from "@/lib/util/notify";
	import {
		getContentTypeFromMedia,
		type ContentType,
		type Media,
		type ImportMapping,
	} from "@/types";

	let mappings: ImportMapping[] = $state([]);
	let isLoading = $state(true);
	let loadError = $state("");

	// The mapping currently being repointed at different content, if any.
	let changing: ImportMapping | undefined = $state(undefined);
	let searchQuery = $state("");
	let searchResults: Media[] = $state([]);
	let isSearching = $state(false);
	// Ids of mappings with a request in flight, so their buttons can be
	// disabled individually rather than locking the whole page.
	let busyIds: number[] = $state([]);

	async function load() {
		isLoading = true;
		loadError = "";
		try {
			mappings = await req.get<ImportMapping[]>("/import/mappings");
		} catch (err) {
			console.error("mappings: failed to load", err);
			loadError = "Couldn't load your saved matches.";
		}
		isLoading = false;
	}

	load();

	function contentLink(m: ImportMapping): string | undefined {
		if (m.tmdbId) {
			return m.type === "movie" ? `/movie/${m.tmdbId}` : `/tv/${m.tmdbId}`;
		}
		if (m.igdbId) {
			return `/game/${m.igdbId}`;
		}
		return undefined;
	}

	async function forget(m: ImportMapping) {
		busyIds = [...busyIds, m.id];
		try {
			await req.delete(`/import/mappings/${m.id}`);
			mappings = mappings.filter((x) => x.id !== m.id);
			notify({ type: "success", text: `Forgot the match for ${m.name}` });
		} catch (err) {
			console.error("mappings: failed to forget", err);
			notify({
				type: "error",
				text: `Couldn't forget the match for ${m.name}`,
			});
		}
		busyIds = busyIds.filter((id) => id !== m.id);
	}

	// Import content types line up with content types, apart from episodes,
	// which are searched for as their show.
	function mappingContentType(t: string): ContentType | undefined {
		if (t === "tv_episode") {
			return "tv";
		}
		if (t === "movie" || t === "tv" || t === "game") {
			return t;
		}
		return undefined;
	}

	async function runSearch() {
		if (!searchQuery) {
			return;
		}
		isSearching = true;
		searchResults = [];
		try {
			const resp = await req.get<{ results: Media[] }>(
				`/search?query=${encodeURIComponent(searchQuery)}`,
			);
			// A mapping keeps the content type it was saved with, so only
			// offer results of that same type. Picking a different type would
			// leave the row keyed as one type while holding another types id,
			// and lookups are done on the type, so it would never match.
			const wanted = changing ? mappingContentType(changing.type) : undefined;
			searchResults = (resp?.results ?? []).filter((r) => {
				const t = getContentTypeFromMedia(r);
				// People, and anything we can't type, can't be mapped to.
				if (!t) {
					return false;
				}
				return !wanted || t === wanted;
			});
		} catch (err) {
			console.error("mappings: search failed", err);
			notify({ type: "error", text: "Search failed" });
		}
		isSearching = false;
	}

	async function pick(m: ImportMapping, media: Media) {
		busyIds = [...busyIds, m.id];
		try {
			const updated = await req.put<ImportMapping>(`/import/mappings/${m.id}`, {
				tmdbId: media.ids.tmdb ?? 0,
				igdbId: media.ids.igdb ?? 0,
			});
			mappings = mappings.map((x) => (x.id === m.id ? updated : x));
			notify({ type: "success", text: `Updated the match for ${m.name}` });
			changing = undefined;
		} catch (err) {
			console.error("mappings: failed to update", err);
			notify({
				type: "error",
				text: `Couldn't update the match for ${m.name}`,
			});
		}
		busyIds = busyIds.filter((id) => id !== m.id);
	}

	function startChanging(m: ImportMapping) {
		changing = m;
		searchQuery = m.name;
		searchResults = [];
		runSearch();
	}
</script>

<svelte:head>
	<title>Saved Import Matches</title>
</svelte:head>

<div class="content">
	<div class="inner">
		<h2>Saved Import Matches</h2>
		<span class="desc">
			When an import can't work out which content a name refers to, it asks you
			to pick. Your choice is saved here so re-importing the same file doesn't
			ask again. Forget a match to be asked about it next time.
		</span>

		{#if isLoading}
			<Spinner />
		{:else if loadError}
			<h3>{loadError}</h3>
		{:else if mappings.length <= 0}
			<h3 class="empty">No saved matches yet.</h3>
			<span class="desc">
				They're created as you resolve prompts during an import.
			</span>
		{:else}
			<div class="table-wrap">
				<table>
					<thead>
						<tr>
							<th>Name in import file</th>
							<th>Type</th>
							<th>Matched to</th>
							<th></th>
						</tr>
					</thead>
					<tbody>
						{#each mappings as m (m.id)}
							<tr>
								<td class="name">{m.name}</td>
								<td>{m.type}</td>
								<td>
									{#if contentLink(m)}
										<a href={contentLink(m)}>
											{m.tmdbId ? `tmdb ${m.tmdbId}` : `igdb ${m.igdbId}`}
										</a>
									{:else}
										<span class="unknown">nothing</span>
									{/if}
								</td>
								<td class="row-btns">
									{#if busyIds.includes(m.id)}
										<SpinnerTiny />
									{:else}
										<button onclick={() => startChanging(m)}>Change</button>
										<button onclick={() => forget(m)}>Forget</button>
									{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}

		<div class="btns">
			<button onclick={() => goto(resolve("/import"))}>
				<Icon i="arrow" />Back
			</button>
		</div>
	</div>
</div>

{#if changing}
	<Modal
		title="Change Match"
		desc="Pick what {changing.name} should import as"
		onClose={() => (changing = undefined)}
	>
		<div class="search-row">
			<input
				class="plain"
				placeholder="Search"
				bind:value={searchQuery}
				onkeydown={(e) => {
					if (e.key === "Enter") {
						runSearch();
					}
				}}
			/>
			<button onclick={() => runSearch()}>Search</button>
		</div>
		{#if isSearching}
			<Spinner />
		{:else if searchResults.length <= 0}
			<h4 class="norm">No results</h4>
		{:else}
			<PosterList type="vertical">
				{#each searchResults as r (r.ids)}
					<Poster
						media={r}
						small={true}
						disableInteraction={true}
						hideButtons={true}
						onClick={() => (changing ? pick(changing, r) : undefined)}
					/>
				{/each}
			</PosterList>
		{/if}
	</Modal>
{/if}

<style lang="scss">
	.content {
		display: flex;
		justify-content: center;
		margin: 0 30px;
	}

	.inner {
		display: flex;
		flex-flow: column;
		max-width: 900px;
		width: 100%;

		h2 {
			margin-top: 20px;
		}

		.desc {
			margin-bottom: 15px;
			font-size: 14px;
			opacity: 0.7;
		}

		.empty {
			margin-top: 20px;
		}
	}

	.table-wrap {
		overflow-x: auto;
	}

	table {
		width: 100%;
		border-collapse: collapse;

		th {
			text-align: left;
			font-size: 14px;
			opacity: 0.7;
			padding-bottom: 5px;
		}

		td {
			padding: 6px 8px 6px 0;
			vertical-align: middle;
		}

		.name {
			font-weight: bold;
			word-break: break-word;
		}

		.unknown {
			opacity: 0.6;
		}
	}

	.row-btns {
		display: flex;
		gap: 5px;
		justify-content: flex-end;

		button {
			width: max-content;
		}
	}

	.search-row {
		display: flex;
		gap: 5px;
		margin-bottom: 10px;

		button {
			width: max-content;
		}
	}

	.btns {
		display: flex;
		flex-flow: row;
		margin: 20px 0;
		gap: 5px;

		button {
			width: max-content;
			gap: 3px;
		}
	}
</style>
