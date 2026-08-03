<script lang="ts">
	import type {
		TMDBSeasonDetailsEpisode,
		WatchedStatus,
		Watched,
	} from "@/types";
	import Icon from "../Icon.svelte";
	import PosterRating from "../poster/PosterRating.svelte";
	import PosterStatus from "../poster/PosterStatus.svelte";
	import { notify } from "../util/notify";
	import { store } from "@/store.svelte";
	import { removeWatchedEpisode, updateWatchedEpisode } from "./api";
	import { onMount } from "svelte";
	import { toRelativeDate } from "../util/helpers";

	interface Props {
		ep: TMDBSeasonDetailsEpisode;
		watchedItem: Watched | undefined;
	}

	let { ep, watchedItem }: Props = $props();

	const we = $derived(
		watchedItem?.watchedEpisodes?.find(
			(s) =>
				s.seasonNumber === ep.season_number &&
				s.episodeNumber === ep.episode_number,
		),
	);

	let isHidden: boolean = $state(!!store?.userSettings?.hideSpoilers);

	const airDate: Date | undefined = $derived.by(() => {
		if (!ep.air_date) {
			return;
		}
		const d = new Date(ep.air_date);
		if (isNaN(d.getTime())) {
			return;
		}
		return d;
	});
	const isUnaired: boolean = $derived.by(() => {
		if (!airDate) {
			return false;
		}
		return airDate.getTime() > new Date().getTime();
	});

	/**
	 * Re-sets `isHidden` state.
	 */
	function reSetIsHidden() {
		// If the episode status is "FINISHED", ensure `isHidden` is set to
		// `false` (so finished episodes aren't blurred when hideSpoilers is on).
		if (we?.status == "FINISHED") {
			isHidden = false;
		}
	}

	onMount(() => {
		reSetIsHidden();
	});

	async function handleStatusClick(type: WatchedStatus | "DELETE") {
		if (!watchedItem) {
			console.error("SeasonListEpisode: handleStatusClick: No watched item.");
			return;
		}
		if (type === "DELETE") {
			if (!we || !we.id) {
				notify({
					text: "Failed to find watched episode id. Please try refreshing.",
					type: "error",
				});
				console.error(
					"handleStatusClick(DELETE): `we` doesn't exist or have an id",
					we,
				);
				return;
			}
			removeWatchedEpisode(watchedItem, we.id);
			// NOTE: Similar to below where we `reSetIsHidden` to unhide spoilers
			// automatically if status is set to FINISHED, we WONT do the opposite
			// here and re-hide the spoilers (if unhidden) after removing an episode
			// because that would probably be annoying to users (eg: click to
			// show spoilers, then delete episode, spoilers re-hidden automatically).
			return;
		}
		await updateWatchedEpisode(
			watchedItem,
			ep.season_number,
			ep.episode_number,
			{
				status: type,
			},
		);
		reSetIsHidden();
	}

	function handleStarClick(rating: number) {
		if (!watchedItem) {
			console.error("handleStarClick: No watchedItem!");
			return;
		}
		updateWatchedEpisode(watchedItem, ep.season_number, ep.episode_number, {
			rating,
		});
	}
</script>

