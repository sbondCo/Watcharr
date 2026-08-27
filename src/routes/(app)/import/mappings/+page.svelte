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
				// Only used by the server when the mapping has no type of its
				// own, which happens for ignored names the import file gave
				// no type to.
				type: getContentTypeFromMedia(media),
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

	// Set while the delete all confirmation modal is open.
	let confirmingDeleteAll = $state(false);
	let isDeletingAll = $state(false);

	async function deleteAll() {
		isDeletingAll = true;
		const nid = notify({ text: "Deleting saved matches", type: "loading" });
		try {
			await req.delete("/import/mappings");
			mappings = [];
			notify({ id: nid, text: "All saved matches deleted!", type: "success" });
			confirmingDeleteAll = false;
		} catch (err) {
			console.error("mappings: failed to delete all", err);
			notify({
				id: nid,
				text: "Couldn't delete your saved matches",
				type: "error",
			});
		}
		isDeletingAll = false;
	}

	// Matches first, ignored names after them, each set alphabetical. Sorted
	// here rather than by the server so a row moves as soon as it changes,
	// without reloading the list.
	const sortedMappings = $derived(
		[...mappings].sort((a, b) => {
			if (a.ignored !== b.ignored) {
				return a.ignored ? 1 : -1;
			}
			return a.name.localeCompare(b.name);
		}),
	);

	function startChanging(m: ImportMapping) {
		changing = m;
		searchQuery = m.name;
		searchResults = [];
		runSearch();
	}
</script>

{#snippet matchLink(m: ImportMapping)}
	{#if m.ignored}
		<span class="ignored">ignored</span>
	{:else if m.tmdbId && m.type === "movie"}
		<a href={resolve("/movie/{m.tmdbId}")}>tmdb {m.tmdbId}</a>
	{:else if m.tmdbId}
		<a href={resolve("/tv/{m.tmdbId}")}>tmdb {m.tmdbId}</a>
	{:else if m.igdbId}
		<a href={resolve("/game/{m.igdbId}")}>igdb {m.igdbId}</a>
	{:else}
		<span class="unknown">nothing</span>
	{/if}
{/snippet}

<svelte:head>
	<title>Saved Import Matches</title>
</svelte:head>

<div class="content">
	<div class="inner">
		<h2>Saved Import Matches</h2>
		<span class="desc">
			When an import can't work out which content a name refers to, it asks you
			to pick. Your choice is saved here so re-importing the same file doesn't
			ask again, including names you chose to ignore. Change an ignored name to
			start importing it again, or forget a match to be asked about it next
			time.
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
						{#each sortedMappings as m (m.id)}
							<tr>
								<td class="name">{m.name}</td>
								<td>{m.type ? m.type : "any"}</td>
								<td>
									{@render matchLink(m)}
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
			{#if mappings.length > 0}
				<button onclick={() => (confirmingDeleteAll = true)}>
					Delete All Matches
				</button>
			{/if}
		</div>
	</div>
</div>

{#if confirmingDeleteAll}
	<Modal
		title="Delete All Saved Matches"
		desc="Are you sure you want to forget all {mappings.length} of your saved matches?"
		maxWidth="500px"
		onClose={() => (confirmingDeleteAll = false)}
	>
		<div class="confirm-inner">
			<span class="desc">
				Your watched list isn't touched, only the remembered choices. You'll be
				asked to pick again for these names on your next import.
			</span>
			<button
				class="delete-all-btn"
				onclick={() => deleteAll()}
				disabled={isDeletingAll}
			>
				Yes, delete all saved matches
			</button>
		</div>
	</Modal>
{/if}

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
		<span class="current-match">
			Currently {changing.ignored ? "" : "matched to "}{@render matchLink(
				changing,
			)}. This does not affect your saved list, you must manually add or remove
			entries to modify your saved list.
		</span>
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

		.unknown,
		.ignored {
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

	.confirm-inner {
		display: flex;
		flex-flow: column;
		gap: 10px;

		// The pages .desc styling is scoped to the table view, so the same
		// look is repeated here for the text inside the modal.
		.desc {
			font-size: 14px;
			opacity: 0.7;
		}
	}

	.delete-all-btn {
		width: max-content;
		margin-left: auto;

		&:hover {
			color: $error;
		}
	}

	.current-match {
		display: block;
		font-size: 14px;
		opacity: 0.7;
		margin-bottom: 10px;
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
