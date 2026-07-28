import type {
	Filters,
	Follow,
	ImportedList,
	PrivateUser,
	ServerFeatures,
	Tag,
	Theme,
	UserSettings,
	WLDetailedViewOption,
} from "./types";
import type { Notification } from "./lib/util/notify";
import { browser } from "$app/environment";
import { toggleTheme } from "./lib/util/theme";

export const defaultSort = ["DATEADDED", "DOWN"];

export type WatchedListMode = "all" | "tv" | "movie";

/** Type filter array for a given view mode. */
export function typeForMode(mode: WatchedListMode): string[] {
	if (mode === "tv") return ["tv"];
	if (mode === "movie") return ["movie"];
	return [];
}

/**
 * Derive the current view mode from a type filter, or undefined when the
 * type filter is a combination that no single mode represents.
 */
function modeOf(type: string[] | undefined): WatchedListMode | undefined {
	if (!type?.length) return "all";
	if (type.length === 1 && type[0] === "tv") return "tv";
	if (type.length === 1 && type[0] === "movie") return "movie";
	return undefined;
}

interface Store {
	userInfo: PrivateUser | undefined;
	userSettings: UserSettings | undefined;
	notifications: Notification[];
	activeSort: string[];
	activeFilters: Filters;
	// Remembered status filter and sort for each watched-list view mode
	// (all/tv/movie), so switching between shows and movies keeps each mode's
	// own filters and sort order.
	filterModes: Record<string, string[]>;
	sortModes: Record<string, string[]>;
	sortAndFiltersForQueryParams: {};
	appTheme: Theme;
	importedList:
		| {
				data: string;
				type:
					| "text-list"
					| "tmdb"
					| "movary"
					| "watcharr"
					| "myanimelist"
					| "ryot"
					| "todomovies"
					| "imdb";
		  }
		| undefined;
	parsedImportedList: ImportedList[] | undefined;
	searchQuery: string;
	serverFeatures: ServerFeatures | undefined;
	follows: Follow[];
	wlDetailedView: WLDetailedViewOption[];
	tags: Tag[];
}

/**
 * This is our actual (private) store.
 */
const _store: Store = $state({
	notifications: [],
	activeSort: defaultSort,
	activeFilters: { type: [], status: [] },
	filterModes: { all: [], tv: [], movie: [] },
	sortModes: { all: defaultSort, tv: defaultSort, movie: defaultSort },
	appTheme: "system",
	sortAndFiltersForQueryParams: {},
	importedList: undefined,
	parsedImportedList: undefined,
	searchQuery: "",
	userInfo: undefined,
	userSettings: undefined,
	serverFeatures: undefined,
	follows: [],
	wlDetailedView: [],
	tags: [],
});

const updateSortAndFiltersForQueryParams = () => {
	try {
		const qp: any = {};
		if (store.activeSort?.length === 2) {
			qp.sort = store.activeSort[0];
			qp.sortDir = store.activeSort[1] === "UP" ? "asc" : "desc";
		}
		if (store.activeFilters) {
			const t = store.activeFilters?.type?.join(",");
			if (t) {
				qp["type"] = t;
			}
			const s = store.activeFilters?.status?.join(",");
			if (s) {
				qp["status"] = s;
			}
		}
		_store.sortAndFiltersForQueryParams = qp;
	} catch (err) {
		console.error("updateSortAndFiltersForQueryParams: Failed!", err);
		_store.sortAndFiltersForQueryParams = {};
	}
};

/**
 * Expose store to app through getters/setters
 * to control what can and can't be accessed.
 * With setters we can easily and more reliably
 * save certain properties to localStorage when
 * they are updated.
 */
export const store = {
	get notifications() {
		return _store.notifications;
	},
	set notifications(v) {
		_store.notifications = v;
	},
	get activeSort() {
		return _store.activeSort;
	},
	set activeSort(v) {
		_store.activeSort = v;
		localStorage.setItem("activeFilter", JSON.stringify(v));
		// Remember this mode's sort so switching modes restores it.
		const mode = modeOf(_store.activeFilters?.type);
		if (mode) {
			_store.sortModes[mode] = v;
			localStorage.setItem("sortModes", JSON.stringify(_store.sortModes));
		}
		console.debug("Store: Saved activeSort:", v);
		updateSortAndFiltersForQueryParams();
	},
	get activeFilters() {
		return _store.activeFilters;
	},
	get hasActiveFilters(): boolean {
		return (
			this.activeFilters &&
			(this.activeFilters.status?.length > 0 ||
				this.activeFilters.type?.length > 0)
		);
	},
	set activeFilters(v) {
		_store.activeFilters = v;
		localStorage.setItem("activeFilterReal", JSON.stringify(v));
		// Remember this mode's status filter so switching modes restores it.
		const mode = modeOf(v?.type);
		if (mode) {
			_store.filterModes[mode] = v?.status ?? [];
			localStorage.setItem("filterModes", JSON.stringify(_store.filterModes));
		}
		console.debug("Store: Saved activeFilters:", v);
		updateSortAndFiltersForQueryParams();
	},
	get filterModes() {
		return _store.filterModes;
	},
	get sortModes() {
		return _store.sortModes;
	},
	/**
	 * Return our `activeSort` and `activeFilters` in an object
	 * that is in the correct format for our get watched page
	 * requests (object that is given to axios for query params).
	 */
	get sortAndFiltersForQueryParams() {
		return _store.sortAndFiltersForQueryParams;
	},
	get appTheme() {
		return _store.appTheme;
	},
	/**
	 * Only set appTheme through toggleTheme() helper.
	 */
	set appTheme(v) {
		_store.appTheme = v;
		localStorage.setItem("theme", v);
		console.debug("Store: Saved appTheme:", v);
	},
	get importedList() {
		return _store.importedList;
	},
	set importedList(v) {
		_store.importedList = v;
	},
	get parsedImportedList() {
		return _store.parsedImportedList;
	},
	set parsedImportedList(v) {
		_store.parsedImportedList = v;
	},
	get searchQuery() {
		return _store.searchQuery;
	},
	set searchQuery(v) {
		_store.searchQuery = v;
	},
	get userInfo() {
		return _store.userInfo;
	},
	set userInfo(v) {
		_store.userInfo = v;
	},
	get userSettings() {
		return _store.userSettings;
	},
	set userSettings(v) {
		_store.userSettings = v;
	},
	get serverFeatures() {
		return _store.serverFeatures;
	},
	set serverFeatures(v) {
		_store.serverFeatures = v;
	},
	get follows() {
		return _store.follows;
	},
	set follows(v) {
		_store.follows = v;
	},
	get wlDetailedView() {
		return _store.wlDetailedView;
	},
	set wlDetailedView(v) {
		_store.wlDetailedView = v;
		if (v) {
			localStorage.setItem(
				"wlDetailedView",
				JSON.stringify(store.wlDetailedView),
			);
			console.debug("Store: Saved wlDetailedView:", v);
		} else {
			localStorage.removeItem("wlDetailedView");
			console.debug("Store: Removed wlDetailedView");
		}
	},
	get tags() {
		return _store.tags;
	},
	set tags(v) {
		_store.tags = v;
	},
};

