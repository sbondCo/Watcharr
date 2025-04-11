<script lang="ts">
	import { store } from "@/store.svelte";
	import { RatingSystem } from "@/types";
	import StarRating from "./StarRating.svelte";
	import ThumbRating from "./ThumbRating.svelte";

	interface Props {
		rating: number | undefined;
		onChange: (newRating: number) => Promise<void>;
	}

	let { rating, onChange }: Props = $props();
</script>

<div class="wrap">
	{#if store.userSettings?.ratingSystem === RatingSystem.Thumbs}
		<ThumbRating {rating} {onChange} />
	{:else}
		<!-- All other systems work with the stars -->
		<StarRating {rating} {onChange} />
	{/if}
</div>

<style lang="scss">
	.wrap {
		width: 377px;

		@media screen and (max-width: 420px) {
			width: 100%;
		}
	}
</style>
