<script lang="ts">
	import PersonPoster from "@/lib/poster/PersonPoster.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import HorizontalList from "@/lib/HorizontalList.svelte";
	import { req, updateWatched } from "@/lib/util/api";
	import { store } from "@/store.svelte";
	import { MediaTypeE, type Media, type WatchedStatus } from "@/types";
	import Activity from "@/lib/Activity.svelte";
	import Title from "@/lib/content/Title.svelte";
	import ProvidersList from "@/lib/content/ProvidersList.svelte";
	import Icon from "@/lib/Icon.svelte";
	import SimilarContent from "@/lib/content/SimilarContent.svelte";
	import Error from "@/lib/Error.svelte";
	import FollowedThoughts from "@/lib/content/FollowedThoughts.svelte";
	import tooltip from "@/lib/actions/tooltip.js";
	import AddToTagButton from "@/lib/tag/AddToTagButton.svelte";
	import PageBackdrop from "@/lib/generic/PageBackdrop.svelte";
	import MyReview from "@/lib/content/MyReview.svelte";
	import PosterImage from "@/lib/content/PosterImage.svelte";
	import WatchedDeleteBtn from "@/lib/content/WatchedDeleteBtn.svelte";

	let { data } = $props();

	let book: Media | undefined = $state();
	let pageError: Error | undefined = $state();

	$effect(() => {
		(async () => {
			try {
				book = undefined;
				pageError = undefined;
				if (!data.bookId) {
					return;
				}
				const resp = (
					await req.get<Media>(`/book/${data.bookId}`, {
						params: { region: store.userSettings?.country },
					})
				);
				if (resp) {
					book = resp;
				} else {
					book = undefined;
				}
			} catch (err: any) {
				book = undefined;
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
			if (!data.bookId) {
				console.error("contentChanged: no bookId");
				return false;
			}
			if (!book) {
				console.error("contentChanged: no book");
				return false;
			}
			book.watched = await updateWatched(book.watched, {
				contentId: data.bookId,
				contentType: "book",
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
	<title>{book?.name ? `${book.name} - ` : ""}Book</title>
</svelte:head>

{#if pageError}
	<Error pretty="Failed to load book!" error={pageError} />
{:else if !book}
	<Spinner />
{:else if Object.keys(book).length > 0}
	{#if book?.extBackdropPath}
		<PageBackdrop src={book.extBackdropPath} />
	{/if}
	<div>
		<div class="content">
			<div class="details-wrap">
				<div class="details-container">
					{#if book?.extPosterPath}
						<PosterImage src={`https://covers.openlibrary.org/w/olid/${book.extPosterPath}-M.jpg`} />
					{/if}

					<div class="details">
						<Title
							title={book.name}
							homepage={book.homepage}
							releaseDate={book.releaseDate
								? new Date(book.releaseDate)
								: undefined}
							voteAverage={book.rating}
							voteCount={book.ratingCount}
						/>

						<span class="quick-info">
							{#if book.genres && book.genres?.length > 0}
								<div>
									{#each book.genres as g, i}
										<span>
											{g.name}{i !== book.genres.length - 1 ? ", " : ""}
										</span>
									{/each}
								</div>
							{/if}
						</span>

						<div style="margin-bottom: 18px;">{@html book.summary}</div>

						<div class="btns">
							{#if book.watched}
								<div class="other-side">
									<AddToTagButton watchedItem={book.watched} />
									<button
										onclick={() => {
											if (book?.watched?.pinned) {
												contentChanged(undefined, undefined, undefined, false);
											} else {
												contentChanged(undefined, undefined, undefined, true);
											}
										}}
										use:tooltip={{
											text: `${book.watched?.pinned ? "Unpin from" : "Pin to"} top of list`,
											pos: "bot",
										}}
									>
										<Icon i={book.watched?.pinned ? "unpin" : "pin"} wh={19} />
									</button>
									<WatchedDeleteBtn
										watchedId={book.watched.id}
										mediaName={book.name}
										onDelete={() => {
											if (book) {
												book.watched = undefined;
											}
										}}
									/>
								</div>
							{/if}
						</div>

						{#if book.providers}
							<ProvidersList
								providers={book.providers}
								fullListLink={book.providersFullListLink}
								fullListLinkText="JustWatch"
							/>
						{/if}
					</div>
				</div>
			</div>

			<MyReview
				watched={book.watched}
				contentTitle={book.name}
				onRatingChanged={(n) => contentChanged(undefined, n)}
				onStatusChanged={(n) => contentChanged(n)}
				onThoughtsChanged={(newThoughts) => {
					return contentChanged(undefined, undefined, newThoughts);
				}}
			/>
		</div>

		<div class="page">
			{#if data.bookId}
				<FollowedThoughts mediaType="book" mediaId={data.bookId} />
			{/if}

			{#if book.authors?.length > 0}
				<HorizontalList title="Authors">
					{#each book.authors as author}
						<PersonPoster
							id={author.id}
							name={author.name}
							path={undefined}
							zoomOnHover={false}
							mediaType={MediaTypeE.olBookAuthor}
						/>
					{/each}
				</HorizontalList>
			{/if}

			{#if book.similar}
				<SimilarContent similar={book.similar} />
			{/if}

			{#if book.watched}
				<Activity bind:activity={book.watched.activity} />
			{/if}
		</div>
	</div>
{:else}
	Book not found
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
