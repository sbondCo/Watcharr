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
	Watched,
} from "./types";
import type { Notification } from "./lib/util/notify";
import { browser } from "$app/environment";
import { toggleTheme } from "./lib/util/theme";

export const defaultSort = ["DATEADDED", "DOWN"];

interface Store {
	userInfo: PrivateUser | undefined;
	userSettings: UserSettings | undefined;
	watchedList: Watched[];
	notifications: Notification[];
	activeSort: string[];
	activeFilters: Filters;
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
	watchedList: [],
	notifications: [],
	activeSort: defaultSort,
	activeFilters: { type: [], status: [] },
	appTheme: "system",
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

/**
 * Expose store to app through getters/setters
 * to control what can and can't be accessed.
 * With setters we can easily and more reliably
 * save certain properties to localStorage when
 * they are updated.
 */
export const store = {
	get watchedList() {
		return _store.watchedList;
	},
	set watchedList(w) {
		_store.watchedList = w;
	},
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
		console.debug("Store: Saved activeSort:", v);
	},
	get activeFilters() {
		return _store.activeFilters;
	},
	set activeFilters(v) {
		_store.activeFilters = v;
		localStorage.setItem("activeFilterReal", JSON.stringify(v));
		console.debug("Store: Saved activeFilters:", v);
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
	store.watchedList = [];
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
