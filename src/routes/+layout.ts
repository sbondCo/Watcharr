export const prerender = false;
export const ssr = false;
export const csr = true;

import { error } from "@sveltejs/kit";
import type { LayoutLoad } from "./$types";
import req from "@/lib/api";
import { watchedList } from "@/store";

export const load = (async () => {
  try {
    if (!localStorage.getItem("token")) {
      return;
    }
    const w = await req("/watched", "GET");
    watchedList.update((wl) => (wl = w.data));
  } catch (err) {
    console.error("Error loading watched content:", err);
    error(500, "Error loading watched content!");
  }
}) satisfies LayoutLoad;
