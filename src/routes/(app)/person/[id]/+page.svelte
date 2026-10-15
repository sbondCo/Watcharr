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

	let { data } = $props();

	let person: PersonDetailsResponse | undefined = $state();
	let pageError: unknown | undefined = $state();
	let sortOption = $state("Vote count");
	let credits: PersonCreditsResponse | undefined = $state();
	let onMyListFilter = $state(false);
	// Credits per department, "Acting" plus the crew's (a director's films
	// are under "Directing"). `credits.credits` holds the one shown.
	let departments: Record<string, Media[]> = {};
	let departmentNames: string[] = $state([]);
	let department = $state("");
	const ALL = "All";
	// title -> department -> the person's jobs there (see updatePersonCredits).
	let jobs: Record<string, Record<string, string[]>> = {};
	let mixedJobs: string[] = [];

	const creditOrder = ["Directing", "Writing", "Production", "Acting"];
	function creditRank(d: string) {
		const i = creditOrder.indexOf(d);
		return i < 0 ? creditOrder.length : i;
	}

	function titleKey(m: Media) {
		return `${m.type}-${m.ids.tmdb}`;
	}

	// Under "All", every job on the title. In a department, its own jobs only
	// when it mixes several ("Producer" / "Executive Producer": "Director" on
	// every Directing poster would say nothing), then the person's other jobs
	// on the title ("Writer · Producer" on a film they directed).
	function badgesFor(m: Media): string[] {
		const held = jobs[titleKey(m)] ?? {};
		const out: string[] = [];
		const add = (l?: string[]) => {
			for (const j of l ?? []) if (!out.includes(j)) out.push(j);
		};
		if (department !== ALL && mixedJobs.includes(department))
			add(held[department]);
		// Credits order ("Director · Writer · Producer"), not the menu's,
		// which is by size.
		const others = departmentNames
			.filter((d) => d !== ALL && (department === ALL || d !== department))
			.sort((a, b) => creditRank(a) - creditRank(b));
		for (const d of others) add(held[d]);
		return out;
	}

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

	function showDepartment(d: string) {
		department = d;
		credits = { credits: departments[d] ?? [] };
		sortCredits(sortOption);
	}

	async function fetchPersonData() {
		try {
			person = undefined;
			pageError = undefined;
			if (!data.personId) {
				return;
			}
			person = await getPerson(data.personId);
			await updatePersonCredits();
			sortCredits(sortOption);
		} catch (err) {
			person = undefined;
			pageError = err;
		}
	}

	async function getPerson(id: number) {
		return await req.get<PersonDetailsResponse>(`/content/person/${id}`);
	}

	async function updatePersonCredits() {
		const r = await req.get<PersonCreditsResponse>(
			`/content/person/${data.personId}/credits`,
		);
		// One object per title across every list: marking a film watched under
		// "Directing" must show under "Writing" and "All" too. Also removes
		// duplicate entries (an actor with several roles in one film, a
		// director with several jobs on it).
		const byTitle: Record<string, Media> = {};
		// The shared objects can't hold per-department jobs, so they're kept
		// aside: title -> department -> jobs.
		jobs = {};
		const titles = (l: Media[] | undefined, dept: string) => {
			const out: Media[] = [];
			for (const m of l ?? []) {
				const key = titleKey(m);
				const held = ((jobs[key] ??= {})[dept] ??= []);
				for (const j of m.jobs ?? []) if (!held.includes(j)) held.push(j);
				if (byTitle[key] && out.includes(byTitle[key])) continue;
				byTitle[key] ??= m;
				out.push(byTitle[key]);
			}
			return out;
		};
		departments = {};
		const acting = titles(r.credits, "Acting");
		if (acting.length) departments["Acting"] = acting;
		for (const [d, l] of Object.entries(r.crew ?? {})) {
			const list = titles(l, d);
			if (list.length) departments[d] = list;
		}
		// Only label a department's posters when its titles differ in job
		// ("Producer" / "Executive Producer"): "Actor" on every acting credit
		// would say nothing.
		mixedJobs = Object.keys(departments).filter(
			(d) =>
				new Set(departments[d].flatMap((m) => jobs[titleKey(m)]?.[d] ?? []))
					.size > 1,
		);
		// The department TMDB says they're known for is shown first (Christopher
		// Nolan: Directing), then the others by size. Size alone would lead with
		// "Crew" for some directors: it holds their thanks credits.
		const known = person?.knownForDepartment;
		const main = known && departments[known] ? known : undefined;
		const names = Object.keys(departments)
			.filter((d) => d !== main)
			.sort((a, b) => departments[b].length - departments[a].length);
		if (main) names.unshift(main);
		const first = names[0] ?? "Acting";
		// "All": every title once, whatever the role. Only worth offering when
		// there is more than one department.
		if (names.length > 1) {
			departments[ALL] = Object.values(byTitle);
			names.unshift(ALL);
		}
		departmentNames = names;
		showDepartment(first);
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
							</div>
						</div>
					</div>
				</div>
			</div>
			{#if credits}
				{#if credits?.credits && credits?.credits?.length > 0}
					<div class="filters">
						<div class="listFilter">
							<span>On my list</span>
							<Checkbox name="On my list" bind:value={onMyListFilter} />
						</div>
						{#if departmentNames.length > 1}
							<DropDown
								bind:active={department}
								placeholder="Department"
								options={departmentNames}
								isDropDownItem={false}
								showActiveElementsInOptions={true}
								onChange={() => showDepartment(department)}
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
									badges={badgesFor(c)}
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
			{:else}
				<Spinner />
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

	.filters {
		align-items: center;
		display: flex;
		justify-content: flex-end;
		gap: 30px;
		margin: 0 auto;
		padding-left: 20px;
		padding-right: 20px;
		width: 100%;
		/* Same as in PosterList */
		max-width: 1200px;

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
