<script lang="ts">
	import { store } from "@/store.svelte";
	import tooltip from "../actions/tooltip";
	import { RatingSystem } from "@/types";
	import Icon from "../Icon.svelte";
	import { toShowableRating, toWhichThumb } from "../rating/helpers";

	interface Props {
		rating?: number | undefined;
		handleStarClick: (rating: number) => void;
		disableInteraction?: boolean;
		/**
		 * When not minimal, we will use user settings to
		 * display ratings as they want.
		 */
		minimal?: boolean;
		direction?: "top" | "bot";
		btnTooltip?: string;
		hideStarWhenRated?: boolean;
	}

	let {
		rating = undefined,
		handleStarClick,
		disableInteraction = false,
		minimal = false,
		direction = "top",
		btnTooltip = "",
		hideStarWhenRated = false,
	}: Props = $props();

	let ratingsShown = $state(false);

	// let settings = $derived($userSettings);
	let isUsingThumbs = $derived(
		store.userSettings &&
			store.userSettings.ratingSystem === RatingSystem.Thumbs,
	);
</script>

<button
	class={[
		"rating",
		minimal ? (!rating ? "minimal" : "minimal-space") : "",
		disableInteraction ? "interaction-disabled" : "",
		minimal ? "is-minimal" : "",
	].join(" ")}
	onclick={(ev) => {
		ev.stopPropagation();
		ratingsShown = !ratingsShown;
	}}
	onmouseleave={() => {
		ratingsShown = false;
	}}
	use:tooltip={{
		text: btnTooltip,
		pos: "top",
		condition: !!btnTooltip && !ratingsShown,
	}}
>
	{#if !isUsingThumbs || (isUsingThumbs && minimal && !rating)}
		<span
			class="star"
			style={hideStarWhenRated && rating ? "display: none" : ""}>*</span
		>
	{/if}
	<span
		class={[
			!rating && disableInteraction ? "unrated-text" : "",
			"rating-text",
		].join(" ")}
	>
		{#if rating}
			{#if isUsingThumbs}
				{@const r = toWhichThumb(rating)}
				{#if r === -1}
					<Icon i="thumb-down" />
				{:else if r === 0}
					<span
						style="display: flex; transform: translate(2px, -7px); font-size: 40px; font-family: 'Shrikhand';"
					>
						-
					</span>
				{:else if r === 1}
					<Icon i="thumb-up" />
				{/if}
			{:else}
				{toShowableRating(rating)}
			{/if}
		{:else if minimal}
			<!-- eslint-disable-next-line svelte/no-useless-mustaches -->
			{""}
		{:else if disableInteraction}
			Unrated
		{:else}
			Rate
		{/if}
	</span>

	<div
		class={[
			ratingsShown ? "shown" : "",
			"small-scrollbar",
			direction,
			isUsingThumbs ? "is-using-thumbs" : "",
			minimal ? "is-minimal" : "",
		].join(" ")}
	>
		{#if isUsingThumbs}
			<!-- svelte-ignore node_invalid_placement_ssr -->
			<button
				onclick={() => handleStarClick(1)}
				class="plain{rating && rating > 0 && rating < 5 ? ' active' : ''}"
				style="display: flex; justify-content: center;"
			>
				<i style="display: flex; width: 35px;"><Icon i="thumb-down" /></i>
			</button>
			<!-- svelte-ignore node_invalid_placement_ssr -->
			<button
				onclick={() => handleStarClick(5)}
				class="plain{rating && rating > 4 && rating < 9 ? ' active' : ''}"
				style="display: flex; justify-content: center;"
			>
				<span
					style="display: flex; transform: translate(0px, -2px); font-size: 40px; height: 40px; font-family: 'Shrikhand';"
				>
					-
				</span>
			</button>
			<!-- svelte-ignore node_invalid_placement_ssr -->
			<button
				onclick={() => handleStarClick(9)}
				class="plain{rating && rating > 8 ? ' active' : ''}"
				style="display: flex; justify-content: center;"
			>
				<i style="display: flex; width: 35px;"><Icon i="thumb-up" /></i>
			</button>
		{:else}
			{@const stars =
				store.userSettings?.ratingSystem == RatingSystem.OutOf5
					? [5, 4, 3, 2, 1]
					: [10, 9, 8, 7, 6, 5, 4, 3, 2, 1]}
			{#each stars as v (v)}
				<!-- svelte-ignore node_invalid_placement_ssr -->
				<button
					class="plain{rating === v ? ' active' : ''}"
					onclick={(ev) => {
						ev.stopPropagation();
						handleStarClick(
							store.userSettings?.ratingSystem === RatingSystem.OutOf5
								? v * 2
								: v,
						);
						ratingsShown = false;
					}}
				>
					{#if store.userSettings?.ratingSystem === RatingSystem.OutOf100}
						{v * 10}
					{:else if store.userSettings?.ratingSystem === RatingSystem.OutOf5}
						{v}
					{:else}
						{v}
					{/if}
				</button>
			{/each}
		{/if}
	</div>
</button>

<style lang="scss">
	button.rating {
		padding: 3px;
		position: relative;
		font-family: "Rampart One";
		width: 100%;
		height: 100%;

		&.is-minimal {
			padding: 3px 8px;
		}

		&.interaction-disabled {
			pointer-events: none;
			cursor: default;
			background-color: transparent;
			border: unset;
			fill: white;
			color: white;

			span {
				color: white !important;
			}

			.unrated-text {
				display: flex;
				align-items: center;
				font-size: 15px !important;
			}
		}

		&.minimal span:first-child {
			letter-spacing: unset;
		}

		&.minimal-space span:first-child {
			letter-spacing: 5px;
		}

		span {
			&.star {
				color: $text-color;
				font-size: 39px;
				letter-spacing: 8px;
				line-height: 52px;
				height: 42px;
			}

			&.rating-text {
				color: $text-color;
				font-size: 22px;
				height: 35px; // quick fix to make the rating num look centered - text-stroke makes it look not centered

				& :global(svg) {
					height: 100%;
					padding: 5px;
				}
			}
		}

		&:hover span,
		&:focus-visible span {
			color: $poster-rating-color;
			fill: $poster-rating-color;
		}

		div {
			display: flex;
			flex-flow: column;
			position: absolute;
			width: 100%;
			height: 200px;
			background-color: $bg-color;
			top: calc(-100% - 170px);
			list-style: none;
			border-radius: 4px 4px 0 0;
			overflow: auto;
			scrollbar-width: thin;
			z-index: 40;
			box-shadow: 0px 0px 1px #000;

			&:not(.shown) {
				display: none;
			}

			&.bot {
				top: calc(100% + 2px);
				border-radius: 0 0 4px 4px;
			}

			button.plain {
				width: 100%;
				color: $text-color;
				fill: $text-color;
				-webkit-text-stroke: 0.5px $text-color;
				font-size: 20px;
				font-family: "Rampart One";

				& :global(svg) {
					width: 100%;
					padding: 0 4.5px;
				}

				&:hover,
				&:focus-visible {
					background-color: rgb(100, 100, 100, 0.25);
				}
			}

			&.is-using-thumbs {
				height: 150px;

				&:not(.is-minimal) {
					top: calc(-100% - 120px);
				}

				button.plain {
					min-height: 40px;
					overflow: hidden;

					span {
						/* Overriding color so dash for thumbs ratings stays text-color */
						color: $text-color;
						font-family: "Rampart One";
					}
				}
			}
		}
	}
</style>
