import { error } from "@sveltejs/kit";
import type { Content, MediaType } from "@/types";
import axios from "axios";

export interface ContentSearch {
  page: number;
  results: ContentSearchResult[];
  total_pages: number;
  total_results: number;
}

export interface ContentSearchResult {
  adult: boolean;
  backdrop_path: string;
  id: number;
  original_language: string;
  overview: string;
  poster_path: string;
  media_type: MediaType;
  genre_ids?: number[];
  popularity: number;
  vote_average: number;
  vote_count: number;
  name?: string;
  original_name?: string;
  first_air_date?: string;
  origin_country?: string[];
  title?: string;
  original_title?: string;
  release_date?: string;
  video?: boolean;
}

export async function load({ params }) {
  const { query } = params;

  if (!query) {
    error(400);
    return;
  }

  try {
    return (await axios.get(`/content/${query}`)).data as ContentSearch;
  } catch (err: any) {
    if (err.response) {
      error(500, err.response.data.error);
    } else {
      error(500, err.message);
    }
  }
}
