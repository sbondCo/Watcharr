<script lang="ts">
	import type {
		TMDBSeasonDetails,
		TMDBShowSeason,
		Watched,
		WatchedSeason,
		WatchedSeasonAddResponse,
		WatchedStatus,
	} from "@/types";
	import axios from "axios";
	import Spinner from "./Spinner.svelte";
	import Error from "./Error.svelte";
	import SeasonsListEpisode from "./SeasonsListEpisode.svelte";
	import PosterStatus from "./poster/PosterStatus.svelte";
	import { notify } from "./util/notify";
	import { store } from "@/store.svelte";
	import PosterRating from "./poster/PosterRating.svelte";
	import Icon from "./Icon.svelte";
	import { watchedStatuses } from "./util/helpers";

	interface Props {
		tvId: number;
		seasons: TMDBShowSeason[];
		watchedItem: Watched | undefined;
	}

	let { tvId, seasons, watchedItem = $bindable() }: Props = $props();

	let activeSeason = $state(
		typeof watchedItem?.lastViewedSeason === "number"
			? watchedItem?.lastViewedSeason
			: 1,
	);
	let seasonDetailsReq: Promise<TMDBSeasonDetails> = $derived(
		sdr(activeSeason),
	);

	async function sdr(seasonNum: number) {
		const resp = await axios.get(`/content/tv/${tvId}/season/${seasonNum}`, {
			params: {
				watchedId: watchedItem?.id,
			},
		});
		try {
			if (watchedItem?.id) {
				// If we sent a watched id, expect a 'watcharr-lastviewedseason-saved' header in the response.
				const hVal = resp.headers["watcharr-lastviewedseason-saved"];
				if (hVal) {
					watchedItem.lastViewedSeason = seasonNum;
					// watchedList.update((w) => w);
				} else {
					console.error(
						"SeasonList: sdr: No header in response indicating that the lastviewedseason was saved.",
					);
					notify({
						type: "error",
						text: "Failed when saving last viewed season",
					});
				}
			}
		} catch (err) {
			console.error(
				"SeasonList: sdr: Failed to process lastviewedseason-saved header.",
				err,
			);
		}
		return resp.data as TMDBSeasonDetails;
	}

	// Add/update watched season
	function updateWatchedSeason(
		seasonNumber: number,
		status?: WatchedStatus,
		rating?: number,
	) {
		if (!watchedItem) {
			console.error("updateWatchedSeason: No watched item.");
			return;
		}
		const nid = notify({ text: `Saving`, type: "loading" });
		axios
			.post<WatchedSeasonAddResponse>(`/watched/season`, {
				watchedId: watchedItem.id,
				seasonNumber: seasonNumber,
				status,
				rating,
			})
			.then((r) => {
				// const wList = get(watchedList);
				const wEntry = store.watchedList.find((w) => w.id === watchedItem.id);
				if (!wEntry) {
					notify({
						id: nid,
						text: `Request succeeded, but failed to find local data. Please refresh.`,
						type: "error",
					});
					return;
				}
				if (r.status === 200) {
					wEntry.watchedSeasons = r.data.watchedSeasons;
					if (wEntry.activity?.length > 0) {
						wEntry.activity.push(r.data.addedActivity);
					} else {
						wEntry.activity = [r.data.addedActivity];
					}
					notify({ id: nid, text: `Saved!`, type: "success" });
				}
			})
			.catch((err) => {
				console.error(err);
				notify({ id: nid, text: "Failed To Update!", type: "error" });
			});
	}

	function handleStatusClick(
		type: WatchedStatus | "DELETE",
		seasonNumber: number,
	) {
		if (!watchedItem) {
			console.error("handleStatusClick: No watched item.");
			return;
		}
		if (type === "DELETE") {
			const ws = watchedItem.watchedSeasons?.find(
				(s) => s.seasonNumber === seasonNumber,
			);
			if (!ws) {
				notify({
					text: "Failed to find watched season id. Please try refreshing.",
					type: "error",
				});
				return;
			}
			const nid = notify({ text: `Saving`, type: "loading" });
			axios
				.delete(`/watched/season/${ws.id}`)
				.then((r) => {
					const wEntry = store.watchedList.find((w) => w.id === watchedItem.id);
					if (!wEntry) {
						notify({
							id: nid,
							text: `Request succeeded, but failed to find local data. Please refresh.`,
							type: "error",
						});
						return;
					}
					if (r.status === 200) {
						wEntry.watchedSeasons = wEntry.watchedSeasons?.filter(
							(s) => s.id !== ws.id,
						);
						if (r.data) {
							if (wEntry.activity?.length > 0) {
								wEntry.activity.push(r.data);
							} else {
								wEntry.activity = [r.data];
							}
						}
						notify({ id: nid, text: `Removed!`, type: "success" });
					}
				})
				.catch((err) => {
					console.error(err);
					notify({ id: nid, text: "Failed To Remove!", type: "error" });
				});
			return;
		}
		updateWatchedSeason(seasonNumber, type);
	}

	function handleStarClick(rating: number, seasonNumber: number) {
		updateWatchedSeason(seasonNumber, undefined, rating);
	}

	function checkSeasonStatus(
		watchedSeasons: WatchedSeason[] | undefined,
		currentSeason: TMDBShowSeason,
	): WatchedStatus | undefined {
		if (watchedSeasons) {
			const watchedSeason = watchedSeasons.find(
				(ws) => ws.seasonNumber === currentSeason.season_number,
			);
			console.debug(
				"checkSeasonStatus:",
				currentSeason.season_number,
				watchedSeason?.status,
			);
			return watchedSeason?.status;
		}
		return undefined;
	}
