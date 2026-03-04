export type WatchedStatus =
	| "PLANNED"
	| "WATCHING"
	| "FINISHED"
	| "HOLD"
	| "DROPPED";
/**
 * Types of media supported by Watcharr in an overarching way.
 */
export type SupportedMedia = "tv" | "movie" | "game";
export type ContentType = SupportedMedia | "tv_episode";
export type MediaType = ContentType | "person";

// Wasn't able to figure out how to import this type from its component file in other places, so its here for now.
export type Icon =
	| "check"
	| "clock"
	| "calendar"
	| "thumb-down"
	| "thumb-up"
	| "play"
	| "pause"
	| "jellyfin"
	| "emby"
	| "plex"
	| "trash"
	| "close"
	| "close-circle"
	| "filter"
	| "filter-circle"
	| "reel"
	| "compass"
	| "document"
	| "add"
	| "arrow"
	| "chevron"
	| "search"
	| "sort"
	| "eye-closed"
	| "people-nocircle"
	| "people"
	| "person"
	| "person-add"
	| "person-minus"
	| "pencil"
	| "eye"
	| "star"
	| "movary"
	| "ryot"
	| "trakt"
	| "myanimelist"
	| "todomovies"
	| "themoviedb"
	| "imdb"
	| "refresh"
	| "gamepad"
	| "film"
	| "tv"
	| "pin"
	| "unpin"
	| "sparkles"
	| "tag"
	| "ticket"
	| "lock-closed"
	| "github"
	| "website"
	| "tmdb"
	| "igdb";

export type Theme = "light" | "dark" | "system";

export type WLDetailedViewOption =
	| "statusRating"
	| "lastWatched"
	| "dateAdded"
	| "dateModified";

export enum UserType {
	Watcharr = 0,
	Jellyfin = 1,
	Plex = 2,
	Proxy = 3,
}

interface dbModel {
	id: number;
	createdAt: string;
	updatedAt: string;
	deletedAt: string;
}

export interface PaginationResponse<T, U> {
	limit: number;
	page: number;
	totalPages: number;
	totalResults: number;
	results: T[] | null;
	meta?: U;
}

export interface Content {
	// id: number; // Not used
	tmdbId: number;
	title: string;
	poster_path: string;
	overview: string;
	type: ContentType;
	release_date: string;
	first_air_date: string;
}

export interface Activity extends dbModel {
	watchedId: number;
	type: string;
	data: string;
	customDate: string;
}

export interface WatchedSeason {
	id: number;
	watchedID: number;
	status: WatchedStatus;
	rating: number;
	seasonNumber: number;
}

export interface WatchedEpisode {
	id: number;
	watchedID: number;
	status: WatchedStatus;
	rating: number;
	seasonNumber: number;
	episodeNumber: number;
}

/**
 * Same as `WatchedDto` on server.
 */
export interface Watched {
	id: number;
	createdAt: string;
	updatedAt: string;

	rating?: number;
	status: WatchedStatus;
	thoughts: string;
	pinned: boolean;

	activity?: Activity[];
	watchedSeasons?: WatchedSeason[];
	watchedEpisodes?: WatchedEpisode[];
	tags?: Tag[];
	lastViewedSeason?: number;

	// 'Watching Season/Ep' Extra detail.
	watchingSeason?: string;

	media?: Media;
}

export interface WatchedAddRequest {
	tmdbId?: number;
	igdbId?: number;
	contentType: SupportedMedia;
	rating?: number;
	status?: WatchedStatus;
}

export interface WatchedUpdateRequest {
	rating?: number;
	status?: WatchedStatus;
	thoughts?: string;
	removeThoughts?: boolean;
	pinned?: boolean;
}

export interface WatchedUpdateResponse {
	newActivity: Activity;
}

export interface ActivityUpdateRequest {
	customDate: string;
}

export interface WatchedSeasonAddResponse {
	watchedSeasons: WatchedSeason[];
	addedActivity: Activity;
}

export interface WatchedEpisodeAddResponse {
	watchedEpisodes: WatchedEpisode[];
	addedActivity: Activity;
	episodeStatusChangedHookResponse?: EpisodeStatusChangedHookResponse;
}

export interface EpisodeStatusChangedHookResponse {
	newShowStatus?: WatchedStatus;
	watchedSeason?: WatchedSeason;
	addedActivities?: Activity[];
	errors?: string[];
}

