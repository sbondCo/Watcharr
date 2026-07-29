<script lang="ts">
	import type {
		MediaSeason,
		TMDBSeasonDetails,
		Watched,
		WatchedSeason,
		WatchedStatus,
	} from "@/types";
	import { req } from "../util/api";
	import Spinner from "@/lib/Spinner.svelte";
	import Error from "@/lib/Error.svelte";
	import SeasonsListEpisode from "./SeasonsListEpisode.svelte";
	import PosterStatus from "@/lib/poster/PosterStatus.svelte";
	import { notify } from "@/lib/util/notify";
	import PosterRating from "@/lib/poster/PosterRating.svelte";
	import Icon from "@/lib/Icon.svelte";
	import { watchedStatuses } from "@/lib/util/helpers";
	import { removeWatchedSeason, updateWatchedSeason } from "./api";
	import { onMount } from "svelte";

	interface Props {
		tvId: number;
		seasons: MediaSeason[];
		watchedItem: Watched | undefined;
		lastViewedSeason: number | undefined;
		lastViewedSeasonChanged: (wid: number, lvs: number) => void;
	}

	let {
		tvId,
		seasons,
		watchedItem,
		lastViewedSeason,
		lastViewedSeasonChanged,
	}: Props = $props();

	let activeSeason = $state(lastViewedSeason ?? 1);
	let seasonDetailsReq: Promise<TMDBSeasonDetails> | undefined = $state();

	onMount(() => {
		console.debug("SeasonsList: Mounted.");
		seasonDetailsReq = sdr(activeSeason);
	});

	async function sdr(seasonNum: number) {
		const wid = watchedItem?.id;
		console.debug("SeasonList: sdr: Called.", tvId, seasonNum, wid);
		const resp = await req.getWhole<TMDBSeasonDetails>(
			`/content/tv/${tvId}/season/${seasonNum}`,
			{
				params: {
					watchedId: wid,
				},
			},
		);
		try {
			if (wid) {
				// If we sent a watched id, expect a 'watcharr-lastviewedseason-saved' header in the response.
				const hVal = resp.headers.get("watcharr-lastviewedseason-saved");
				if (hVal) {
					lastViewedSeasonChanged(wid, seasonNum);
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
		return resp.body;
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
			removeWatchedSeason(watchedItem, ws.id);
			return;
		}
		updateWatchedSeason(watchedItem, seasonNumber, { status: type });
	}

	function handleStarClick(rating: number, seasonNumber: number) {
		if (!watchedItem) {
			console.error("handleStarClick: No watchedItem!");
			return;
		}
		updateWatchedSeason(watchedItem, seasonNumber, { rating });
	}

	function checkSeasonStatus(
		watchedSeasons: WatchedSeason[] | undefined,
		currentSeason: MediaSeason,
	): WatchedStatus | undefined {
		if (watchedSeasons) {
			const watchedSeason = watchedSeasons.find(
				(ws) => ws.seasonNumber === currentSeason.number,
			);
			console.debug(
				"checkSeasonStatus:",
				currentSeason.number,
				watchedSeason?.status,
			);
			return watchedSeason?.status;
		}
		return undefined;
	}
</script>

<div class="ctr">
	<ul class="seasons">
		{#each seasons as season (season.number)}
			<button
				class="plain"
				class:active={activeSeason === season.number}
				onclick={() => {
					console.debug("SeasonsList: Season button pressed", season.number);
					activeSeason = season.number;
					seasonDetailsReq = sdr(activeSeason);
				}}
			>
				<div>
					<h1 class="season-name">{season.name}</h1>
					<h2 class="season-episodes">
						{#if season.episodeCount > 0}
							{season.episodeCount} Episodes
						{:else}
							No Episodes Yet
						{/if}
					</h2>
				</div>
				<div>
					{#if season.releaseDate}
						<h2 class="season-date">
							{new Date(Date.parse(season.releaseDate)).getFullYear()}
						</h2>
					{:else if season.number > 0}
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

	{#if seasonDetailsReq}
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
						{#each season.episodes as ep (ep.id)}
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
	{:else}
		<h3 class="norm">Try selecting a season.</h3>
	{/if}
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

		:global(body.nav-shown) ul.seasons {
			top: $nav-height;
		}

		ul.seasons {
			flex-flow: row;
			flex-wrap: nowrap;
			position: unset;
			height: unset;
			justify-content: unset;
			overflow: auto;
			min-width: unset;
			position: sticky;
			top: 0px;
			padding: 10px 0;
			z-index: 5;
			@include nav-blur;

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
