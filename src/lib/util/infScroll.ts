import { page } from "$app/state";

export interface ToolTipOptions {
	/**
	 * Infinite scroll threshold.
	 * Distance from bottom.
	 */
	threshold?: number;
	/**
	 * Ran when we reach the end of scroll.
	 * This callback should load the new data.
	 */
	callback: () => Promise<void>;
}

/**
 * Infinite scroll helper.
 *
 * **Ensure:**
 * - `destroy()` is ran after the component
 * this is used in is destroyed.
 * - `dataLoaded()` is ran after a page of data
 * is loaded.
 * - `opts.callback()` has its own 'isLoading' logic
 * to prevent itself from running extra times while
 * still loading a previous request. This will be the
 * thing stopping extra scrolls while data is still
 * loading from causing extra data requests, etc.
 */
export default function infScroll(opts: ToolTipOptions) {
	let { threshold = 150, callback } = opts;

	// Store current pathname at point of infScroll
	// initialization, this ensures we have a point
	// of reference for our fix below (that ensures
	// we don't allow asking for next data load if
	// user navigates to a different page).
	const startPagePathLower = page.url?.pathname?.toLowerCase();

	const addEvents = () => {
		console.debug("infScroll->addEvents()");
		window.addEventListener("scroll", run);
		window.addEventListener("resize", run);
	};

	const removeEvents = () => {
		console.debug("infScroll->removeEvents()");
		window.removeEventListener("scroll", run);
		window.removeEventListener("resize", run);
	};

	const isAtBottom = () => {
		return (
			window.innerHeight + Math.round(window.scrollY) + threshold >=
			document.body.offsetHeight
		);
	};

	const run = async () => {
		if (isAtBottom()) {
			console.log("infiniteScroll: Reached end");
			removeEvents();
			await callback();
			addEvents();
			console.log("infiniteScroll: Callback ran");
		} else {
			console.debug("infiniteScroll: Not at bottom");
		}
	};

	/**
	 * This needs to be called after data is loaded.
	 * It runs the infScroll logic incase the user is
	 * already at the bottom after initial data load
	 * without scroll/resize (eg: using high dpi screen).
	 */
	const dataLoaded = () => {
		// If results don't fill the page enough to enable scrolling,
		// the user could be stuck and not be able to get more results
		// to show, run `infiniteScroll` to load more if we can.
		// Smol timeout to give ui time to render so end of page calc
		// can be accurate.
		setTimeout(() => {
			// Quick fix, if user navigates away from search page while response is loading,
			// we don't want to call infiniteScroll or we could end up loading all pages
			// in the background.
			if (startPagePathLower === page.url?.pathname?.toLowerCase()) {
				console.debug(
					"infiniteScroll->dataLoaded(): Still at bottom.. asking for more data.",
				);
				run();
			} else {
				console.debug(
					"infiniteScroll->dataLoaded(): No longer on initial page, not calling infiniteScroll.",
				);
			}
		}, 250);
	};

	addEvents();

	return {
		isAtBottom,
		run,
		dataLoaded,
		destroy: () => {
			console.debug("infScroll->destroy()");
			removeEvents();
		},
	};
}
