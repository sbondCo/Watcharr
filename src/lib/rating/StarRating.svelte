<!-- To be used from Rating component -->

<script lang="ts">
	import { run } from "svelte/legacy";

	import { store } from "@/store.svelte";
	import { RatingStep, RatingSystem } from "@/types";
	import { onMount } from "svelte";
	import { isTouch } from "../util/helpers";

	interface Props {
		rating: number | undefined;
		onChange: (newRating: number) => Promise<boolean>;
	}

	let { rating, onChange }: Props = $props();

	let hoveredRating: number | undefined = $state();
	let shownRating: number | undefined = $state();
	let shownPerc: number | undefined = $state();
	let ratingContainer: HTMLDivElement | undefined = $state();
	let ratingWrapEl: HTMLDivElement | undefined = $state();
	let ratingText: HTMLSpanElement | undefined = $state();
	let highlightContainer: HTMLDivElement | undefined = $state();
	let normalContainer: HTMLDivElement | undefined = $state();
	/**
	 * Percentage star step.
	 */
	let starStep = $state(10);
	/**
	 * The scroll position at start of star pointer click,
	 * saved for comparison at end of star click. If same,
	 * new rating will be saved, otherwise ignored to fix
	 * issue where user is just trying to scroll down, but
	 * presses star accidentally.
	 */
	let scrollLocAtStart = 0;

	const ratingDesc = [
		"Apalling",
		"Horrible",
		"Very Bad",
		"Bad",
		"Average",
		"Fine",
		"Good",
		"Very Good",
		"Great",
		"Masterpiece",
	];

	async function saveSelectedRating() {
		if (!shownPerc) {
			console.warn("saveSelectedRating: Rating not set, ignoring call.");
			return;
		}
		// We attempt to avoid accidental clicks on stars being saved while trying
		// to scroll down the page, so if the scroll pos has changed since the press
		// started, ignore it. This issue only affects touch devices, so ignore on pc.
		if (isTouch() && scrollLocAtStart !== window.scrollY) {
			console.info(
				"saveSelectedRating: Scroll location has changed since star was pressed, not saving rating to avoid a misclick.",
			);
			return;
		}
		// Rating needs to always scale to out of 10 before saving,
		// scaling down the shown percentage will work.
		const r = Math.round(shownPerc) / 10;
		console.log("saveSelectedRating:", shownPerc, "=", r);
		if (r === rating) {
			console.warn("saveSelectedRating: Rating not changed, ignoring call.");
			return;
		}
		return await onChange(r);
	}

	function setHoveredRatingFromPerc(perc: number) {
		let hovR = perc;
		if (store.userSettings?.ratingSystem === RatingSystem.OutOf5) {
			hovR = hovR / 20;
		} else {
			hovR = hovR / 10;
		}
		hoveredRating = Math.max(Math.min(hovR, stars[0]), 0);
	}

	/**
	 * Show RatingText over hovered star.
	 */
	function moveRatingText() {
		try {
			if (!hoveredRating) {
				handleRatingHoverEnd();
				return;
			}
			if (!ratingContainer || !ratingWrapEl) {
				console.error("moveRatingText: Required elements not found!");
				handleRatingHoverEnd();
				return;
			}
			// Get star number we are putting text above
			let r: number;
			if (store.userSettings?.ratingSystem === RatingSystem.OutOf5) {
				r = Math.ceil(hoveredRating);
			} else {
				r = Math.ceil(hoveredRating);
			}
			const start = ratingContainer?.getBoundingClientRect()?.x;
			const starl = ratingWrapEl?.getBoundingClientRect()?.left;
			const bodyRect = document.body.getBoundingClientRect();
			const oneStarWidth = bodyRect.width <= 420 ? 33.5 : 37.5;
			const offset = (r - 1) * oneStarWidth;
			let prospectLeft = starl + offset - start + 11.5;
			const bodyXX = bodyRect.width - 120;
			if (bodyXX < prospectLeft) {
				// Move text back if going off right
				prospectLeft = prospectLeft - (prospectLeft - bodyXX);
			} else if (bodyRect.x + 20 > prospectLeft && starl < 60) {
				// This should stop the first star text from going off left
				prospectLeft = prospectLeft + 30;
			}
			if (ratingText) ratingText.style.left = `${prospectLeft}px`;
		} catch (err) {
			console.error("moveRatingText: Failed!", err);
		}
	}

	function handleRatingHoverEnd() {
		console.debug("handleRatingHoverEnd");
		hoveredRating = undefined;
		if (ratingText) {
			ratingText.style.left = "50%";
		}
	}

	function handleMouseOver(
		ev: (TouchEvent | MouseEvent) & {
			currentTarget: EventTarget & HTMLDivElement;
		},
		start = false,
	) {
		const rect = ev.currentTarget.getBoundingClientRect();
		const x =
			(ev instanceof MouseEvent ? ev.clientX : ev.touches[0].clientX) -
			rect.left; // rel to start of container
		const perc =
			Math.ceil(Math.round((x * 100) / rect.width) / starStep) * starStep;
		setHoveredRatingFromPerc(perc);
		moveRatingText();
		if (start) {
			scrollLocAtStart = window.scrollY;
		}
	}

	function handleKeyDown(
		ev: KeyboardEvent & {
			currentTarget: EventTarget & HTMLDivElement;
		},
	) {
		console.log("handleKeyDown:", ev);
		if (ev.code === "ArrowRight") {
			console.debug("handleKeyDown: Increasing selected rating.");
			setHoveredRatingFromPerc((shownPerc ?? 0) + starStep);
		} else if (ev.code === "ArrowLeft") {
			console.debug("handleKeyDown: Decreasing selected rating.");
			setHoveredRatingFromPerc(shownPerc ? shownPerc - starStep : 0);
		} else if (ev.key === "Enter") {
			console.debug("handleKeyDown: Enter pressed.. saving rating.");
			saveSelectedRating();
		}
	}

	function showRating(perc: number) {
		try {
			if (!highlightContainer || !normalContainer) {
				console.warn("showRating: Containers not defined yet.");
				return;
			}
			// console.debug("showRating: perc", perc, starStep);
			perc = Math.max(Math.min(Math.round(perc / starStep) * starStep, 100), 0);
			// console.debug("showRating: perc2", perc);
			if (perc > 1) {
				let percToHighlight = perc;
				let percToHide = 100 - perc;
				if (starStep == 5) {
					// On step of 5, it looks nicer when we take highlight back one percent,
					// otherwise more than half the star looks highlighted... ew. This is only visual.
					percToHighlight--;
					percToHide++;
				}
				highlightContainer.style.display = "flex";
				highlightContainer.style.width = `${percToHighlight}%`;
				// We shrink this container too because of what seems to be a bug
				// where the text-stroke draws on the upper layer too, instead
				// of being hidden by the highlighted stars overlay.
				normalContainer.style.width = `${percToHide}%`;
			} else {
				highlightContainer.style.display = "none";
				highlightContainer.style.width = "0";
				normalContainer.style.width = "100%";
			}
			shownPerc = perc;
		} catch (err) {
			console.error("showRating: Failed!", err);
		}
	}

	onMount(() => {
		if (rating) showRating(Math.round((rating * 100) / 10));
	});
	let stars = $derived(
		store.userSettings?.ratingSystem == RatingSystem.OutOf5
			? [5, 4, 3, 2, 1]
			: [10, 9, 8, 7, 6, 5, 4, 3, 2, 1],
	);
	run(() => {
		if (hoveredRating !== undefined && hoveredRating > 0) {
			console.debug("showRatingCaller: We have a hoveredRating.");
			shownRating = hoveredRating;
			showRating(
				Math.round(
					(hoveredRating * 100) /
						(store.userSettings?.ratingSystem === RatingSystem.OutOf5 ? 5 : 10),
				),
			);
		} else if (rating !== undefined) {
			console.debug("showRatingCaller: We have a rating.");
			shownRating = rating;
			showRating(Math.round((rating * 100) / 10));
		} else {
			console.debug("showRatingCaller: We have nothing.");
			shownRating = undefined;
			showRating(0);
		}
	});
	run(() => {
		// console.log("block", starStep, settings?.ratingStep, settings.ratingSystem);
		try {
			if (store.userSettings) {
				if (typeof store.userSettings.ratingSystem === "number") {
					// Set default star step for system.
					switch (store.userSettings.ratingSystem) {
						case RatingSystem.OutOf100:
							starStep = 1;
							break;

						case RatingSystem.OutOf5:
							starStep = 20;
							break;

						case RatingSystem.OutOf10:
						default:
							starStep = 10;
							break;
					}
				}
				// Override default with user set step if supported by this system.
				if (typeof store.userSettings.ratingStep === "number") {
					if (
						store.userSettings.ratingSystem === RatingSystem.OutOf5 ||
						store.userSettings.ratingSystem === RatingSystem.OutOf10 ||
						!store.userSettings.ratingSystem
					) {
						if (
							store.userSettings.ratingStep === RatingStep.Point1 ||
							store.userSettings.ratingStep === RatingStep.Point5 ||
							store.userSettings.ratingStep === RatingStep.One
						) {
							// Turn enum value into an actual step value.
							const actualRatingStep =
								store.userSettings.ratingStep === RatingStep.Point1
									? 0.1
									: store.userSettings.ratingStep === RatingStep.Point5
										? 0.5
										: 1;
							console.log("actualRatingStep", actualRatingStep);
							starStep =
								store.userSettings.ratingSystem === RatingSystem.OutOf5
									? actualRatingStep * 20
									: actualRatingStep * 10;
							console.debug("Set starStep from setting:", starStep);
						} else {
							starStep =
								store.userSettings.ratingSystem === RatingSystem.OutOf5
									? 20
									: 10;
							console.debug("Set starStep using default:", starStep);
						}
					}
				}
			}
		} catch (err) {
			console.error("Failed to set startStep from settings:", err);
		}
	});