</script>

<div class="ctr">
	<ul class="seasons">
		{#each seasons as season}
			<button
				class={`plain${activeSeason === season.season_number ? " active" : ""}`}
				onclick={() => {
					activeSeason = season.season_number;
				}}
			>
				<div>
					<h1 class="season-name">{season.name}</h1>
					{#if season.episode_count > 0}
						<h2 class="season-episodes">{season.episode_count} Episodes</h2>
					{/if}
				</div>
				<div>
					{#if season.air_date}
						<h2 class="season-date">
							{new Date(Date.parse(season.air_date)).getFullYear()}
						</h2>
					{:else if season.season_number > 0}
						<h2>TBD</h2>
					{/if}

					{#if watchedItem}
						{@const status = checkSeasonStatus(
							watchedItem.watchedSeasons,
							season,
						)}
						{#if status}
							<div class="plain season-status">
								<Icon i={watchedStatuses[status]} />
							</div>
						{/if}
					{/if}
				</div>
			</button>
		{/each}
		<div class="last"></div>
	</ul>

	<div class="episodes">
		{#await seasonDetailsReq}
			<Spinner />
		{:then season}
			<div class="episodes-topbar">
				<h3>{season.name}</h3>
				{#if watchedItem}
					{@const ws = watchedItem?.watchedSeasons?.find(
						(s) => s.seasonNumber === season.season_number,
					)}
					{#if ws}
						<div class="rating">
							<PosterRating
								rating={ws?.rating}
								btnTooltip="Season Rating"
								handleStarClick={(r) =>
									handleStarClick(r, season.season_number)}
								minimal={true}
								direction="bot"
							/>
						</div>
					{/if}
					<div class="status">
						<PosterStatus
							status={ws?.status}
							btnTooltip="Season Status"
							handleStatusClick={(t) =>
								handleStatusClick(t, season.season_number)}
							direction="bot"
							width="100%"
							small
						/>
					</div>
				{/if}
			</div>
			{#if season?.episodes?.length > 0}
				<ul>
					{#each season.episodes as ep}
						<SeasonsListEpisode {ep} {watchedItem} />
					{/each}
				</ul>
			{:else}
				<h3 class="norm">No episodes in this season yet!</h3>
			{/if}
		{:catch err}
			<Error pretty="Failed to load season details!" error={err} />
		{/await}
	</div>
</div>

<style lang="scss">
	.ctr {
		display: flex;
		flex-flow: row;
		gap: 20px;
		width: 100%;
	}

	.episodes {
		width: 100%;

		ul {
			display: flex;
			flex-flow: column;
			list-style: none;
			gap: 20px;
		}
	}

	ul.seasons {
		display: flex;
		flex-flow: column;
		list-style: none;
		gap: 8px;
		min-width: fit-content;
		height: 100vh;
		overflow: auto;
		position: sticky;
		top: 0px;
		transition: top 200ms ease-in-out;

		button {
			display: flex;
			flex-flow: row;
			gap: 0 18px;
			align-items: center;
			border: 2px solid #302d2d;
			border-radius: 8px;
			padding: 4px 8px;
			cursor: pointer;
			min-width: 160px;
			max-width: 220px;
			transition: background-color 100ms ease;

			& > div {
				&:first-of-type {
					display: flex;
					flex-flow: column;
				}

				&:last-of-type {
					display: flex;
					flex-flow: column;
					margin-left: auto;
					margin-bottom: auto;
					padding-top: 5px;
				}
			}

			&:first-of-type {
				margin-top: 10px;
			}

			.season-name {
				text-align: left;
			}

			.season-name,
			.season-episodes {
				margin-right: auto;
			}

			.season-date,
			.season-status {
				margin-left: auto;
			}

			.season-episodes {
				color: $text-color-accent;
			}

			.season-status {
				fill: $text-color;

				:global(svg) {
					width: 20px;
					height: 20px;
				}
			}

			h1 {
				font-size: 18px;
			}

			h2 {
				font-size: 12px;
			}

			h1,
			h2 {
				font-family: sans-serif;
			}

			&:hover,
			&.active {
				color: $bg-color;
				background-color: $text-color;

				.season-status {
					fill: $bg-color;
				}

				.season-episodes {
					color: $bg-color-accent;
				}
			}

			&.active {
				position: sticky;
				top: 10px;
				bottom: 10px;

				.season-status {
					fill: $bg-color;
				}
			}
		}

		/* hack to get extra scroll space under last el */
		.last {
			padding: 1px;
		}
	}

	.episodes-topbar {
		display: flex;
		align-items: center;
		gap: 10px;
		margin-bottom: 20px;
		min-height: 40px;

		div {
			transition: width 100ms ease;

			&:first-of-type {
				margin-left: auto;
			}

			&.rating {
				height: 40px;
				min-height: 40px;
			}

			&.status {
				width: 45px;
				min-height: 40px;
				height: 40px;
				overflow: visible;
			}
		}
	}

	@media screen and (min-width: 960px) {
		:global(body.nav-shown) ul.seasons {
			top: $nav-height;
			height: calc(100vh - $nav-height);
		}
	}

	@media screen and (max-width: 960px) {
		.ctr {
			flex-flow: column;
		}

		ul.seasons {
			flex-flow: row;
			flex-wrap: wrap;
			position: unset;
			height: unset;
			justify-content: center;

			button {
				&:first-of-type {
					margin-top: unset;
				}
				&.active {
					position: unset;
				}
			}
		}
	}
</style>
