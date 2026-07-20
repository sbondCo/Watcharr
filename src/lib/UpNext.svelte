<script lang="ts">
	import axios from "axios";
	import { onMount } from "svelte";
	import { goto } from "$app/navigation";
	import { notify } from "@/lib/util/notify";
	import HorizontalList from "@/lib/HorizontalList.svelte";
	import Icon from "@/lib/Icon.svelte";

	interface UpNextItem {
		watchedId: number;
		tmdbId: number;
		showTitle: string;
		posterPath: string;
		seasonNumber: number;
		episodeNumber: number;
		episodeName: string;
		stillPath: string;
		airDate: string;
	}

	let items = $state<UpNextItem[]>([]);
	let loading = $state(true);
	let busyId = $state<number | undefined>(undefined);

	async function load() {
		try {
			const r = await axios.get<UpNextItem[]>("/watched/upnext");
			items = r.data ?? [];
		} catch (err) {
			console.error("UpNext: load failed", err);
		} finally {
			loading = false;
		}
	}

	async function markWatched(item: UpNextItem) {
		busyId = item.watchedId;
		const nid = notify({
			text: `Marking ${item.showTitle} S${item.seasonNumber}E${item.episodeNumber}`,
			type: "loading",
		});
		try {
			await axios.post("/watched/episode", {
				watchedId: item.watchedId,
				seasonNumber: item.seasonNumber,
				episodeNumber: item.episodeNumber,
				status: "FINISHED",
			});
			notify({ id: nid, text: "Marked as watched!", type: "success" });
			await load(); // refresh: the card advances to the next episode (or drops off)
		} catch (err) {
			console.error("UpNext: markWatched failed", err);
			notify({ id: nid, text: "Failed to mark as watched!", type: "error" });
		} finally {
			busyId = undefined;
		}
	}

	function img(item: UpNextItem) {
		const p = item.stillPath || item.posterPath;
		return p ? `https://image.tmdb.org/t/p/w300${p}` : "";
	}

	onMount(load);
</script>

{#if !loading && items.length > 0}
	<HorizontalList title="Up Next">
		{#each items as item (item.watchedId)}
			<li class="up-next-card">
				<button
					class="thumb"
					onclick={() => goto(`/tv/${item.tmdbId}`)}
					title={item.showTitle}
				>
					{#if img(item)}
						<img src={img(item)} alt="" />
					{:else}
						<div class="noimg"><Icon i="reel" wh={30} /></div>
					{/if}
				</button>
				<div class="meta">
					<span class="show">{item.showTitle}</span>
					<span class="ep">
						S{item.seasonNumber}E{item.episodeNumber}{item.episodeName
							? ` · ${item.episodeName}`
							: ""}
					</span>
					{#if item.airDate}
						{@const future = new Date(item.airDate) > new Date()}
						<span class="air" class:future title="TV broadcast date">
							{future ? "📺 Airs " : ""}{new Date(item.airDate).toLocaleDateString()}
						</span>
					{/if}
				</div>
				<button
					class="mark"
					disabled={busyId === item.watchedId}
					onclick={() => markWatched(item)}
				>
					✓ Watched
				</button>
			</li>
		{/each}
	</HorizontalList>
{/if}

<style lang="scss">
	.up-next-card {
		display: flex;
		flex-flow: column;
		width: 220px;
		min-width: 220px;
		gap: 6px;
	}
	.thumb {
		padding: 0;
		border: none;
		background: none;
		cursor: pointer;
		width: 100%;
		aspect-ratio: 16 / 9;
		border-radius: 8px;
		overflow: hidden;
	}
	.thumb img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.noimg {
		width: 100%;
		height: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
		background: rgba(0, 0, 0, 0.2);
	}
	.meta {
		display: flex;
		flex-flow: column;
	}
	.show {
		font-weight: bold;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.ep {
		font-size: 12px;
		opacity: 0.8;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.air {
		font-size: 11px;
		opacity: 0.55;
	}
	.air.future {
		opacity: 1;
		font-weight: 600;
		color: #e0a13c;
	}
	.mark {
		display: flex;
		align-items: center;
		gap: 4px;
		justify-content: center;
		cursor: pointer;
	}
</style>
