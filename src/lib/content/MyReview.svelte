<script lang="ts">
	import type { Watched, WatchedStatus } from "@/types";
	import Rating from "../rating/Rating.svelte";
	import Status from "../Status.svelte";
	import MyThoughts from "./MyThoughts.svelte";

	interface Props {
		watched?: Watched;
		contentTitle?: string;
		onRatingChanged: (newRating: number) => Promise<boolean>;
		onStatusChanged: (newStatus: WatchedStatus) => Promise<boolean>;
		onThoughtsChanged: (newThoughts: string) => Promise<boolean>;
	}

	let {
		watched,
		contentTitle,
		onRatingChanged,
		onStatusChanged,
		onThoughtsChanged,
	}: Props = $props();
</script>

<div class="review">
	<Rating rating={watched?.rating} onChange={onRatingChanged} />
	<Status status={watched?.status} onChange={onStatusChanged} />
	{#if watched}
		<MyThoughts
			{contentTitle}
			thoughts={watched?.thoughts}
			onChange={onThoughtsChanged}
		/>
		{#if typeof watched.plays == "number" && watched.plays > 0}
			<div>
				{watched.plays}
				{watched.plays > 1 ? "Plays" : "Play"}
			</div>
		{/if}
	{/if}
</div>

<style lang="scss">
	.review {
		display: flex;
		flex-flow: column;
		gap: 10px;
		width: 100%;
		max-width: 380px;
		color: $text-color;
		margin-left: auto;
		margin-right: auto;
		margin-top: 22px;

		@media screen and (max-width: 420px) {
			max-width: 340px;
		}
	}
</style>