export interface Profile {
	joined: Date;
	showsWatched: number;
	moviesWatched: number;
	moviesWatchedRuntime: number;
	showsWatchedRuntime: number;
}

export interface UserSettings {
	private: boolean;
	privateThoughts: boolean;
	hideSpoilers: boolean;
	includePreviouslyWatched: boolean;
	country: string;
	automateShowStatuses: boolean;
	ratingSystem?: RatingSystem;
	/**
	 * A rating step decided by the user, only
	 * applicable for OutOf10 and OutOf5 rating systems.
	 * Supported: 1, 0.5, 0.1 (must validate).
	 */
	ratingStep?: RatingStep;
}

export enum RatingSystem {
	OutOf10, // default
	OutOf100,
	OutOf5,
	Thumbs,
}

export enum RatingStep {
	One, // default
	Point5,
	Point1,
}

export interface ChangePasswordForm {
	currentPassword: string;
	newPassword: string;
	reEnteredNewPassword: string;
}

// What the user search returns
export interface PublicUser {
	id: number;
	username: string;
	avatar?: Image;
	bio?: string;
}

// PrivateUser - Current users info
export interface PrivateUser {
	username: string;
	type: UserType;
	permissions: UserPermission;
	avatar: Image;
	bio: string;
}

export enum UserPermission {
	PERM_NONE = 1,
	PERM_ADMIN = 2,
	PERM_REQUEST_CONTENT = 4,
	PERM_REQUEST_CONTENT_AUTO_APPROVE = 8,
}

export interface Image {
	createdAt: Date;
	blurHash: string;
	path: string;
}

export interface JellyfinFoundContent {
	hasContent: boolean;
	url: string;
}

export interface AvailableAuthProviders {
	available: string[];
	signupEnabled: boolean;
	isInSetup: boolean;
	useEmby: boolean;
	headerAuthAutoLogin: boolean;
}

export interface TokenClaims {
	userId: number;
	username: string;
	type: number;
}

export interface MediaIDs {
	tmdb?: number;
	imdb?: string;
	wikidata?: string;
	tvdb?: number;

	igdb?: number;
}

export enum MediaTypeE {
	tmdbMovie = "tmdb_movie",
	tmdbShow = "tmdb_tv",
	tmdbPerson = "tmdb_person",
	igdbGame = "igdb_game",
}

export interface Media {
	type?: MediaTypeE;
	ids: MediaIDs;
	name?: string;
	summary?: string;
	poster?: Image;
	extPosterPath?: string;
	rating?: number;
	ratingCount?: number;
	watched?: Watched;
	similar?: Media[];
	releaseDate?: string;
	extBackdropPath?: string;
	genres?: MediaGenre[];
	homepage?: string;
	videos?: MediaVideo[];
	runtime?: number;
	providers?: MediaProvider[];
	providersFullListLink?: string;
	gameModes?: MediaGenre[];
	seasons?: MediaSeason[];
	isShowAnime?: boolean;
}

export function getContentTypeFromMedia(m: Media): ContentType | undefined {
	switch (m.type) {
		case MediaTypeE.tmdbMovie:
			return "movie";
		case MediaTypeE.tmdbShow:
			return "tv";
		case MediaTypeE.igdbGame:
			return "game";
	}
	return;
}

export interface MediaGenre {
	// ID Is for external db id.
	id: number;
	// Genre name.
	name: string;
}

export interface MediaProvider {
	name: string;
	link: string;
}

export enum MediaVideoType {
	trailer = "trailer",
	other = "other",
}

export interface MediaVideo {
	id?: string;
	name?: string;
	type?: MediaVideoType;
	best?: boolean;
}

export interface MediaSeason {
	name?: string;
	number: number;
	episodeCount: number;
	releaseDate?: string;
}

interface PaginationParams {
	limit?: number;
	page?: number;
}

export enum SearchType {
	multi = "multi",
	movie = "movie",
	show = "show",
	person = "person",
	game = "game",
}

export interface SearchRequest extends PaginationParams {
	type?: SearchType;
	query: string;
}

export interface SearchResponseMeta {
	fromMyList?: boolean;
}

export enum DiscoverFilter {
	trending = "trending",
	popular = "popular",
	upcoming = "upcoming",
	streaming = "streaming",
	inTheatres = "intheatres",
}

