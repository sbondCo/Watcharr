<script lang="ts">
	import { afterNavigate, goto } from "$app/navigation";
	import { resolve } from "$app/paths";
	import { page } from "$app/state";
	import Error from "@/lib/Error.svelte";
	import Icon from "@/lib/Icon.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import { req } from "@/lib/util/api";
	import { isTouch } from "@/lib/util/helpers";
	import { store } from "@/store.svelte";
	import type {
		ServerFeatures,
		Follow,
		PrivateUser,
		Tag,
		UserSettings,
	} from "@/types";
	import { onMount } from "svelte";
	import { SvelteURLSearchParams } from "svelte/reactivity";
	import HandbookModal from "@/lib/handbook/HandbookModal.svelte";
	import Buttons from "@/lib/nav/Buttons.svelte";

	interface Props {
		children?: import("svelte").Snippet;
	}

	let { children }: Props = $props();

	let navEl: HTMLElement | undefined = $state();
	let mainSearchEl: HTMLInputElement | undefined = $state();
	let searchTimeout: number;
	let scroll = window.scrollY;

	/**
	 * Handle KeyDown event on the search bar.
	 */
	function searchBarKeyDown(ev: KeyboardEvent) {
		if (
			ev.key === "ContextMenu" ||
			ev.key === "Home" ||
			ev.key === "End" ||
			ev.key === "PageDown" ||
			ev.key === "PageUp" ||
			ev.key === "NumLock" ||
			ev.key === "Escape" ||
			ev.key === "Tab" ||
			ev.key === "CapsLock" ||
			ev.key === "OS" ||
			ev.key === "ArrowLeft" ||
			ev.key === "ArrowRight" ||
			ev.key === "ArrowUp" ||
			ev.key === "ArrowDown" ||
			ev.key === "Control" ||
			ev.key === "Alt" ||
			ev.key === "AltGraph" ||
			ev.key === "Shift" ||
			ev.key === "Meta"
		) {
			return;
		}

		clearTimeout(searchTimeout);

		if (ev.key === "Enter") {
			handleSearch(ev);
		} else {
			searchTimeout = window.setTimeout(
				() => {
					handleSearch(ev);
				},
				isTouch() ? 800 : 400,
			);
		}
	}

	/**
	 * Processes the query in search bar after a new query is input and
	 * navigates us to the new search page.
	 */
	function handleSearch(ev: KeyboardEvent) {
		const target = ev.target as HTMLInputElement;
		if (!target) {
			console.error("handleSearch: Event target doesn't exist!");
			return;
		}
		if (!target.value.trim()) {
			console.debug("handleSearch: Search box has no valid value.");
			return;
		}
		// Only trim the start of the query incase the debounced search
		// fires before the user finishes typing the full query (which
		// would remove any last space if the query is multi-word).
		let query = target.value.trimStart();
		if (ev.key === "Enter") {
			// If enter is the key that triggered the search, we will
			// also trim the end of the query, since we can be sure
			// the user was done typing.
			query = query.trimEnd();
		}
		const currentSearchType = page.url.searchParams.get("type");
		const searchParams = new SvelteURLSearchParams({
			query: encodeURIComponent(query),
			preferMyList: "true",
		});
		if (page.route?.id === "/(app)/search" && currentSearchType) {
			// If we are already on the search page, we can attempt
			// to keep any existing type filter on the next query.
			searchParams.set("type", currentSearchType);
		}
		// Enable autofocus before running `goto` because on chromium
		// the .focus() call won't work, even after a timeout.
		// Using autofocus seems to work. Disables after goto runs.
		// https://github.com/sbondCo/Watcharr/issues/169
		target.autofocus = true;
		goto(resolve(`/search?${searchParams.toString()}`)).then(() => {
			// Use mainSearchEl if nav not split, otherwise use ev target.
			if (!document.body.classList.contains("split-nav") && mainSearchEl) {
				mainSearchEl.focus();
				mainSearchEl.autofocus = false;
			} else {
				target?.focus();
			}
			target.autofocus = false;
		});
	}

	async function getInitialData() {
		if (!localStorage.getItem("token")) {
			console.warn("getInitialData: No token found, redirecting to login!");
			goto(resolve("/login?again=1"));
			return;
		}
		const [u, s, f, fo, ts] = await Promise.all([
			req.get<PrivateUser>("/user"),
			req.get<UserSettings>("/user/settings"),
			req.get<ServerFeatures>("/features"),
			req.get<Follow[]>("/follow"),
			req.get<Tag[]>("/tag"),
		]);
		if (u) {
			store.userInfo = u;
		}
		if (s) {
			store.userSettings = s;
		}
		if (f) {
			store.serverFeatures = f;
		}
		if (fo) {
			store.follows = fo;
		}
		if (ts) {
			store.tags = ts;
		}
	}

	/**
	 * Adds or removed `split-nav` tag to body depending
	 * on how big the main search bar is.
	 */
	function decideOnNavSplit() {
		if (window.innerWidth <= 305) {
			document.body.classList.add("split-nav");
			return;
		}
		const bigInput = navEl?.querySelector("input:not(.small)");
		if (bigInput) {
			const b = bigInput.getBoundingClientRect();
			console.debug("decideOnNavSplit: bigInput width:", b.width);
			if (b.width <= 45) {
				document.body.classList.add("split-nav");
				console.debug("decideOnNavSplit: Splitting nav.");
			} else {
				document.body.classList.remove("split-nav");
				console.debug("decideOnNavSplit: Unsplitting nav.");
			}
		} else {
			console.warn("decideOnNavSplit: bigInput not found!", bigInput);
		}
	}

	function docOnScroll() {
		if (scroll > window.scrollY) {
			navEl?.classList.remove("scrolled-down");
			document.body.classList.add("nav-shown");
		} else {
			navEl?.classList.add("scrolled-down");
			document.body.classList.remove("nav-shown");
		}
		scroll = window.scrollY;
	}

	function focusSearch() {
		try {
			if (!mainSearchEl) {
				console.warn("focusSearch: mainSearchEl not defined!");
				return;
			}
			if (document.activeElement === mainSearchEl) {
				console.debug("focusSearch: mainSearchEl is already focused.");
				return;
			}
			mainSearchEl.focus();
		} catch (err) {
			console.error("focusSearch: Failed!", err);
		}
	}

	function handleGlobalKeybind(ev: KeyboardEvent) {
		switch (ev.key.toLowerCase()) {
			case "s":
				if (ev.ctrlKey) {
					ev.preventDefault();
					focusSearch();
				}
				break;
		}
	}

	afterNavigate(() => {
		decideOnNavSplit();
	});

	onMount(() => {
		if (navEl) {
			decideOnNavSplit();
			window.addEventListener("resize", decideOnNavSplit);
			window.document.addEventListener("scroll", docOnScroll);
			window.document.addEventListener("keydown", handleGlobalKeybind);

			return () => {
				window.removeEventListener("resize", decideOnNavSplit);
				window.document.removeEventListener("scroll", docOnScroll);
				window.document.removeEventListener("keydown", handleGlobalKeybind);
			};
		} else {
			console.error(
				"navEl doesn't exist, failed to initialize up/down listener",
			);
		}
	});
