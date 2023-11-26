import { writable } from "svelte/store";
import type {
  FileWithData,
  Filters,
  Follow,
  ImportedList,
  PrivateUser,
  ServerFeatures,
  Theme,
  UserSettings,
  Watched
} from "./types";
import type { Notification } from "./lib/util/notify";
import { browser } from "$app/environment";
import { toggleTheme } from "./lib/util/helpers";

export const userInfo = writable<PrivateUser | undefined>();
export const userSettings = writable<UserSettings | undefined>();
export const watchedList = writable<Watched[]>([]);
export const notifications = writable<Notification[]>([]);
export const activeSort = writable<string[]>(["DATEADDED", "DOWN"]);
export const activeFilters = writable<Filters>({ type: [], status: [] });
export const appTheme = writable<Theme>();
export const importedList = writable<FileWithData | undefined>();
export const parsedImportedList = writable<ImportedList[] | undefined>();
export const searchQuery = writable<string>("");
export const serverFeatures = writable<ServerFeatures>();
export const follows = writable<Follow[]>();

export const clearAllStores = () => {
  watchedList.set([]);
  notifications.set([]);
  activeSort.set(["DATEADDED", "DOWN"]);
  activeFilters.set({ type: [], status: [] });
  importedList.set(undefined);
  parsedImportedList.set(undefined);
  searchQuery.set("");
  userInfo.set(undefined);
  userSettings.set(undefined);
  follows.set([]);
};

if (browser) {
  // Rehydrate
  const raf = localStorage.getItem("activeFilter");
  if (raf) {
    activeSort.update((v) => (v = JSON.parse(raf)));
  }

  const filters = localStorage.getItem("activeFilterReal");
  if (filters) {
    activeFilters.update((v) => (v = JSON.parse(filters)));
  }

  const theme = localStorage.getItem("theme") as Theme;
  if (theme) {
    appTheme.update((t) => (t = theme));
    toggleTheme(theme);
  } else {
    let defTheme: Theme = "light";
    if (window.matchMedia("(prefers-color-scheme: dark)").matches) {
      defTheme = "dark";
    }
    console.log("Theme not set, setting default theme from system theme:", defTheme);
    appTheme.update((t) => (t = defTheme));
    toggleTheme(defTheme);
  }

  // Save changes
  activeSort.subscribe((v) => {
    localStorage.setItem("activeFilter", JSON.stringify(v));
  });

  activeFilters.subscribe((v) => {
    localStorage.setItem("activeFilterReal", JSON.stringify(v));
  });

  appTheme.subscribe((v) => {
    localStorage.setItem("theme", v);
  });
}
