import { error } from "@sveltejs/kit";
import type { TMDBMovieDetails } from "@/types";
import axios from "axios";
import type { PageLoad } from "../../search/[query]/$types";

export const load = (async ({ params }) => {
  const { id } = params;

  if (!id) {
    error(400);
    return;
  }

  try {
    return (await axios.get(`/content/movie/${id}`)).data as TMDBMovieDetails;
  } catch (err: any) {
    if (err.response) {
      error(500, err.response.data.error);
    } else {
      error(500, err.message);
    }
  }
}) satisfies PageLoad;