</script>

<nav bind:this={navEl}>
	<div class="wrapper">
		<div class="left-side">
			<a href={resolve("/")}>
				<span class="large">Watcharr</span>
				<span class="small">W</span>
			</a>
		</div>
		<div class="search">
			<input
				bind:this={mainSearchEl}
				type="text"
				placeholder="Search"
				bind:value={store.searchQuery}
				onkeydown={searchBarKeyDown}
			/>
			<Icon i="search" wh={19} />
		</div>
		<div class="right-side">
			<Buttons />
		</div>
	</div>
	<input
		class="small"
		type="text"
		placeholder="Search"
		bind:value={store.searchQuery}
		onkeydown={searchBarKeyDown}
	/>
</nav>

{#await getInitialData()}
	<Spinner />
{:then}
	{@render children?.()}
{:catch err}
	<Error
		pretty="Couldn't fetch app data!"
		error={err}
		onRetry={() => {
			location.reload();
		}}
	/>
{/await}

{#if page.state.handbook}
	<HandbookModal onClose={() => history.back()} />
{/if}

<style lang="scss">
	nav {
		display: flex;
		flex-flow: column;
		margin-bottom: 20px;
		padding: 10px 20px;
		position: sticky;
		top: 0;
		gap: 3px;
		z-index: 99990;
		transition: top 200ms ease-in-out;
		@include nav-blur;

		&:global(.scrolled-down) {
			top: -110px;
		}

		.wrapper {
			display: flex;
			flex-flow: row;
			gap: 20px;
			justify-content: space-between;
			align-items: center;

			.left-side,
			.right-side {
				/* This makes the logo on left and icons on right the same
				width, ensuring the main search bar can stay truly centered
				when possible. */
				flex: 1;
			}

			@media screen and (max-width: 435px) {
				gap: 15px;
			}
		}

		.left-side {
			a {
				display: inline-flex;
				text-decoration: none;
				font-family:
					"Shrikhand",
					system-ui,
					-apple-system,
					BlinkMacSystemFont;
				font-size: 35px;
				transition:
					-webkit-text-stroke 150ms ease,
					color 150ms ease,
					font-weight 150ms ease;

				&:hover,
				&:focus-visible {
					color: $bg-color;
					-webkit-text-stroke: 1.5px $text-color;
					font-weight: bold;
				}

				span.large {
					display: block;
					width: 185.2px;
				}

				span.small {
					display: none;
					width: 40px;
				}

				@media screen and (max-width: 620px) {
					span.large {
						display: none;
					}
					span.small {
						display: block;
					}
				}
			}
		}

		.search {
			width: 100%;
			position: relative;

			// Make the box look a little more centered, inline with the rest of the nav items.
			margin-bottom: 2px;

			:global(svg) {
				display: none;
				position: absolute;
				top: 50%;
				left: 50%;
				transform: translate(-50%, -50%);
				pointer-events: none;
				user-select: none;
			}

			input:focus-within + :global(svg),
			input:not(:placeholder-shown) + :global(svg) {
				display: none;
			}

			@media screen and (min-width: 666px) {
				max-width: 250px;
			}

			@media screen and (max-width: 666px) {
				& input:not(.small) {
					width: 100%;
				}

				&:focus-within + .right-side {
					display: none;
				}
			}

			@media screen and (max-width: 460px) {
				:global(svg) {
					display: block;
				}

				input::placeholder {
					color: transparent;
				}
			}
		}

		:global(body.split-nav) & {
			.search {
				/* We hide with visibility: hidden, so the decideOnNavSplit can
				still get the search width for it's decision logic. */
				opacity: 0;
				visibility: hidden;
			}

			input.small {
				display: block;
			}
		}

		input {
			width: 100%;
			font-weight: bold;
			text-align: center;
			color: black;
			background-color: white;
			box-shadow: 4px 4px 0px 0px $text-color;
			border: 2px solid black;
			text-overflow: ellipsis;
			transition:
				width 150ms ease,
				box-shadow 150ms ease;

			&.small {
				display: none;
				margin-left: auto;
				margin-right: auto;
			}

			&:hover,
			&:focus {
				box-shadow: 2px 2px 0px 0px $text-color;
			}

			@media screen and (max-width: 290px) {
				&.small {
					width: 100%;
				}
			}
		}
	}
</style>
