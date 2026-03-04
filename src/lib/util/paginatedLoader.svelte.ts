import type { PaginationResponse } from "@/types";
import type { AxiosResponse, GenericAbortSignal } from "axios";

export enum PaginatedLoaderRunFnAction {
	// Reset state before running fn. Simplifies running a clean request.
	// Eg: We are running a clean request because filters have changed.
	Reset = 1,
	// Mainly for use for 'Try Again' buttons.
	// Eg: If we fail on trying to load the first page, the reset should also
	// reset state to do a full clean retry, but if we fail on 2nd, etc page,
	// then we can't reset state otherwise we'd be resetting our page num,
	// results, etc.
	ResetIfOnFirstOrNoPage = 2,
}

export default function paginatedLoader<T, U>(
	fn: (
		sig: GenericAbortSignal,
	) => Promise<AxiosResponse<PaginationResponse<T, U>, any> | undefined>,
) {
	let reqController = new AbortController();

	const state: {
		data: T[];
		meta?: U;
		page: number;
		pageMax: number;
		reqLoading: boolean;
		reqLoadError: Error | undefined;
	} = $state({
		data: [],
		meta: undefined,
		page: 0,
		pageMax: 1,
		reqLoading: false,
		reqLoadError: undefined,
	});

	/**
	 * Resets state.
	 */
	const reset = () => {
		state.data = [];
		state.page = 0;
		state.pageMax = 1;
		state.reqLoading = false;
		state.reqLoadError = undefined;
		// Abort any existing request to ensure we don't end up having multiple
		// at same time.
		abortReq("state reset");
		console.log("paginatedLoader->reset(): Finished.");
	};

	/**
	 * Abort the request.
	 */
	const abortReq = (reason: string) => {
		reqController.abort(reason);
	};

	const runFnAction = (action: PaginatedLoaderRunFnAction) => {
		switch (action) {
			case PaginatedLoaderRunFnAction.Reset:
				reset();
				break;
			case PaginatedLoaderRunFnAction.ResetIfOnFirstOrNoPage:
				if (state.page <= 1) {
					reset();
				}
				break;
		}
	};

	/**
	 * Runs the paginated request passed in as `fn` with our
	 * paginated logic wrapped around.
	 */
	const runFn = async (action?: PaginatedLoaderRunFnAction) => {
		const logStyle = "font-weight: bold; font-size: 18px;";

		if (action) {
			runFnAction(action);
		}

		if (state.reqLoading) {
			console.warn("%cpaginatedLoader->runFn: already running", logStyle);
			return;
		}
		if (state.page >= state.pageMax) {
			console.warn("%cpaginatedLoader->runFn: max page reached", logStyle);
			return;
		}

		state.reqLoading = true;
		reqController = new AbortController();
		try {
			const resp = await fn(reqController.signal);
			if (!resp) {
				state.reqLoading = false;
				console.error(
					"paginatedLoader->runFn: fn returned nothing! May not have wanted to for some reason..",
				);
				return;
			}
			state.page = resp.data.page;
			state.pageMax = resp.data.totalPages;
			console.debug(
				`%cpaginatedLoader->runFn: Loaded Page=${state.page} Max=${state.pageMax}`,
				logStyle,
			);
			if (!resp.data.results || resp.data.results.length <= 0) {
				state.reqLoading = false;
				console.warn("loadWatchedList: No results.");
				return;
			}
			state.data.push(...resp.data.results);
			state.data = state.data;
			state.meta = resp.data.meta;
		} catch (err: any) {
			if (err?.code === "ERR_CANCELED") {
				console.warn("loadWatchedList: Cancelled, not showing error.");
				// If request cancelled (likely by us aborting), then return
				// here to avoid updating reqLoading state to false below.
				// This fixes the case where we abort a request and start the
				// next one before this one throws, which sets reqLoading to
				// false for our next request (race condition).
				return;
			} else {
				console.error("loadWatchedList: failed!", err);
				state.reqLoadError = err;
			}
		}
		state.reqLoading = false;
	};

	return {
		runFn,
		reset,
		abortReq,
		state,
	};
}
