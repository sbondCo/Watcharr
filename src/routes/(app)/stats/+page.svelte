<script lang="ts">
	import Error from "@/lib/Error.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import Stat from "@/lib/stats/Stat.svelte";
	import Stats from "@/lib/stats/Stats.svelte";
	import BarChart from "@/lib/stats/charts/BarChart.svelte";
	import DoughnutChart from "@/lib/stats/charts/DoughnutChart.svelte";
	import LineChart from "@/lib/stats/charts/LineChart.svelte";
	import { getStats } from "@/lib/util/api";
	import type { StatsResponse } from "@/types";

	let activeTab: "movie" | "tv" = $state("movie");
	let statsPromise: Promise<StatsResponse> = $state(getStats("movie"));

	function switchTab(tab: "movie" | "tv") {
		if (tab === activeTab) return;
		activeTab = tab;
		statsPromise = getStats(tab);
	}

	function formatRuntime(minutes: number): string {
		const hours = Math.floor(minutes / 60);
		const mins = minutes % 60;
		if (hours === 0) return `${mins}m`;
		if (mins === 0) return `${hours}h`;
		return `${hours}h ${mins}m`;
	}

	const statusColors: Record<string, string> = {
		FINISHED: "rgba(123, 213, 85, 0.8)",
		WATCHING: "rgba(61, 100, 242, 0.8)",
		PLANNED: "rgba(247, 154, 99, 0.8)",
		HOLD: "rgba(232, 186, 63, 0.8)",
		DROPPED: "rgba(232, 93, 117, 0.8)",
	};

	function getStatusColors(statuses: string[]): string[] {
		return statuses.map((s) => statusColors[s] ?? "rgba(128,128,128,0.8)");
	}

	function formatStatusLabel(s: string): string {
		return s.charAt(0) + s.slice(1).toLowerCase();
	}
</script>

<svelte:head>
	<title>Stats</title>
</svelte:head>

<div class="content">
	<div class="inner">
		<h2>Stats</h2>

		<div class="tabs">
			<button
				class={`tab ${activeTab === "movie" ? "active" : ""}`}
				onclick={() => switchTab("movie")}
			>
				Movies
			</button>
			<button
				class={`tab ${activeTab === "tv" ? "active" : ""}`}
				onclick={() => switchTab("tv")}
			>
				TV Shows
			</button>
		</div>

		{#await statsPromise}
			<Spinner />
		{:then stats}
			{#if stats.summary.totalCount === 0}
				<p class="empty">
					No {activeTab === "movie" ? "movies" : "TV shows"} tracked yet.
				</p>
			{:else}
				<Stats>
					<Stat name="Total" value={stats.summary.totalCount} large />
					{#if activeTab === "tv"}
						<Stat name="Episodes" value={stats.summary.episodesWatched} large />
					{/if}
					<Stat name="Time" value={formatRuntime(stats.summary.totalRuntime)} />
					<Stat name="Days" value={stats.summary.daysWatched} />
					<Stat name="Mean Score" value={stats.summary.meanScore || "—"} />
				</Stats>

				<section>
					<h3>Rating Distribution</h3>
					<BarChart
						labels={stats.ratingDistribution.map((b) => b.label)}
						values={stats.ratingDistribution.map((b) => b.count)}
					/>
				</section>

				{#if stats.statusDistribution.length > 0}
					<section>
						<h3>Status Distribution</h3>
						<DoughnutChart
							labels={stats.statusDistribution.map((b) =>
								formatStatusLabel(b.status),
							)}
							values={stats.statusDistribution.map((b) => b.count)}
							colors={getStatusColors(
								stats.statusDistribution.map((b) => b.status),
							)}
						/>
					</section>
				{/if}

				{#if activeTab === "tv" && stats.episodeCountDistribution && stats.episodeCountDistribution.length > 0}
					<section>
						<h3>Episode Count</h3>
						<BarChart
							labels={stats.episodeCountDistribution.map((b) => b.label)}
							values={stats.episodeCountDistribution.map((b) => b.count)}
						/>
					</section>
				{/if}

				{#if stats.releaseYear.length > 0}
					<section>
						<h3>Release Year</h3>
						<LineChart
							labels={stats.releaseYear.map((b) => String(b.year))}
							values={stats.releaseYear.map((b) => b.count)}
						/>
					</section>
				{/if}

				{#if stats.watchYear.length > 0}
					<section>
						<h3>Watch Year</h3>
						<LineChart
							labels={stats.watchYear.map((b) => String(b.year))}
							values={stats.watchYear.map((b) => b.count)}
						/>
					</section>
				{/if}
			{/if}
		{:catch err}
			<Error error={err} pretty="Failed to load stats!" />
		{/await}
	</div>
</div>

<style lang="scss">
	.content {
		padding: 20px;
	}

	.inner {
		max-width: 800px;
		margin: 0 auto;
	}

	h2 {
		margin: 0 0 16px;
	}

	.tabs {
		display: flex;
		gap: 8px;
		margin-bottom: 24px;
	}

	.tab {
		padding: 8px 20px;
		border: none;
		border-radius: 8px;
		background-color: $accent-color;
		color: $text-color;
		cursor: pointer;
		font-size: 14px;
		font-weight: 500;
		transition: background-color 0.15s;

		&:hover {
			opacity: 0.8;
		}

		&.active {
			background-color: $accent-color-hover;
			color: $bg-color;
		}
	}

	section {
		margin-top: 32px;

		h3 {
			margin: 0 0 12px;
			font-size: 16px;
			font-weight: 600;
		}
	}

	.empty {
		margin-top: 48px;
		text-align: center;
		color: $text-color-accent;
	}
</style>
