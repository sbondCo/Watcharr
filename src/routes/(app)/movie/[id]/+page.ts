import { error } from "@sveltejs/kit";
import type { PageLoad } from "../../search/[query]/$types";

export const load = (async ({ params }) => {
  const { id } = params;

  if (!id) {
    error(400);
    return;
  }

  return {
    movieId: Number(id)
  };
}) satisfies PageLoad;
