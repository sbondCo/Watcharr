import { writable } from "svelte/store";
import type { Watched } from "./types";
import type { Notification } from "./lib/util/notify";
import { browser } from "$app/environment";

export const watchedList = writable<Watched[]>([]);
export const notifications = writable<(Notification & { id: number })[]>([]);
export const activeFilter = writable<string[]>(["DATEADDED", "DOWN"]);

export const clearAllStores = () => {
  watchedList.set([]);
  notifications.set([]);
  activeFilter.set([]);
};

// Rehydrate
if (browser) {
  const raf = localStorage.getItem("activeFilter");
  if (raf) {
    activeFilter.update((v) => (v = JSON.parse(raf)));
  }
}

// Save changes
activeFilter.subscribe((v) => {
  if (browser) localStorage.setItem("activeFilter", JSON.stringify(v));
});
