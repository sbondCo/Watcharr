import { error } from "@sveltejs/kit";
import type { PageLoad } from "./$types";
import type { Watched } from "@/types";
import req from "@/lib/api";

export const load = (async () => {
  try {
    const w = await req("/watched", "GET");

    return {
      watched: w.data as Watched[]
    };
  } catch (err) {
    console.error("Error loading watched content:", err);
    error(500, "Error loading watched content!");
  }
}) satisfies PageLoad;
