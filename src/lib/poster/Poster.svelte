<script lang="ts">
	import {
		type WatchedStatus,
		type Watched,
		type Media,
		MediaTypeE,
		type SupportedMedia,
	} from "@/types";
	import {
		calculateTransformOrigin,
		isTouch,
		mouseOverEl,
	} from "@/lib/util/helpers";
	import { goto } from "$app/navigation";
	import { baseURL, removeWatched, updateWatched } from "../util/api";
	import { notify } from "../util/notify";
	import { onMount } from "svelte";
	import PosterStatus from "./PosterStatus.svelte";
	import PosterRating from "./PosterRating.svelte";
	import ExtraDetails from "./ExtraDetails.svelte";
	import { buildExtraDetails } from "./lib";
	import { decode } from "blurhash";

	interface Props {
		media: Media;
		/**
		 * If this content is on our watched list,
		 * the entry should be provided in full.
		 */
		watched?: Watched;
		small?: boolean;
		disableInteraction?: boolean;
		hideButtons?: boolean;
		fluidSize?: boolean;
		/**
		 * If the poster should be hidden if not on users watched list (no `id`).
		 * Doing it this way so we can quickly hide posters with css and avoid
		 * triggering the #each block again where we create poster lists,
		 * which makes this functionality more performant (because we don't have
		 * support for virtual lists yet, we are re-creating all posters in places).
		 * Notably 'On my list' feature (eg on person page).
		 */
		hideIfNotOnList?: boolean;
		// When provided, default click handlers will instead run this callback.
		onClick?: (() => void) | undefined;
		/**
		 * Ran when watched item is updated via poster.
		 */
		onUpdated?: (() => void) | undefined;
	}

	let {
		media,
		// The `watched` prop is bindable so that when we update
		// or add content to our list, we can let the update flow
		// back to our parent that passed it in (so state is in sync
		// between parent and this component).
		watched = $bindable(undefined),
		small = false,
		disableInteraction = false,
		hideButtons = false,
		fluidSize = false,
		hideIfNotOnList = false,
		onClick = undefined,
		onUpdated = undefined,
	}: Props = $props();

	// If poster is active (scaled up)
	let posterActive = $state(false);
	// If mouse in on poster. Added to fix #656.
	let mouseOverPoster = $state(false);
	// If the image is loaded or failed. -1 = fail, 0 = false, 1 = true
	let posterImgLoaded = $state(0);

	let containerEl: HTMLDivElement | undefined = $state();
	let bhCanvas: HTMLCanvasElement | undefined = $state();

	const meta:
		| {
				id: number | undefined;
				type: SupportedMedia;
		  }
		| undefined = $derived.by(() => {
		let id: number | undefined;
		let type: SupportedMedia;
		switch (media.type) {
			case MediaTypeE.tmdbMovie:
				id = media.ids.tmdb;
				type = "movie";
				break;
			case MediaTypeE.tmdbShow:
				id = media.ids.tmdb;
				type = "tv";
				break;
			case MediaTypeE.igdbGame:
				id = media.ids.igdb;
				type = "game";
				break;
			default:
				return;
		}
		return {
			id,
			type,
		};
	});
	const poster = $derived.by(() => {
		if (media.poster?.path) {
			return `${baseURL}/${media.poster.path}`;
		}

		// Logic below uses `extPosterPath`, so check here first.
		// If it doesn't exist, return undefined.
		if (!media.extPosterPath) {
			return;
		}

		if (
			media.type == MediaTypeE.tmdbMovie ||
			media.type == MediaTypeE.tmdbShow
		) {
			if (watched) {
				// For now, if the content is on watched list, we can assume we have a local
				// cached image. Could be improved, since we could have a cached image for
				// show not on someone elses watched list.
				return `${baseURL}/img${media.extPosterPath}`;
			} else {
				return `https://image.tmdb.org/t/p/w500${media.extPosterPath}`;
			}
		} else if (media.type == MediaTypeE.igdbGame) {
			return `https://images.igdb.com/igdb/image/upload/t_cover_big/${media.extPosterPath}.jpg`;
		}
	});
	const link = $derived(meta?.id ? `/${meta.type}/${meta.id}` : undefined);
	const year = $derived(
		media.releaseDate ? new Date(media.releaseDate).getFullYear() : undefined,
	);

	function handleStarClick(r: number) {
		if (r == watched?.rating || !meta?.id) return;
		updateWatched(watched, {
			contentId: meta.id,
			contentType: meta.type,
			rating: r,
		}).then((w) => {
			if (typeof onUpdated === "function") {
				onUpdated();
				runPosterMouseLeaveIfNeeded();
			}
			// If watched was just added, we need to assign
			// it to our `watched` var to get the update.
			watched = w;
		});
	}

	function handleStatusClick(type: WatchedStatus | "DELETE") {
		if (type === "DELETE") {
			if (!watched) {
				notify({
					text: "Content has no watched list entry, can't delete.",
					type: "error",
				});
				return;
			}
			removeWatched(watched.id).then((removed) => {
				if (removed) {
					watched = undefined;
				}
			});
			return;
		}
		if (type == watched?.status || !meta?.id) return;
		updateWatched(watched, {
			contentId: meta.id,
			contentType: meta.type,
			status: type,
		}).then((w) => {
			if (typeof onUpdated === "function") {
				onUpdated();
				runPosterMouseLeaveIfNeeded();
			}
			// If watched was just added, we need to assign
			// it to our `watched` var to get the update.
			watched = w;
		});
	}

	function handleInnerKeyUp(e: KeyboardEvent) {
		if (
			e.key === "Enter" &&
			(e.target as HTMLElement)?.id === "ilikemoviessueme"
		) {
			if (typeof onClick !== "undefined") {
				onClick();
				return;
			}
			if (link) {
				goto(link);
			}
		}
	}

	onMount(() => {
		if (containerEl) {
			if (small) {
				containerEl.classList.add("small");
			}
			if (fluidSize) {
				containerEl.classList.add("fluid-size");
			}
		}
	});

	/**
	 * Manual way to check if we need to run the mouseleave
	 * event for the poster containerEl.
	 */
	function runPosterMouseLeaveIfNeeded() {
		// Timeout to give enough time for the element to
		// actually move if it needs to (which can happen if
		// certain filters/sorts are applied).
		setTimeout(() => {
			if (!mouseOverEl(containerEl)) {
				posterOnMouseLeave();
			}
		}, 100);
	}

	function posterOnMouseLeave() {
		mouseOverPoster = false;
		posterActive = false;
		const ae = document.activeElement;
		if (
			ae &&
			ae instanceof HTMLElement &&
			(ae.parentElement?.id === "ilikemoviessueme" ||
				ae.parentElement?.parentElement?.id === "ilikemoviessueme")
		) {
			// Stops the poster being re-focused after the browser window
			// loses focus, then regains it (ex: you middle click the poster,
			// go to the opened tab (or lose browser window focus, then when
			// you come back the poster is sent `focusin` and stuck activated
			// until mouseleave again).
			ae.blur();
		}
	}

	onMount(() => {
		if (containerEl) {
			if (small) {
				containerEl.classList.add("small");
			}
			if (fluidSize) {
				containerEl.classList.add("fluid-size");
			}
		}

		// Show blurhash if we can
		if (media.poster?.path && media.poster?.blurHash && bhCanvas) {
			const pixels = decode(media.poster.blurHash, 170, 256);
			const ctx = bhCanvas.getContext("2d");
			if (ctx) {
				const imageData = ctx.createImageData(170, 256);
				imageData.data.set(pixels);
				ctx.putImageData(imageData, 0, 0);
			}
		}
	});
