import { writable } from "svelte/store";
import type { Watched } from "./types";
import type { Notification } from "./lib/util/notify";

export const watchedList = writable<Watched[]>([]);
export const notifications = writable<(Notification & { id: number })[]>([]);

export const clearAllStores = () => {
  watchedList.set([]);
  notifications.set([]);
};