export type DiscoverFilterOption = `${DiscoverFilter}`;

export interface DiscoverRequest extends PaginationParams {
	type?: SearchType;
	filter?: DiscoverFilter;
}

export interface PersonDetailsResponse {
	name?: string;
	birthday?: string;
	deathday?: string;
	age?: number;
	placeOfBirth?: string;
	knownForDepartment?: string;
	biography?: string;
	homepage?: string;
	extPosterPath?: string;
}

export interface PersonCreditsResponse {
	credits?: Media[];
}

export interface TMDBSeasonDetails {
	_id: string;
	air_date: string;
	episodes: TMDBSeasonDetailsEpisode[];
	name: string;
	overview: string;
	id: number;
	poster_path: string;
	season_number: number;
}

export interface TMDBSeasonDetailsEpisode {
	air_date: string;
	episode_number: number;
	id: number;
	name: string;
	overview: string;
	production_code: string;
	runtime: number;
	season_number: number;
	show_id: number;
	still_path: string;
	vote_average: number;
	vote_count: number;
	crew: {
		department: string;
		job: string;
		credit_id: string;
		adult: boolean;
		gender: number;
		id: number;
		known_for_department: string;
		name: string;
		original_name: string;
		popularity: number;
		profile_path: string;
	}[];
	guest_stars: {
		character: string;
		credit_id: string;
		order: number;
		adult: boolean;
		gender: number;
		id: number;
		known_for_department: string;
		name: string;
		original_name: string;
		popularity: number;
		profile_path: string;
	}[];
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

export interface TMDBRegions {
	results: {
		iso_3166_1: string;
		english_name: string;
		native_name: string;
	}[];
}

export enum ImportResponseType {
	IMPORT_SUCCESS = "IMPORT_SUCCESS",
	IMPORT_FAILED = "IMPORT_FAILED",
	IMPORT_MULTI = "IMPORT_MULTI",
	IMPORT_NOTFOUND = "IMPORT_NOTFOUND",
	IMPORT_EXISTS = "IMPORT_EXISTS",
}

export interface ImportResponse {
	type: ImportResponseType;
	results?: Media[];
	match?: Media;
	watchedEntry?: Watched;
}

export interface ImportedList {
	tmdbId?: number;
	name: string;
	year?: number;
	type?: ContentType;
	state?: string;
	rating?: number;
	ratingCustomDate?: Date;
	status?: WatchedStatus;
	thoughts?: string;
	datesWatched?: Date[];
	activity?: Activity[];
	watchedEpisodes?: WatchedEpisode[];
	watchedSeasons?: WatchedSeason[];
	tags?: TagAddRequest[];
	imdbId?: string;
}

export interface Filters {
	type: string[];
	status: string[];
}

export interface ManagedUser {
	id: number;
	createdAt: Date;
	username: string;
	type: UserType;
	permissions: number;
	private: boolean;
}

export interface ServerConfig {
	DEFAULT_COUNTRY: string;
	JELLYFIN_HOST: string;
	USE_EMBY: boolean;
	SIGNUP_ENABLED: boolean;
	TMDB_KEY: string;
	PLEX_HOST: string;
	PLEX_MACHINE_ID: string;
	SONARR: SonarrSettings[];
	RADARR: RadarrSettings[];
	TWITCH?: TwitchSettings;
	DEBUG: boolean;
}

export interface ServerConfigByName<T> {
	value: T;
}

export interface SonarrSettings {
	name: string;
	host?: string;
	key?: string;
	qualityProfile?: number;
	rootFolder?: number;
	languageProfile?: number;
	automaticSearch?: boolean;
}

export interface RadarrSettings {
	name: string;
	host?: string;
	key?: string;
	qualityProfile?: number;
	rootFolder?: number;
	automaticSearch?: boolean;
}

interface ArrSettingsPublicResponseBase {
	name: string;
	host?: string;
	qualityProfile?: number;
	rootFolder?: number;
	automaticSearch: boolean;
}

export interface SonarrSettingsPublicResponseResult
	extends ArrSettingsPublicResponseBase {
	languageProfile?: number;
}

export interface RadarrSettingsPublicResponseResult
	extends ArrSettingsPublicResponseBase {}

export interface TwitchSettings {
	clientId?: string;
	clientSecret?: string;
}

export interface TrustedHeaderAuthSetting {
	enabled: boolean;
	headerName: string;
	autoLogin: boolean;
	logoutUrl: string;
}

export interface TrustedHeaderAuthLogoutDetailsResponse {
	logoutUrl?: string;
}

export interface DropDownItem {
	id: number | string;
	value: string;
	icon?: Icon;
}

export interface ListBoxItem {
	id: number;
	value: boolean;
	displayValue: string;
}

export interface QualityProfile {
	name: string;
	upgradeAllowed: boolean;
	cutoff: number;
	items: {
		quality?: {
			id: number;
			name: string;
			source: string;
			resolution: number;
		};
		items: any[];
		allowed: boolean;
		name?: string;
		id?: number;
	}[];
	id: number;
}

export interface RootFolder {
	path: string;
	accessible: boolean;
	freeSpace: number;
	unmappedFolders: any[];
	id: number;
}

export interface LanguageProfile {
	name: string;
	upgradeAllowed: boolean;
	cutoff: {
		id: number;
		name: string;
	};
	languages: {
		language: {
			id: number;
			name: string;
		};
		allowed: boolean;
	}[];
	id: number;
}

export interface SonarrTestResponse {
	qualityProfiles: QualityProfile[];
	rootFolders: RootFolder[];
	languageProfiles: LanguageProfile[];
}

export interface RadarrTestResponse {
	qualityProfiles: QualityProfile[];
	rootFolders: RootFolder[];
}

export type ArrRequestStatus =
	| "PENDING"
	| "APPROVED"
	| "AUTO_APPROVED"
	| "DENIED"
	| "FOUND";

export interface ArrRequestResponse {
	id: number;
	createdAt: string;
	updatedAt: string;
	serverName: string;
	arrId: number;
	content: Content;
	status: ArrRequestStatus;
	requestJson: string;
	username: string;
}

export interface ArrDetailsResponse {
	progress: number;
	estimatedCompletionTime: string;
	status: string;
	trackedDownloadStatus: string;
	trackedDownloadState: string;
}

export interface ArrInfoResponse {
	hasFile: boolean;
	isAvailable: boolean;
	added: string;
}

export interface ServerFeatures {
	sonarr: boolean;
	radarr: boolean;
	games: boolean;
}

export interface Follow {
	createdAt: Date;
	followedUser: PublicUser;
}

interface MovaryExportBase {
	title: string;
	year: string;
	tmdbId: string;
	imdbId: string;
}

export interface MovaryHistory extends MovaryExportBase {
	watchedAt: string;
	comment: string;
}

export interface MovaryRatings extends MovaryExportBase {
	userRating: string;
}

export interface MovaryWatchlist extends MovaryExportBase {
	addedAt: string;
}

export interface TodoMoviesExport {
	Movie: TodoMoviesMovie[];
	MovieList: TodoMoviesCustomList[];
}

export interface TodoMoviesMovie {
	Attrs: {
		tmdbID: number;
		title: string;
		isWatched: number;
		insertionDate: {
			Value: number;
			Class: string;
		};
		myScore: number;
	};
	Rels: {
		lists: {
			Items: string[];
			Entity: string;
		};
	};
	ObjectID: string;
}

export interface TodoMoviesCustomList {
	Attrs: {
		colorInHex: string;
		order: number;
		iconFileName: string;
		featuredListID: number;
		name: string;
	};
	Rels: {
		movies: {
			Items: string[];
			Entity: string;
		};
	};
	ObjectID: string;
}

// General interface for all requests that return a job that was started.
export interface JobCreatedResponse {
	jobId: string;
}

export enum JobStatus {
	CREATED = "CREATED",
	RUNNING = "RUNNING",
	DONE = "DONE",
	CANCELLED = "CANCELLED",
}

export interface GetJobResponse {
	name: string;
	status: JobStatus;
	currentTask?: string;
	errors: string[];
}

export interface TaskRescheduleRequest {
	seconds: number;
}

export interface AllTasksResponse {
	name: string;
	nextRun: Date;
	seconds: number;
}

export interface Tag extends dbModel {
	name: string;
	color: string;
	bgColor: string;
}

export interface TagAddRequest {
	name: string;
	color: string;
	bgColor: string;
}
