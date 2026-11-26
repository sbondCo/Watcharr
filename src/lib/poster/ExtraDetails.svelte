<!-- Extra Details View For Posters -->
<script lang="ts">
	import { store } from "@/store.svelte";
	import { RatingSystem } from "@/types";
	import {
		getOrdinalSuffix,
		monthsShort,
		watchedStatuses,
	} from "../util/helpers";
	import Icon from "../Icon.svelte";
	import { toShowableRating, toWhichThumb } from "../rating/helpers";
	import { page } from "$app/state";
	import type { PosterExtraDetails } from "./lib";

	let {
		rating,
		status,
		dateAdded,
		dateModified,
		lastWatched,
	}: PosterExtraDetails = $props();

	let isUsingThumbs = $derived(
		store.userSettings &&
			store.userSettings.ratingSystem === RatingSystem.Thumbs,
	);

	function formatDate(e: number) {
		if (!e) {
			return "Unknown";
		}
		const d = new Date(e);
		const now = new Date(Date.now());
		const year =
			d.getFullYear() == now.getFullYear()
				? ""
				: `'${String(d.getFullYear()).substring(2, 4)}`;
		return `${d.getDate()}${getOrdinalSuffix(d.getDate())} ${monthsShort[d.getMonth()]} ${year}`;
	}
</script>

{#if (page.url?.pathname === "/" || page.url?.pathname.startsWith("/search")) && store.wlDetailedView && store.wlDetailedView.length > 0}
	<div class="extra-details">
		<!--
      The `if` statements can't be on their own line to look pretty
      because that will cause whitespace in the generated markup,
      which causes the :empty css tag to not work.
      Can be reverted when its possible to trim whitespace in svelte
      OR when :empty tag is updated in browsers to new spec and counts whitespace as empty.
    -->
		{#if dateAdded && store.wlDetailedView.includes("dateAdded")}
			<span title="Date added to watch list">
				<i><Icon i="calendar" /></i>
				<span>
					{formatDate(Date.parse(dateAdded))}
				</span>
			</span>
		{/if}{#if dateModified && store.wlDetailedView.includes("dateModified")}
			<span title="Date last modified">
				<i><Icon i="pencil" wh={15} /></i>
				<span>
					{formatDate(Date.parse(dateModified))}
				</span>
			</span>
		{/if}{#if lastWatched && store.wlDetailedView.includes("lastWatched")}
			<span title="Latest season watched">
				<i><Icon i="play" wh={15} /></i>
				<span>{lastWatched}</span>
			</span>
		{/if}{#if store.wlDetailedView.includes("statusRating")}
			<!-- NOTE: To avoid having to migrate localStorage, reusing
				 'statusRating' for just rating, splitting status to its own.
				 This is probably moving from localStorage to server soon so no
				 point wasting time setting up ls migrations. -->
			<span class="rating" title="Rating">
				{#if !isUsingThumbs}
					<i><Icon i="star" /></i>
				{/if}
				<span class="rating-span">
					{#if isUsingThumbs}
						{@const r = toWhichThumb(rating)}
						{#if r === -1}
							<i><Icon i="thumb-down" /></i>
						{:else if r === 0}
							<span
								style="display: flex; transform: translate(2px, 3px); font-size: 40px; font-family: 'Shrikhand';"
							>
								-
							</span>
						{:else if r === 1}
							<i><Icon i="thumb-up" /></i>
						{/if}
					{:else}
						{toShowableRating(rating)}
					{/if}
				</span>
			</span>
		{/if}{#if store.wlDetailedView.includes("status")}
			<span title="Status">
				{#if status}
					<i><Icon i={watchedStatuses[status]} wh={15} /></i>
				{/if}
			</span>
		{/if}
	</div>
{/if}

<style lang="scss">
	.extra-details {
		position: absolute;
		bottom: 5px;
		left: 50%;
		transform: translateX(-50%);
		display: flex;
		flex-flow: row;
		flex-wrap: wrap;
		gap: 5px;
		row-gap: 3px;
		font-size: 14px;
		width: 160px;
		color: white;
		transition: opacity 100ms ease-out;
		pointer-events: none !important;

		& > span {
			display: flex;
			flex-flow: row;
			align-items: center;
			gap: 6px;
			height: 15px;
			font-weight: bold;
			padding: 10px 5px;
			background-color: $poster-extra-detail-bg-color;
			border-radius: 3px;

			i {
				display: flex;
				width: 15px;
				fill: white;
			}
		}

		.rating .rating-span:empty {
			/* Can happen when content is not rated and using thumbs system */
			display: none;
		}
	}
</style>
