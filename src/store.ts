import { writable } from "svelte/store";
import type { Watched } from "./types";

export const watchedList = writable<Watched[]>([]);
