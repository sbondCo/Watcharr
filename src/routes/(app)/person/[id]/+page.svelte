<script lang="ts">
	import Error from "@/lib/Error.svelte";
	import Poster from "@/lib/poster/Poster.svelte";
	import PosterList from "@/lib/poster/PosterList.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import DropDown from "@/lib/DropDown.svelte";
	import type {
		Media,
		PersonCreditsResponse,
		PersonDetailsResponse,
	} from "@/types";
	import { req } from "@/lib/util/api.js";
	import Checkbox from "@/lib/Checkbox.svelte";
	import Icon from "@/lib/Icon.svelte";
	import PageBackdrop from "@/lib/generic/PageBackdrop.svelte";
	import PosterImage from "@/lib/content/PosterImage.svelte";
	import ExpandableText from "@/lib/content/ExpandableText.svelte";
	import CountBadge from "@/lib/generic/CountBadge.svelte";
	import tooltip from "@/lib/actions/tooltip.js";

	let { data } = $props();

	let person: PersonDetailsResponse | undefined = $state();
	let pageError: unknown | undefined = $state();
	let sortOption = $state("Vote count");
	let onMyListFilter = $state(false);
	let activeCreditType: string | undefined = $state();

	let credits: PersonCreditsResponse | undefined = $state();
	let creditsError: unknown | undefined = $state();

	$effect(() => {
		if (data.personId) {
			fetchPersonData();
		}
	});

	$effect(() => {
		if (sortOption && credits) {
			sortCredits(sortOption);
		}
	});

	async function fetchPersonData() {
		try {
			person = undefined;
			pageError = undefined;
			if (!data.personId) {
				return;
			}
			person = await getPerson(data.personId);
			activeCreditType = person.knownForDepartment;
			getCredits(activeCreditType);
		} catch (err) {
			person = undefined;
			pageError = err;
		}
	}

	async function getPerson(id: number) {
		return await req.get<PersonDetailsResponse>(`/content/person/${id}`);
	}

	async function getCredits(type?: string) {
		try {
			console.log("getCredits: type:", type);
			credits = undefined;
			creditsError = undefined;
			credits = await req.get<PersonCreditsResponse>(
				`/content/person/${data.personId}/credits`,
				{
					params: {
						type: type,
					},
				},
			);
			credits.credits = credits.credits?.filter(
				(v, i, a) => a.findIndex((t) => t.ids.tmdb === v.ids.tmdb) === i,
			); // remove duplicate entries. If an actor has multiple roles in a single movie, it would otherwise show up multiple times
			sortCredits(sortOption);
		} catch (err) {
			credits = undefined;
			creditsError = err;
		}
	}

	function newestOldestSort(
		a: Media,
		b: Media,
		/**
		 * 0 = Newest,
		 * 1 = Oldest
		 */
		n: 0 | 1,
	) {
		// Assume missing release date means future release (TBD)
		if (!a.releaseDate && !b.releaseDate) {
			// Both releases have no date, return as equals
			// here to avoid an infinite loop.
			return 0;
		}
		if (!a.releaseDate && !a.releaseDate) return n === 0 ? -1 : 1;
		if (!b.releaseDate && !b.releaseDate) return n === 0 ? 1 : -1;

		const dateA = new Date(a.releaseDate).valueOf();
		const dateB = new Date(b.releaseDate).valueOf();

		if (n === 0) {
			return dateB - dateA;
		} else {
			return dateA - dateB;
		}
	}

	function sortCredits(sortOption: string) {
		if (!credits || !credits.credits) return;
		switch (sortOption) {
			case "Vote count":
				credits.credits.sort(
					(a, b) => (b.ratingCount ?? 0) - (a.ratingCount ?? 0),
				);
				break;
			case "Newest":
				credits.credits.sort((a, b) => newestOldestSort(a, b, 0));
				break;
			case "Oldest":
				credits.credits.sort((a, b) => newestOldestSort(a, b, 1));
				break;
		}
		credits.credits = credits.credits;
	}
</script>

<svelte:head>
	<title>{person?.name ? `${person.name} - ` : ""}Person</title>
</svelte:head>