/**
 * Reset everything in `store` back to default values.
 */
export const clearAllStores = () => {
	store.notifications = [];
	store.activeSort = defaultSort;
	store.appTheme = "system";
	store.importedList = undefined;
	store.parsedImportedList = undefined;
	store.searchQuery = "";
	store.userInfo = undefined;
	store.userSettings = undefined;
	store.serverFeatures = undefined;
	store.follows = [];
	store.wlDetailedView = [];
	store.tags = [];
	clearActiveFilters();
};

export const clearActiveFilters = () => {
	store.activeFilters = { type: [], status: [] };
};

/**
 * Switch the watched-list view mode (all/tv/movie), restoring the status
 * filter last used in that mode.
 */
export const setWatchedListMode = (mode: WatchedListMode) => {
	// Set the type filter first so the mode is derivable, then restore this
	// mode's remembered status and sort.
	store.activeFilters = {
		...store.activeFilters,
		type: typeForMode(mode),
		status: store.filterModes[mode] ?? [],
	};
	store.activeSort = store.sortModes[mode] ?? defaultSort;
};

if (browser) {
	rehydrateStore();
}

/**
 * Restore state from localStorage and apply values into
 * our `store`.
 * Rehydrates directly into `_store` (the real store)
 * to avoid the setters that would trigger a save right
 * after rehydrate.
 */
function rehydrateStore() {
	console.info("rehydrateStore: Running..");
	// Restore activeSort
	const raf = localStorage.getItem("activeFilter");
	if (raf) {
		_store.activeSort = JSON.parse(raf);
		console.debug(
			"rehydrateStore: Restored activeSort:",
			$state.snapshot(store.activeSort),
		);
	}
	// Restore activeFilters
	const filters = localStorage.getItem("activeFilterReal");
	if (filters) {
		_store.activeFilters = JSON.parse(filters);
		console.debug(
			"rehydrateStore: Restored activeFilters:",
			$state.snapshot(store.activeFilters),
		);
	}
	// Restore per-mode remembered filters
	const fm = localStorage.getItem("filterModes");
	if (fm) {
		_store.filterModes = { all: [], tv: [], movie: [], ...JSON.parse(fm) };
		console.debug(
			"rehydrateStore: Restored filterModes:",
			$state.snapshot(store.filterModes),
		);
	}
	// Restore per-mode remembered sort
	const sm = localStorage.getItem("sortModes");
	if (sm) {
		_store.sortModes = {
			all: defaultSort,
			tv: defaultSort,
			movie: defaultSort,
			...JSON.parse(sm),
		};
		console.debug(
			"rehydrateStore: Restored sortModes:",
			$state.snapshot(store.sortModes),
		);
	}
	// After restoring activeSort and activeFilter, set
	// an initial value for our related query param state.
	updateSortAndFiltersForQueryParams();
	// Restore appTheme
	const theme = localStorage.getItem("theme") as Theme;
	if (theme) {
		_store.appTheme = theme;
		toggleTheme(theme, false);
		console.debug(
			"rehydrateStore: Restored appTheme:",
			$state.snapshot(store.appTheme),
		);
	} else {
		let defTheme: Theme = "system";
		_store.appTheme = defTheme;
		toggleTheme(defTheme, false);
		console.debug(
			"rehydrateStore: appTheme hydrated from system default (wont save):",
			defTheme,
		);
	}
	// Restore wlDetailedView
	const wlDetailedViewR = localStorage.getItem("wlDetailedView");
	if (wlDetailedViewR) {
		_store.wlDetailedView = JSON.parse(wlDetailedViewR);
		console.debug(
			"rehydrateStore: Restored wlDetailedView:",
			$state.snapshot(store.wlDetailedView),
		);
	}
	console.info("rehydrateStore: Done.");
}
