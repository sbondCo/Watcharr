<script lang="ts">
	import Activity from "@/lib/Activity.svelte";
	import Error from "@/lib/Error.svelte";
	import HorizontalList from "@/lib/HorizontalList.svelte";
	import Icon from "@/lib/Icon.svelte";
	import PersonPoster from "@/lib/poster/PersonPoster.svelte";
	import SeasonsList from "@/lib/season/SeasonsList.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import ProvidersList from "@/lib/content/ProvidersList.svelte";
	import SimilarContent from "@/lib/content/SimilarContent.svelte";
	import Title from "@/lib/content/Title.svelte";
	import {
		contentExistsOnJellyfin,
		removeWatched,
		updateWatched,
	} from "@/lib/util/api";
	import { getTopCrew } from "@/lib/util/helpers.js";
	import { store } from "@/store.svelte.js";
	import type {
		Media,
		TMDBContentCredits,
		TMDBContentCreditsCrew,
		WatchedStatus,
	} from "@/types";
	import axios from "axios";
	import RequestShow from "@/lib/request/RequestShow.svelte";
	import FollowedThoughts from "@/lib/content/FollowedThoughts.svelte";
	import ArrRequestButton from "@/lib/request/ArrRequestButton.svelte";
	import tooltip from "@/lib/actions/tooltip.js";
	import AddToTagButton from "@/lib/tag/AddToTagButton.svelte";
	import PageBackdrop from "@/lib/generic/PageBackdrop.svelte";
	import MyReview from "@/lib/content/MyReview.svelte";
	import ViewTrailerButton from "@/lib/content/ViewTrailerButton.svelte";
	import PosterImage from "@/lib/content/PosterImage.svelte";
	import ExpandableText from "@/lib/content/ExpandableText.svelte";

	let { data } = $props();

	let requestModalShown = $state(false);
	let jellyfinUrl: string | undefined = $state();
	let arrRequestButtonComp: ArrRequestButton | undefined = $state();
	let show: Media | undefined = $state();
	let pageError: Error | undefined = $state();

	$effect(() => {
		(async () => {
			try {
				show = undefined;
				pageError = undefined;
				if (!data.tvId) {
					return;
				}
				const resp = (
					await axios.get(`/content/tv/${data.tvId}`, {
						params: { region: store.userSettings?.country },
					})
				).data as Media;
				if (resp) {
					if (resp.name && resp.ids.tmdb) {
						contentExistsOnJellyfin("tv", resp.name, resp.ids.tmdb).then(
							(j) => {
								if (j?.hasContent && j?.url !== "") {
									jellyfinUrl = j.url;
								}
							},
						);
					}
					show = resp;
				} else {
					show = undefined;
				}
			} catch (err: any) {
				show = undefined;
				pageError = err;
			}
		})();
	});

	async function getTvCredits() {
		const credits = (await axios.get(`/content/tv/${data.tvId}/credits`))
			.data as TMDBContentCredits & { topCrew: TMDBContentCreditsCrew[] };
		if (credits.crew?.length > 0) {
			credits.topCrew = getTopCrew(credits.crew);
		}
		return credits;
	}

	async function contentChanged(
		newStatus?: WatchedStatus,
		newRating?: number,
		newThoughts?: string,
		pinned?: boolean,
	) {
		if (!data.tvId) {
			console.error("contentChanged: no tvId");
			return;
		}
		if (!show) {
			console.error("contentChanged: no show");
			return;
		}
		await updateWatched(show.watched, {
			contentId: data.tvId,
			contentType: "tv",
			status: newStatus,
			rating: newRating,
			thoughts: newThoughts,
			pinned: pinned,
		})
			.then((w) => {
				if (show) {
					show.watched = w;
				}
			})
			.catch(() => {
				/* Default handling inside updateWatched is good enough here */
			});
	}

	async function deleteWatched() {
		if (show?.watched) {
			if (await removeWatched(show.watched.id)) {
				show.watched = undefined;
			}
			return;
		}
		console.error("deleteWatched: no wlistItem.. can't delete");
	}
</script>

<svelte:head>
	<title>{show?.name ? `${show.name} - ` : ""}Show</title>
</svelte:head>

