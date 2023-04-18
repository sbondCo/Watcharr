export type WatchedStatus = "PLANNED" | "WATCHING" | "FINISHED" | "HOLD" | "DROPPED";
export type ContentType = "tv" | "movie";
export type MediaType = ContentType | "person";

export interface Content {
  id: number;
  title: string;
  poster_path: string;
  overview: string;
  type: ContentType;
}

export interface Watched {
  id: number;
  watched: boolean;
  rating?: number;
  content: Content;
  status: WatchedStatus;
}

export interface WatchedAddRequest {
  contentId: number;
  contentType: ContentType;
  rating?: number;
  status: WatchedStatus;
}

export interface WatchedUpdateRequest {
  rating?: number;
  status?: WatchedStatus;
}

export interface TMDBContentDetails {
  id: number;
  backdrop_path: string;
  genres: {
    id: number;
    name: string;
  }[];
  poster_path: string;
  homepage: string;
  popularity: number;
  overview: string;
  original_language: string;
  production_companies: {
    id: number;
    logo_path: string;
    name: string;
    origin_country: string;
  }[];
  production_countries: {
    iso_3166_1: string;
    name: string;
  }[];
  status: string;
  tagline: string;
  vote_average: number;
  vote_count: number;
  spoken_languages: {
    english_name: string;
    iso_639_1: string;
    name: string;
  }[];
}

export interface TMDBMovieDetails extends TMDBContentDetails {
  adult: boolean;
  belongs_to_collection: any;
  budget: number;
  imdb_id: string;
  original_title: string;
  release_date: string;
  revenue: number;
  runtime: number;
  title: string;
  video: boolean;
}

export interface TMDBShowDetails extends TMDBContentDetails {
  created_by: {
    id: number;
    credit_id: string;
    name: string;
    gender: number;
    profile_path: string;
  }[];
  episode_run_time: number[];
  first_air_date: string;
  in_production: boolean;
  languages: string[];
  last_air_date: string;
  last_episode_to_air: {
    air_date: string;
    episode_number: number;
    id: number;
    name: string;
    overview: string;
    production_code: string;
    season_number: number;
    still_path: string;
    vote_average: number;
    vote_count: number;
  };
  name: string;
  next_episode_to_air: any;
  networks: {
    name: string;
    id: number;
    logo_path: string;
    origin_country: string;
  }[];
  number_of_episodes: number;
  number_of_seasons: number;
  origin_country: string[];
  original_name: string;
  seasons: {
    air_date: string;
    episode_count: number;
    id: number;
    name: string;
    overview: string;
    poster_path: string;
    season_number: number;
  }[];
  type: string;
}