</script>

<!-- HACK: disabled this issue for now, it should probably be fixed properly -->
<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<li
	onmouseenter={(e) => {
		mouseOverPoster = true;
		if (!posterActive) calculateTransformOrigin(e);
		if (!isTouch()) {
			posterActive = true;
		}
	}}
	onfocusin={(e) => {
		if (!posterActive) calculateTransformOrigin(e);
		if (!isTouch()) {
			posterActive = true;
		}
	}}
	onfocusout={() => {
		if (!isTouch() && !mouseOverPoster) {
			// Only on !isTouch (to match focusin) to avoid breaking a tap and hold on link on mobile.
			// and only if mouse isn't still over the poster, fixes focusout on click of rating/status
			// poster buttons causing poster to shrink until refocused with click/mouse out & in again.
			posterActive = false;
		}
	}}
	onmouseleave={posterOnMouseLeave}
	onclick={() => (posterActive = true)}
	onkeyup={(e) => {
		if (e.key === "Tab") {
			e.currentTarget.scrollIntoView({ block: "center" });
		}
	}}
	onkeypress={() => console.log("on kpress")}
	class={`${posterActive ? "active " : ""}${watched?.pinned ? "pinned " : ""}${hideIfNotOnList && !watched ? "hidden " : ""}`}
>
	<div
		class={`container${!poster || posterImgLoaded == -1 ? " details-shown" : ""}`}
		bind:this={containerEl}
	>
		{#if poster}
			{#if posterImgLoaded == 0}
				<!-- We only show the loader (or blurhash) while the image
				  is still loading. -->
				{#if media?.poster?.blurHash}
					<canvas
						width="170"
						height="256"
						bind:this={bhCanvas}
						class="img-loader"
					></canvas>
				{:else}
					<div class="img-loader"></div>
				{/if}
			{/if}
			<img
				loading="lazy"
				src={poster}
				alt=""
				onload={(e) => {
					posterImgLoaded = 1;
				}}
				onerror={(e) => {
					posterImgLoaded = -1;
				}}
			/>
		{/if}
		{#if watched && meta && !posterActive}
			<!-- Must be on watched list, and poster not hovered -->
			<ExtraDetails {...buildExtraDetails(meta.type, watched)} />
		{/if}
		<div
			onclick={(e) => {
				if (typeof onClick !== "undefined") {
					onClick();
					// Prevent the link inside this div from being clicked in this case.
					e.preventDefault();
					return;
				}
				if (posterActive && link) goto(link);
			}}
			onkeyup={handleInnerKeyUp}
			id="ilikemoviessueme"
			class="inner"
			role="button"
			tabindex="-1"
		>
			<a data-sveltekit-preload-data="tap" href={link} class="small-scrollbar">
				<h2>
					{media.name}
					{#if year}
						<time>{year}</time>
					{/if}
				</h2>
				<span>{media.summary}</span>
			</a>

			{#if !hideButtons}
				<div class="buttons">
					<PosterRating
						rating={watched?.rating}
						{handleStarClick}
						{disableInteraction}
					/>
					<PosterStatus
						status={watched?.status}
						{handleStatusClick}
						{disableInteraction}
					/>
				</div>
			{/if}
		</div>
	</div>
</li>

<style lang="scss">
	li.hidden {
		display: none;
	}

	li.active {
		cursor: pointer;
	}

	li.pinned:not(.active) .container {
		outline: 3px solid gold;
	}

	li {
		&:not(.active) {
			.container .inner,
			.container .inner .buttons {
				pointer-events: none !important;
			}
		}
	}

	.container {
		display: flex;
		flex-flow: column;
		background-color: rgb(48, 45, 45);
		overflow: hidden;
		flex: 1 1;
		border-radius: 5px;
		min-width: 170px;
		width: 170px;
		position: relative;
		aspect-ratio: 170000/256367;
		transition: transform 150ms ease;

		&.fluid-size {
			height: 100%;
			width: 100%;
		}

		img {
			width: 100%;
			height: 100%;
		}

		&.details-shown .img-loader {
			display: none;
		}

		.img-loader {
			position: absolute;
			width: 100%;
			height: 100%;
			background-color: gray;
			background: linear-gradient(359deg, #5c5c5c, #2c2929, #2c2424);
			background-size: 400% 400%;
			animation: imgloader 4s ease infinite;

			@keyframes imgloader {
				0% {
					background-position: 50% 0%;
				}
				50% {
					background-position: 50% 100%;
				}
				100% {
					background-position: 50% 0%;
				}
			}
		}

		.inner {
			position: absolute;
			opacity: 0;
			display: flex;
			flex-flow: column;
			top: 0;
			height: 100%;
			width: 100%;
			padding: 10px;
			background-color: transparent;
			transition: opacity 150ms cubic-bezier(0.19, 1, 0.22, 1);

			& > a {
				height: 100%;
				overflow: auto;
			}

			h2 {
				font-family:
					sans-serif,
					system-ui,
					-apple-system,
					BlinkMacSystemFont;
				font-size: 18px;
				color: white;
				word-wrap: break-word;

				time {
					font-size: 14px;
					font-weight: 400;
					color: rgba(255, 255, 255, 0.7);
				}
			}

			span {
				color: white;
				margin: 5px 0 5px 0;
				font-size: 9px;
				display: -webkit-box;
				-webkit-line-clamp: 5;
				-webkit-box-orient: vertical;
				hyphens: auto;
				overflow: hidden;
			}

			.buttons {
				display: flex;
				flex-flow: row;
				margin-top: auto;
				gap: 10px;
				height: 35px;
			}
		}

		&.small .inner span {
			font-size: 11px;
		}

		.active & {
			transform: scale(1.3);
			z-index: 99;
		}

		.active &.small {
			transform: scale(1.1);
		}

		.active &,
		&:global(.details-shown) {
			img {
				filter: blur(4px) grayscale(80%);
				// This makes the background very dark,
				// but atleast the text is visible.. may want to change later.
				mix-blend-mode: multiply;
			}

			.inner {
				color: white;
				opacity: 1;
			}
		}
	}
</style>
