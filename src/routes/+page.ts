import { goto } from "$app/navigation";
import { watchedList } from "@/store";
import { error } from "@sveltejs/kit";
import axios from "axios";
import type { PageLoad } from "./$types";

export const load = (async () => {
  try {
    if (localStorage.getItem("token")) {
      const w = await axios.get("/watched");
      if (w?.data?.length > 0) {
        watchedList.update((wl) => (wl = w.data));
      }
    } else {
      goto("/login?again=1");
    }
  } catch (err) {
    console.error("Error loading watched content:", err);
    error(500, "Error loading watched content!");
  }
}) satisfies PageLoad;
