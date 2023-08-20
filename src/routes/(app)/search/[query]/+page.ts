import { error } from "@sveltejs/kit";

export interface ContentSearch {
  page: number;
  results: (ContentSearchMovie | ContentSearchTv | ContentSearchPerson)[];
  total_pages: number;
  total_results: number;
}

export interface ContentSearchMovie {
  poster_path?: string;
  adult?: boolean;
  overview?: string;
  release_date?: string;
  original_title?: string;
  genre_ids?: number[];
  id: number;
  media_type: "movie";
  original_language?: string;
  title?: string;
  backdrop_path?: string;
  popularity?: number;
  vote_count?: number;
  video?: boolean;
  vote_average?: number;
}

export interface ContentSearchTv {
  poster_path?: string;
  popularity?: number;
  id: number;
  overview?: string;
  backdrop_path?: string;
  vote_average?: number;
  media_type: "tv";
  first_air_date?: string;
  origin_country?: string[];
  genre_ids?: number[];
  original_language?: string;
  vote_count?: number;
  name?: string;
  original_name?: string;
}

export interface ContentSearchPerson {
  profile_path?: string;
  adult?: boolean;
  id?: number;
  media_type: "person";
  known_for?: ContentSearchMovie | ContentSearchTv;
  name?: string;
  popularity?: number;
}

export async function load({ params }) {
  const { query } = params;

  if (!query) {
    error(400);
    return;
  }

  return {
    slug: query
  };
}