</script>

<!-- step: {starStep}
stepSetting: {settings?.ratingStep}<br />
hoveredRating: {hoveredRating}<br />
shownPerc: {shownPerc}<br /> -->

<div class="rating-container" bind:this={ratingContainer}>
	<span bind:this={ratingText}>
		{#if hoveredRating}
			{#if store.userSettings?.ratingSystem === RatingSystem.OutOf5 && shownPerc}
				{ratingDesc[Math.ceil(shownPerc / 10) - 1]}
			{:else}
				{ratingDesc[Math.ceil(hoveredRating) - 1]}
			{/if}
			{#if store.userSettings?.ratingSystem === RatingSystem.OutOf100}
				({shownPerc})
			{:else}
				({hoveredRating})
			{/if}
		{:else if typeof rating === "number" && rating > 0}
			{ratingDesc[Math.ceil(rating) - 1]}
			{#if shownPerc}
				{#if store.userSettings?.ratingSystem === RatingSystem.OutOf100}
					({shownPerc})
				{:else if store.userSettings?.ratingSystem === RatingSystem.OutOf5}
					({shownPerc / 20})
				{:else}
					({shownPerc / 10})
				{/if}
			{/if}
		{:else}
			Select Your Rating
		{/if}
	</span>
	<div
		class="rating-wrap"
		bind:this={ratingWrapEl}
		onpointermove={(ev) => handleMouseOver(ev)}
		ontouchmove={(ev) => handleMouseOver(ev)}
		ontouchstart={(ev) => {
			// Prevent `click` event from triggering.
			// Mainly for firefox mobile, these events dont work the same
			// there (as chromium mobile), so forcing touch devices to use
			// this flow (touchstart -> touchend) for clicks too will
			// workaround that difference.
			ev.preventDefault();
			handleMouseOver(ev, true);
		}}
		onmouseleave={(ev) => {
			if (!ev.relatedTarget) {
				// When not focused on the browser, and then clicking a star directly
				// without first focusing the browser, this event can be triggered,
				// which causes hoveredRating to reset. This check seems to fix that.
				console.debug(
					"rating-wrap: mouseleave event triggered, but me think mistake.. ignoring.",
				);
				return;
			}
			console.debug("rating-wrap: mouseleave");
			handleRatingHoverEnd();
		}}
		ontouchend={async () => {
			console.debug("rating-wrap: touchend");
			await saveSelectedRating();
			handleRatingHoverEnd();
		}}
		onblur={() => {
			console.debug("rating-wrap: blur");
			handleRatingHoverEnd();
		}}
		onclick={() => {
			saveSelectedRating();
		}}
		onkeydown={(ev) => handleKeyDown(ev)}
		role="button"
		tabindex="0"
	>
		<!-- The unlit stars. -->
		<div
			bind:this={normalContainer}
			class="rating the-normal-one"
			tabindex="-1"
		>
			{#each stars as _}
				<button class="plain" tabindex="-1">*</button>
			{/each}
		</div>
		<!-- Overlays on stars above to show them as highlighted. -->
		<div
			bind:this={highlightContainer}
			class="rating the-highlight-one"
			tabindex="-1"
		>
			{#each stars as _}
				<button class="plain lit" tabindex="-1">*</button>
			{/each}
		</div>
		<!-- Hidden stars, just to keep correct layout since the two above are abolute. -->
		<div class="rating the-hidden-one-for-layout-reasons" tabindex="-1">
			{#each stars as _}
				<button
					class="plain"
					style="opacity: 0; pointer-events: none;"
					tabindex="-1">*</button
				>
			{/each}
		</div>
	</div>
	<span class="keyboard-tip"
		>Left/Right Arrows to change rating, Enter to save.</span
	>
</div>

<style lang="scss">
	.rating-container {
		display: flex;
		flex-flow: column;
		overflow: visible;

		& > span {
			position: relative;
			transition:
				left 100ms ease-in,
				transform 100ms ease-in;
			max-width: max-content;
			left: 50%;
			transform: translateX(-50%);
		}
	}

	.rating-wrap {
		position: relative;
		user-select: none;
		cursor: pointer;
		width: max-content;
		margin-left: auto;
		margin-right: auto;
		/* For chromium on mobile, looks a lil gross */
		-webkit-tap-highlight-color: transparent;

		&:focus-visible {
			+ .keyboard-tip {
				display: unset;
			}
		}
	}

	.keyboard-tip {
		// For when the rating-wrap is accessed via keyboard.
		font-size: 12px;
		margin-top: 5px;
		margin-bottom: 5px;
		display: none;
	}

	.rating {
		display: flex;
		align-items: center;
		color: $text-color;
		overflow: hidden;
		margin: 10px 0 10px 0;
		padding: 1px;

		&.the-highlight-one {
			width: 0%;
			display: none;
			overflow: hidden;
			-webkit-text-stroke: 1.5px gold;
			pointer-events: none;
		}

		&:not(.the-highlight-one) {
			flex-flow: row-reverse;
			justify-content: center;
		}

		&:not(.the-hidden-one-for-layout-reasons) {
			position: absolute;
			left: 0;
			top: 0;
		}

		&.the-normal-one {
			justify-content: unset;
			white-space: nowrap;
			left: unset;
			right: 0;
			-webkit-text-stroke: 1.5px $text-color;
			pointer-events: none;
		}

		button {
			font-size: 55px;
			font-family: "Rampart One";
			letter-spacing: 10px;
			line-height: 52px;
			height: 38px;

			&:global(.lit),
			&:global(.lit ~ button) {
				color: gold;
				-webkit-text-stroke: 1.5px gold;
			}

			@media screen and (max-width: 420px) {
				letter-spacing: 6px;
			}
		}
	}
</style>
