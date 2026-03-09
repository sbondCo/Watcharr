<script lang="ts">
	import Spinner from "@/lib/Spinner.svelte";
	import HorizontalList from "@/lib/HorizontalList.svelte";
	import { type Media, type WatchedStatus } from "@/types";
	import axios from "axios";
	import Activity from "@/lib/Activity.svelte";
	import Title from "@/lib/content/Title.svelte";
	import Error from "@/lib/Error.svelte";
	import FollowedThoughts from "@/lib/content/FollowedThoughts.svelte";
	import { removeWatched, updateWatched } from "@/lib/util/api.js";
	import tooltip from "@/lib/actions/tooltip.js";
	import Icon from "@/lib/Icon.svelte";
	import AddToTagButton from "@/lib/tag/AddToTagButton.svelte";
	import PageBackdrop from "@/lib/generic/PageBackdrop.svelte";
	import MyReview from "@/lib/content/MyReview.svelte";
	import ViewTrailerButton from "@/lib/content/ViewTrailerButton.svelte";
	import ProvidersList from "@/lib/content/ProvidersList.svelte";
	import PosterImage from "@/lib/content/PosterImage.svelte";
	import Poster from "@/lib/poster/Poster.svelte";
	import ExpandableText from "@/lib/content/ExpandableText.svelte";
	import WatchedDeleteBtn from "@/lib/content/WatchedDeleteBtn.svelte";

	let { data } = $props();

	let game: Media | undefined = $state();
	let pageError: Error | undefined = $state();
	let backdropSrc = $derived.by(() => {
		const base = "https://images.igdb.com/igdb/image/upload/t_1080p/";

		if (game?.extBackdropPath) {
			return base + game.extBackdropPath + ".jpg";
		} else if (game?.extPosterPath) {
			return base + game.extPosterPath + ".jpg";
		}
	});

	$effect(() => {
		(async () => {
			try {
				game = undefined;
				pageError = undefined;
				if (!data.gameId) {
					return;
				}
				const resp = (await axios.get(`/game/${data.gameId}`)).data as Media;
				game = resp;
			} catch (err: any) {
				game = undefined;
				pageError = err;
			}
		})();
	});

	async function contentChanged(
		newStatus?: WatchedStatus,
		newRating?: number,
		newThoughts?: string,
		pinned?: boolean,
	): Promise<boolean> {
		try {
			if (!data.gameId) {
				console.error("contentChanged: no gameId");
				return false;
			}
			if (!game) {
				console.error("contentChanged: no game");
				return false;
			}
			game.watched = await updateWatched(game.watched, {
				contentId: data.gameId,
				contentType: "game",
				status: newStatus,
				rating: newRating,
				thoughts: newThoughts,
				pinned: pinned,
			});
			return true;
		} catch {
			return false;
		}
	}
</script>

<svelte:head>
	<title>{game?.name ? `${game.name} - ` : ""}Game</title>
</svelte:head>

{#if pageError}
	<Error pretty="Failed to load game!" error={pageError} />
{:else if !game}
	<Spinner />
{:else if Object.keys(game).length > 0}
	{#if backdropSrc}
		<PageBackdrop src={backdropSrc} />
	{/if}
	<div>
		<div class="content">
			<div class="details-wrap">
				<div class="details-container">
					<PosterImage
						src={"https://images.igdb.com/igdb/image/upload/t_cover_big/" +
							game.extPosterPath +
							".jpg"}
					/>

					<div class="details">
						<Title
							title={game.name}
							homepage={game.homepage}
							releaseDate={game.releaseDate
								? new Date(game.releaseDate)
								: undefined}
							voteAverage={game.rating}
							voteCount={game.ratingCount}
						/>

						<span class="quick-info">
							{#if game.genres && game.genres?.length > 0}
								<div>
									{#each game.genres as g, i}
										<span
											>{g.name}{i !== game.genres.length - 1 ? ", " : ""}</span
										>
									{/each}
								</div>
							{:else}
								<span>Unknown Genres</span>
							{/if}
							<span></span>
							<div>
								{#if game.gameModes && game.gameModes?.length > 0}
									{#each game.gameModes as g, i}
										<span
											>{g.name}{i !== game.gameModes.length - 1
												? ", "
												: ""}</span
										>
									{/each}
								{:else}
									<span>Unknown Game Modes</span>
								{/if}
							</div>
						</span>

						<ExpandableText text={game.summary} style="margin-bottom: 18px;" />

						<div class="btns">
							<ViewTrailerButton videos={game.videos} />
							{#if game.watched}
								<div class="other-side">
									<AddToTagButton watchedItem={game.watched} />
									<button
										onclick={() => {
											if (game?.watched?.pinned) {
												contentChanged(undefined, undefined, undefined, false);
											} else {
												contentChanged(undefined, undefined, undefined, true);
											}
										}}
										use:tooltip={{
											text: `${game.watched?.pinned ? "Unpin from" : "Pin to"} top of list`,
											pos: "bot",
										}}
									>
										<Icon i={game.watched?.pinned ? "unpin" : "pin"} wh={19} />
									</button>
									<WatchedDeleteBtn
										watchedId={game.watched.id}
										mediaName={game.name}
										onDelete={() => {
											if (game) {
												game.watched = undefined;
											}
										}}
									/>
								</div>
							{/if}
						</div>

						{#if game.providers}
							<ProvidersList providers={game.providers} />
						{/if}
					</div>
				</div>
			</div>

			<MyReview
				watched={game.watched}
				contentTitle={game.name}
				onRatingChanged={(n) => contentChanged(undefined, n)}
				onStatusChanged={(n) => contentChanged(n)}
				onThoughtsChanged={(newThoughts) => {
					return contentChanged(undefined, undefined, newThoughts);
				}}
			/>
		</div>

		<div class="page">
			{#if data.gameId}
				<FollowedThoughts mediaType="game" mediaId={data.gameId} />
			{/if}

			{#if game.similar && game.similar?.length > 0}
				<HorizontalList title="Similar">
					{#each game.similar as g, i}
						<Poster
							media={g}
							bind:watched={game.similar[i].watched}
							small={true}
						/>
					{/each}
				</HorizontalList>
			{/if}

			{#if game.watched}
				<Activity bind:activity={game.watched.activity} />
			{/if}
		</div>
	</div>
{:else}
	<Error error="Game not found" pretty="Game not found" />
{/if}

<style lang="scss">
	@use "../../../../lib/content/page.scss";

	.content {
		position: relative;
		color: white;

		.details-container .details {
			.quick-info {
				display: flex;
				gap: 10px;
				margin-bottom: 8px;
			}

			.btns {
				display: flex;
				flex-flow: row;
				flex-wrap: wrap;
				gap: 8px;
				margin-top: auto;

				button {
					max-width: fit-content;
					overflow: hidden;
					animation: 50ms cubic-bezier(0.86, 0, 0.07, 1) forwards otherbtn;
					white-space: nowrap;
					gap: 6px;
					justify-content: flex-start;
					font-size: 14px;

					@keyframes otherbtn {
						from {
							width: 0px;
						}
						to {
							width: 100%;
						}
					}
				}

				.other-side {
					display: flex;
					flex-flow: row;
					gap: 8px;

					@media screen and (min-width: 900px) {
						margin-left: auto;
					}
				}
			}
		}
	}

	.page {
		display: flex;
		flex-flow: column;
		align-items: center;
		margin-left: auto;
		margin-right: auto;
		gap: 30px;
		padding: 20px 50px;
		max-width: 1200px;

		@media screen and (max-width: 500px) {
			padding: 20px;
		}
	}
</style>