{#if pageError}
	<Error pretty="Failed to load tv show!" error={pageError} />
{:else if !show}
	<Spinner />
{:else if Object.keys(show).length > 0}
	{#if show?.extBackdropPath}
		<PageBackdrop
			src={"https://www.themoviedb.org/t/p/w1920_and_h800_multi_faces" +
				show.extBackdropPath}
		/>
	{/if}
	<div>
		<div class="content">
			<div class="details-wrap">
				<div class="details-container">
					<PosterImage
						src={"https://image.tmdb.org/t/p/w500" + show.extPosterPath}
					/>

					<div class="details">
						<Title
							title={show.name}
							homepage={show.homepage}
							releaseDate={show.releaseDate
								? new Date(show.releaseDate)
								: undefined}
							voteAverage={show.rating}
							voteCount={show.ratingCount}
						/>

						<span class="quick-info">
							{#if show.genres && show.genres?.length > 0}
								<div>
									{#each show.genres as g, i}
										<span
											>{g.name}{i !== show.genres.length - 1 ? ", " : ""}</span
										>
									{/each}
								</div>
							{:else}
								<span>Unknown Genres</span>
							{/if}
						</span>

						<ExpandableText text={show.summary} style="margin-bottom: 18px;" />

						<div class="btns">
							<ViewTrailerButton videos={show.videos} />
							{#if jellyfinUrl}
								<a class="btn" href={jellyfinUrl} target="_blank">
									{#if localStorage.getItem("useEmby")}
										<Icon i="emby" wh={14} />Play On Emby
									{:else}
										<Icon i="jellyfin" wh={14} />Play On Jellyfin
									{/if}
								</a>
							{/if}
							{#if store.serverFeatures?.sonarr && data.tvId}
								<ArrRequestButton
									type="tv"
									tmdbId={data.tvId}
									openRequestModal={() =>
										(requestModalShown = !requestModalShown)}
									bind:this={arrRequestButtonComp}
								/>
							{/if}
							{#if show.watched}
								<div class="other-side">
									<AddToTagButton watchedItem={show.watched} />
									<button
										onclick={() => {
											if (show?.watched?.pinned) {
												contentChanged(undefined, undefined, undefined, false);
											} else {
												contentChanged(undefined, undefined, undefined, true);
											}
										}}
										use:tooltip={{
											text: `${show.watched?.pinned ? "Unpin from" : "Pin to"} top of list`,
											pos: "bot",
										}}
									>
										<Icon i={show.watched?.pinned ? "unpin" : "pin"} wh={19} />
									</button>
									<button
										class="delete-btn"
										onclick={() => deleteWatched()}
										use:tooltip={{ text: "Delete", pos: "bot" }}
									>
										<Icon i="trash" wh={19} />
									</button>
								</div>
							{/if}
						</div>

						{#if show.providers}
							<ProvidersList
								providers={show.providers}
								fullListLink={show.providersFullListLink}
								fullListLinkText="JustWatch"
							/>
						{/if}
					</div>
				</div>
			</div>

			<MyReview
				watched={show.watched}
				contentTitle={show.name}
				onRatingChanged={(n) => contentChanged(undefined, n)}
				onStatusChanged={(n) => contentChanged(n)}
				onThoughtsChanged={(newThoughts) => {
					return contentChanged(undefined, undefined, newThoughts);
				}}
			/>
		</div>

		{#if requestModalShown}
			<RequestShow
				content={show}
				onClose={(reqResp) => {
					requestModalShown = false;
					if (reqResp) {
						arrRequestButtonComp?.setExistingRequest(reqResp);
					}
				}}
			/>
		{/if}

		<div class="page">
			{#if data.tvId}
				<FollowedThoughts mediaType="tv" mediaId={data.tvId} />
			{/if}

			{#await getTvCredits()}
				<Spinner />
			{:then credits}
				{#if credits.topCrew?.length > 0}
					<div class="creators">
						{#each credits.topCrew as crew}
							<div>
								<span>{crew.name}</span>
								<span>{crew.job}</span>
							</div>
						{/each}
					</div>
				{/if}

				{#if credits.cast?.length > 0}
					<HorizontalList title="Cast">
						{#each credits.cast?.slice(0, 50) as cast}
							<PersonPoster
								id={cast.id}
								name={cast.name}
								path={cast.profile_path}
								role={cast.character}
								zoomOnHover={false}
							/>
						{/each}
					</HorizontalList>
				{/if}
			{:catch err}
				<Error error={err} pretty="Failed to load cast!" />
			{/await}

			{#if show.similar}
				<SimilarContent similar={show.similar} />
			{/if}

			{#if show.watched}
				<Activity bind:activity={show.watched.activity} />
			{/if}

			{#if data?.tvId && show.seasons}
				<SeasonsList
					tvId={data.tvId}
					seasons={show.seasons}
					watchedItem={show.watched}
					lastViewedSeason={show.watched?.lastViewedSeason}
					lastViewedSeasonChanged={(wid, lvs) => {
						if (show?.watched && show.watched.id === wid) {
							show.watched.lastViewedSeason = lvs;
						}
					}}
				/>
			{/if}
		</div>
	</div>
{:else}
	Show not found
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

				a.btn,
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

				.delete-btn {
					&:hover {
						color: $error;
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

	.creators {
		display: flex;
		flex-wrap: wrap;
		justify-content: center;
		gap: 35px;
		margin: 10px 60px;

		div {
			display: flex;
			flex-flow: column;
			min-width: 150px;

			span:first-child {
				font-weight: bold;
			}
		}
	}
</style>