<div>
	{#if pageError}
		<Error pretty="Failed to load person!" error={pageError} />
	{:else if !person}
		<Spinner />
	{:else if Object.keys(person).length > 0}
		{#if Object.keys(person).length > 0}
			{#if credits?.credits && credits.credits.length > 0 && credits.credits[0].extBackdropPath}
				<PageBackdrop
					src={"https://www.themoviedb.org/t/p/w1920_and_h800_multi_faces" +
						credits.credits[0].extBackdropPath}
				/>
			{/if}
			<div class="content">
				<div class="details-wrap">
					<div class="details-container">
						{#if person.extPosterPath}
							<PosterImage
								src={"https://image.tmdb.org/t/p/w500" + person.extPosterPath}
							/>
						{/if}

						<div class="details">
							<span class="title-container">
								<a href={person.homepage} rel="external" target="_blank">
									{person.name}
								</a>
								<span></span>
							</span>

							<ExpandableText title="Biography" text={person.biography} />

							<div class="detail-info">
								{#if person.knownForDepartment}
									<div>
										<span>Department</span>
										<span>{person.knownForDepartment}</span>
									</div>
								{/if}
								{#if person.placeOfBirth}
									<div>
										<span>Born In</span>
										<span>{person.placeOfBirth}</span>
									</div>
								{/if}
								{#if person.birthday}
									<div>
										<span>Born On</span>
										<span
											>{new Date(
												Date.parse(person.birthday),
											).toLocaleDateString()}</span
										>
									</div>
								{/if}
								{#if person.deathday}
									<div>
										<span>Died On</span>
										<span>
											{new Date(
												Date.parse(person.deathday),
											).toLocaleDateString()}
										</span>
									</div>
								{/if}
								{#if person.age}
									<div>
										<span>Age</span>
										<span>{person.age} Years</span>
									</div>
								{/if}
								{#if person.knownCredits}
									<div>
										<span>Known Credits</span>
										<span>{person.knownCredits}</span>
									</div>
								{/if}
								{#if person.firstKnownCredit}
									{@const fkc = person.firstKnownCredit}
									<div
										use:tooltip={{
											text: `When they were ${fkc.age}!`,
											condition: Boolean(fkc.age),
										}}
									>
										<span>First Credit</span>
										<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
										<a href={`/${fkc.type}/${fkc.id}`}>
											{new Date(Date.parse(fkc.date)).toLocaleDateString()}
										</a>
									</div>
								{/if}
							</div>
						</div>
					</div>
				</div>
			</div>
			{#if creditsError}
				<Error
					pretty="Failed to load credits!"
					error={creditsError}
					onRetry={() => {
						getCredits(activeCreditType);
					}}
				/>
			{:else if !credits}
				<Spinner />
			{:else if credits}
				{#if credits?.credits && credits?.credits?.length > 0}
					<div class="control-bar">
						<div class="title">
							<h2>Credits</h2>
							<CountBadge count={credits.credits.length} />
						</div>
						<div class="listFilter">
							<span>On my list</span>
							<Checkbox name="On my list" bind:value={onMyListFilter} />
						</div>
						{#if person.creditTypes && person.creditTypes.length > 0}
							<DropDown
								bind:active={
									() => activeCreditType,
									(t) => {
										activeCreditType = t;
										getCredits(t);
									}
								}
								placeholder="Credit Type"
								options={person.creditTypes}
								isDropDownItem={false}
								showActiveElementsInOptions={true}
								// min-width so when a small entry like "Art" is
								// selected, the dropdown box doesn't become too
								// small for other options text and cause
								// overflowing.
								minWidth="100px"
							/>
						{/if}
						<DropDown
							bind:active={sortOption}
							placeholder="Vote count"
							options={["Vote count", "Newest", "Oldest"]}
							isDropDownItem={false}
							showActiveElementsInOptions={true}
						/>
					</div>
					<div class="page">
						<PosterList>
							{#each credits.credits as c, i (`${i}-${c.ids.tmdb}`)}
								<Poster
									media={c}
									bind:watched={credits.credits[i].watched}
									fluidSize
									hideIfNotOnList={onMyListFilter}
								/>
							{/each}
						</PosterList>
					</div>
				{:else}
					<div class="no-credits-message">
						<Icon i="star" wh={80} />
						<h2 class="norm">We found no credits!</h2>
						<h4 class="norm">It seems that this person has no credits.</h4>
					</div>
				{/if}
			{/if}
		{:else}
			person not found
		{/if}
	{:else}
		<Error error="Person not found" pretty="Person not found" />
	{/if}
</div>

<style lang="scss">
	@use "../../../../lib/content/page.scss";

	.control-bar {
		align-items: center;
		display: flex;
		flex-wrap: wrap;
		gap: 20px;
		margin: 0 auto;
		padding-left: 20px;
		padding-right: 20px;
		width: 100%;
		/* Same as in PosterList */
		max-width: 1200px;

		.title {
			display: flex;
			flex-flow: row;
			gap: 10px;
			align-items: center;
			margin-right: auto;
		}

		.listFilter {
			display: flex;
			align-items: center;
			gap: 8px;
		}
	}

	.content {
		position: relative;
		color: white;
		margin-bottom: 15px;

		.details-container .details {
			.title-container {
				a {
					color: white;
					text-decoration: none;
					font-size: 30px;
					font-weight: bold;
					padding-right: 3px;
				}

				span {
					font-size: 20px;
					color: rgba($color: #fff, $alpha: 0.7);
				}
			}

			.detail-info {
				display: flex;
				flex-wrap: wrap;
				gap: 15px 30px;
				margin-top: 10px;
				font-size: 14px;

				div {
					display: flex;
					flex-flow: column;

					span:first-child {
						font-weight: bold;
					}

					a {
						color: white;
						text-decoration: underline;
					}
				}
			}
		}
	}

	.page {
		display: flex;
		flex-flow: column;
		align-items: center;
		gap: 30px;
		padding: 10px 0px;
	}

	.no-credits-message {
		display: flex;
		flex-flow: column;
		gap: 5px;
		align-items: center;
		margin-top: 20px;

		h2 {
			margin-top: 10px;
		}

		h4 {
			font-weight: normal;
		}
	}
</style>