<li class={isHidden ? "dont-spoil" : ""}>
	{#if ep.still_path}
		<img
			src={`https://www.themoviedb.org/t/p/w227_and_h127_bestv2/${ep.still_path}`}
			alt=""
		/>
	{:else}
		<div class="no-still"></div>
	{/if}
	<div class="info">
		<div>
			<span>
				<b>{ep.episode_number}</b>
				<span class="episode-name">{ep.name}</span>
				{#if ep.runtime}
					<span
						class="episode-runtime"
						title="This episode has a runtime of {ep.runtime} minutes."
						>{ep.runtime} min</span
					>
				{/if}
				{#if !isUnaired && airDate}
					<!-- Air date for aired episodes. -->
					<span class="episode-air-date">
						{toRelativeDate(airDate)}
					</span>
				{/if}
			</span>
			{#if ep.vote_count <= 0 && isUnaired}
				<!-- If no votes (and subsequently no rating) AND episode is
				 unaired, no need to show the rating star. -->
			{:else}
				<span
					class="rating"
					title={`TMDB Rating: ${ep.vote_average} out of 10 (based on ${ep.vote_count} votes)`}
				>
					<span>*</span>
					{Math.round(ep.vote_average * 10) / 10}
				</span>
			{/if}
		</div>
		{#if isUnaired && airDate}
			<!-- Air date for unaired episodes. -->
			<span class="episode-air-date">
				Airs on {toRelativeDate(airDate)}
			</span>
		{/if}
		<span class="overview">{ep.overview}</span>
	</div>
	{#if watchedItem}
		<div class="status-rating-ctr">
			<div class="rating" style="width: 45px">
				<PosterRating
					rating={we?.rating}
					btnTooltip={`Episode ${ep.episode_number} Rating`}
					handleStarClick={(r) => handleStarClick(r)}
					minimal={true}
					direction="bot"
					hideStarWhenRated
				/>
			</div>
			<div class="status">
				<PosterStatus
					status={we?.status}
					btnTooltip={`Episode ${ep.episode_number} Status`}
					handleStatusClick={(t) => handleStatusClick(t)}
					direction="bot"
					width="100%"
					small
				/>
			</div>
		</div>
	{/if}
	{#if isHidden}
		<button class="plain spoiler-text" onclick={() => (isHidden = false)}>
			<Icon i="eye-closed" wh={34} />
			<span>Click To Reveal</span>
		</button>
	{/if}
</li>

<style lang="scss">
	li {
		display: flex;
		flex-flow: row;
		gap: 8px;
		position: relative;

		img,
		.no-still {
			width: 227px;
			min-width: 227px;
			height: 127px;
			min-height: 127px;
			border-radius: 10px;
			background-color: rgb(0, 0, 0);
			object-fit: fill;

			@media screen and (max-width: 590px) {
				width: 80%;
				height: auto;
			}

			@media screen and (max-width: 450px) {
				width: 100%;
			}
		}

		.info {
			display: flex;
			flex-flow: column;

			& > div {
				display: flex;
				flex-flow: row;
				align-items: center;

				.episode-name,
				.episode-runtime {
					text-transform: lowercase;
					font-variant: small-caps;
					font-weight: bold;
					font-size: 16px;
				}

				.episode-runtime {
					font-size: 14px;
					padding: 0 2px;
				}

				.rating {
					display: flex;
					align-items: start;
					justify-content: center;
					min-width: 43px;
					font-size: 15px;
					color: $rating-color;
					font-weight: bolder;
					overflow: hidden;

					span {
						margin-top: 2px;
						font-family: "Rampart One";
						-webkit-text-stroke: 1px $rating-color;
						font-size: 25px;
						line-height: 0.7;
					}
				}
			}

			.episode-air-date {
				font-size: 14px;
				color: $text-color-accent;
				padding: 0 2px;
				text-transform: lowercase;
				font-variant: small-caps;
				font-weight: bold;
			}
		}

		.status-rating-ctr {
			display: flex;
			align-items: center;
			flex-flow: column-reverse;
			gap: 10px;
			margin-bottom: auto;
			min-height: 40px;
			margin-left: auto;

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

					/* z-index of 2 so the button is higher than .spoiler-text
					   which makes it clickable while whole ep is still hidden.
					   Which is useful if you don't want spoilers while setting
					   the episode to WATCHING, etc. */
					z-index: 2;

					&:hover {
						/* On hover, the z-index is higher than all other status
						   buttons on the page to avoid the active one being put
						   below others (making it unuseable). */
						z-index: 3;
					}
				}
			}
		}

		span {
			padding: 3px 5px;

			@media screen and (max-width: 590px) {
				text-align: center;
			}
		}

		.spoiler-text {
			display: flex;
			flex-flow: column;
			align-items: center;
			justify-content: center;
			gap: 8px;
			position: absolute;
			width: 100%;
			height: 100%;
			font-weight: bolder;
			font-size: 20px;
			fill: $text-color;
			opacity: 0;
			transition:
				visibility 150ms ease-in,
				opacity 150ms ease-in;
			cursor: pointer;

			span {
				text-shadow: 0 0 6px $bg-color;
			}

			:global(svg) {
				filter: drop-shadow(0 0 8px $bg-color);
			}

			&:hover,
			&:active,
			&:focus {
				opacity: 1;
			}
		}

		img,
		.episode-name,
		.rating,
		.overview {
			transition: filter 150ms ease-out;
		}

		&.dont-spoil {
			.episode-name,
			.rating,
			.overview {
				filter: blur(4px);
			}

			img {
				filter: blur(6px);
			}
		}
	}

	@media screen and (max-width: 590px) {
		li {
			align-items: center;
			flex-flow: column;
			width: 100%;
			height: 100%;

			.info {
				align-items: center;
			}

			.status-rating-ctr {
				flex-flow: row;
				justify-content: center;
				margin-left: unset;
			}
		}

		.rating {
			margin-left: auto;
		}
	}
</style>
