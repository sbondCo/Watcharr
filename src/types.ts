export type WatchedStatus = "PLANNED" | "WATCHING" | "FINISHED" | "HOLD" | "DROPPED";
export type ContentType = "tv" | "movie";
export type MediaType = ContentType | "person";

// Wasn't able to figure out how to import this type from its component file in other places, so its here for now.
export type Icon =
  | "check"
  | "clock"
  | "calendar"
  | "thumb-down"
  | "play"
  | "pause"
  | "jellyfin"
  | "trash"
  | "close"
  | "filter";

export interface Content {
  // id: number; // Not used
  tmdbId: number;
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

export interface TMDBContentCredits {
  id: number;
  cast: {
    adult: boolean;
    gender: number;
    id: number;
    known_for_department: string;
    name: string;
    original_name: string;
    popularity: number;
    profile_path: string;
    cast_id: number;
    character: string;
    credit_id: string;
    order: number;
  }[];
  crew: TMDBContentCreditsCrew[];
}

export interface TMDBContentCreditsCrew {
  adult: boolean;
  gender: number;
  id: number;
  known_for_department: string;
  name: string;
  original_name: string;
  popularity: number;
  profile_path: string;
  credit_id: string;
  department: string;
  job: string;
}

export interface TMDBPersonDetails {
  birthday?: string;
  known_for_department?: string;
  deathday?: string;
  id?: number;
  name?: string;
  also_known_as?: string[];
  gender?: number;
  biography?: string;
  popularity?: number;
  place_of_birth?: string;
  profile_path?: string;
  adult?: boolean;
  imdb_id?: string;
  homepage?: string;
}

export interface TMDBPersonCombinedCredits {
  id: number;
  cast: TMDBPersonCombinedCreditsCast[];
}

export interface TMDBPersonCombinedCreditsCast {
  id: number;
  original_language: string;
  episode_count: number;
  overview: string;
  origin_country: string[];
  original_name: string;
  genre_ids: number[];
  name: string;
  media_type: MediaType;
  poster_path: string;
  first_air_date: string;
  vote_average: number;
  vote_count: number;
  character: string;
  backdrop_path: string;
  popularity: number;
  credit_id: string;
  original_title: string;
  video: boolean;
  release_date: string;
  title: string;
  adult: boolean;
}
