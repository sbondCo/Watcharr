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
 * Ensure `destroy` func is ran after the component
 * this is used in is destroyed.
 */
export default function infScroll(opts: ToolTipOptions) {
	let { threshold = 150, callback } = opts;

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
		console.debug("infiniteScroll()");
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

	addEvents();

	return {
		isAtBottom,
		run,
		destroy: () => {
			console.debug("infScroll->destroy()");
			removeEvents();
		},
	};
}
