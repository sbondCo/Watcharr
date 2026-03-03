package tmdb

import (
	"log/slog"
	"strings"
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/util"
)

// Separated from `TMDBSearchResponse` so we can embed it for
// easily assigning all page fields in one.
type TMDBPageFields struct {
	Page         int `json:"page"`
	TotalPages   int `json:"total_pages"`
	TotalResults int `json:"total_results"`
}

type TMDBSearchResponse[R any] struct {
	TMDBPageFields
	Results []R `json:"results"`
}

// A common "base" type for search results.
// Some properties are used commonly for all types except Person, but
// are still embedded in person for ease of use right now.
type TMDBSearchResult struct {
	// TMDB ID
	ID int `json:"id"`
	// Media Type (movie, tv, person)
	// **Some requests won't return this value
	// (namely any request other than a multi
	// type search), but we add it in manually.**
	MediaType string `json:"media_type"`
	// Summary (only for movie/tv)
	Overview string `json:"overview"`
	// Poster path (only for movie/tv)
	PosterPath string `json:"poster_path"`
	// Rating (only for movie/tv)
	VoteAverage float32 `json:"vote_average"`
	// Amount of votes for rating (only for movie/tv)
	VoteCount uint32 `json:"vote_count"`
}

// Adds the base items to a Media struct, which can be used in the
// structs that embed TMDBSearchResult to simplify and reduce duplication.
func (t *TMDBSearchResult) AsMedia() domain.Media {
	m := domain.Media{
		IDs: domain.MediaIDs{
			TMDB: t.ID,
		},
		Summary:       t.Overview,
		ExtPosterPath: t.PosterPath,
		Rating:        uint(t.VoteAverage * 10),
		RatingCount:   uint(t.VoteCount),
	}
	switch t.MediaType {
	case "movie":
		m.Type = domain.MediaTypeTMDBMovie
	case "tv":
		m.Type = domain.MediaTypeTMDBShow
	case "person":
		m.Type = domain.MediaTypeTMDBPerson
	}
	return m
}

//
// Multi Search
//

type TMDBSearchMultiResponse struct {
	TMDBSearchResponse[TMDBSearchMultiResult]
}

type TMDBSearchMultiResult struct {
	TMDBSearchResult
	Adult            bool     `json:"adult"`
	BackdropPath     string   `json:"backdrop_path"`
	Title            string   `json:"title,omitempty"`
	OriginalLanguage string   `json:"original_language"`
	OriginalTitle    string   `json:"original_title,omitempty"`
	ProfilePath      string   `json:"profile_path"`
	GenreIds         []int64  `json:"genre_ids"`
	Popularity       float32  `json:"popularity"`
	ReleaseDate      string   `json:"release_date,omitempty"`
	Video            bool     `json:"video,omitempty"`
	Name             string   `json:"name,omitempty"`
	OriginalName     string   `json:"original_name,omitempty"`
	FirstAirDate     string   `json:"first_air_date,omitempty"`
	OriginCountry    []string `json:"origin_country,omitempty"`
	// Below are for tv episode results
	AirDate        string `json:"air_date,omitempty"`
	EpisodeNumber  int    `json:"episode_number,omitempty"`
	EpisodeType    string `json:"episode_type,omitempty"`
	ProductionCode string `json:"production_code,omitempty"`
	Runtime        int    `json:"runtime,omitempty"`
	SeasonNumber   int    `json:"season_number,omitempty"`
	ShowId         int    `json:"show_id,omitempty"`
	StillPath      string `json:"still_path,omitempty"`
}

func (t *TMDBSearchMultiResult) AsMedia() domain.Media {
	m := t.TMDBSearchResult.AsMedia()

	m.Name = t.Title
	if t.Name != "" {
		m.Name = t.Name
	}

	var tmdbReleaseDate string
	switch t.MediaType {
	case "movie":
		tmdbReleaseDate = t.ReleaseDate
	case "tv":
		tmdbReleaseDate = t.FirstAirDate
	case "person":
		m.ExtPosterPath = t.ProfilePath
	}
	if releaseDate, err := time.Parse("2006-01-02", tmdbReleaseDate); err == nil {
		m.ReleaseDate = releaseDate
	} else {
		slog.Error("AsMedia: Failed to parse release date", "name", m.Name, "error", err)
	}
	return m
}

type TMDBSearchMultiResponseWithWatched struct {
	TMDBSearchResponse[TMDBSearchMultiResultWithWatched]
}

type TMDBSearchMultiResultWithWatched struct {
	TMDBSearchMultiResult
	Watched *entity.Watched `json:"watched,omitempty"`
}

//
// Movie Search
//

type TMDBSearchMoviesResponse struct {
	TMDBSearchResponse[TMDBSearchMovieResult]
}

type TMDBSearchMovieResult struct {
	TMDBSearchResult
	Adult            bool    `json:"adult"`
	BackdropPath     string  `json:"backdrop_path"`
	GenreIds         []int   `json:"genre_ids"`
	OriginalLanguage string  `json:"original_language"`
	OriginalTitle    string  `json:"original_title"`
	Popularity       float64 `json:"popularity"`
	ReleaseDate      string  `json:"release_date"`
	Title            string  `json:"title"`
	Video            bool    `json:"video"`
}

func (t *TMDBSearchMovieResult) AsMedia() domain.Media {
	m := t.TMDBSearchResult.AsMedia()
	m.Name = t.Title
	if releaseDate, err := time.Parse("2006-01-02", t.ReleaseDate); err == nil {
		m.ReleaseDate = releaseDate
	} else {
		slog.Error("AsMedia: Failed to parse release date", "name", m.Name, "error", err)
	}
	return m
}

type TMDBSearchMoviesResponseWithWatched struct {
	TMDBSearchResponse[TMDBSearchMovieResultWithWatched]
}

type TMDBSearchMovieResultWithWatched struct {
	TMDBSearchMovieResult
	Watched *entity.Watched `json:"watched,omitempty"`
}

//
// Tv Shows Search
//

type TMDBSearchShowsResponse struct {
	TMDBSearchResponse[TMDBSearchShowsResult]
}

type TMDBSearchShowsResult struct {
	TMDBSearchResult
	Adult            bool     `json:"adult"`
	BackdropPath     string   `json:"backdrop_path"`
	GenreIds         []int    `json:"genre_ids"`
	OriginCountry    []string `json:"origin_country"`
	OriginalLanguage string   `json:"original_language"`
	OriginalName     string   `json:"original_name"`
	Popularity       float64  `json:"popularity"`
	FirstAirDate     string   `json:"first_air_date"`
	Name             string   `json:"name"`
}

func (t *TMDBSearchShowsResult) AsMedia() domain.Media {
	m := t.TMDBSearchResult.AsMedia()
	m.Name = t.Name
	if releaseDate, err := time.Parse("2006-01-02", t.FirstAirDate); err == nil {
		m.ReleaseDate = releaseDate
	} else {
		slog.Error("AsMedia: Failed to parse release date", "name", m.Name, "error", err)
	}
	return m
}

type TMDBSearchShowsResponseWithWatched struct {
	TMDBSearchResponse[TMDBSearchShowsResultWithWatched]
}

type TMDBSearchShowsResultWithWatched struct {
	TMDBSearchShowsResult
	Watched *entity.Watched `json:"watched,omitempty"`
}

//
// People Search
//

type TMDBSearchPeopleResult struct {
	TMDBSearchResult
	Adult              bool    `json:"adult"`
	Gender             int     `json:"gender"`
	KnownForDepartment string  `json:"known_for_department"`
	Name               string  `json:"name"`
	OriginalName       string  `json:"original_name"`
	Popularity         float64 `json:"popularity"`
	ProfilePath        string  `json:"profile_path"`
	KnownFor           []struct {
		Adult            bool    `json:"adult"`
		BackdropPath     string  `json:"backdrop_path"`
		ID               int     `json:"id"`
		Title            string  `json:"title"`
		OriginalLanguage string  `json:"original_language"`
		OriginalTitle    string  `json:"original_title"`
		Overview         string  `json:"overview"`
		PosterPath       string  `json:"poster_path"`
		MediaType        string  `json:"media_type"`
		GenreIds         []int   `json:"genre_ids"`
		Popularity       float64 `json:"popularity"`
		ReleaseDate      string  `json:"release_date"`
		Video            bool    `json:"video"`
		VoteAverage      float64 `json:"vote_average"`
		VoteCount        int     `json:"vote_count"`
	} `json:"known_for"`
}

func (t *TMDBSearchPeopleResult) AsMedia() domain.Media {
	m := t.TMDBSearchResult.AsMedia()
	m.Name = t.Name
	m.ExtPosterPath = t.ProfilePath
	return m
}

type TMDBSearchPeopleResponse struct {
	TMDBSearchResponse[TMDBSearchPeopleResult]
}

//
// Search By External ID
//

type TMDBFindByExternalIdResponse struct {
	// These are all a TMDBSearchMultiResult so our search func can easily
	// combine all of them into one []TMDBSearchMultiResult for response
	// to client (seems not easy to convert to TMDBSearchMultiResult for
	// concatenation after unmarshalling to correct type).
	MovieResults     []TMDBSearchMultiResult `json:"movie_results"`
	PersonResults    []TMDBSearchMultiResult `json:"person_results"`
	TvResults        []TMDBSearchMultiResult `json:"tv_results"`
	TvSeasonResults  []TMDBSearchMultiResult `json:"tv_season_results"`
	TvEpisodeResults []TMDBSearchMultiResult `json:"tv_episode_results"`
}

//
// Content Details
// A base for details structs.
//

type TMDBContentDetails struct {
	ID           int    `json:"id"`
	PosterPath   string `json:"poster_path"`
	BackdropPath string `json:"backdrop_path"`
	Genres       []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"genres"`
	Homepage            string  `json:"homepage"`
	Popularity          float32 `json:"popularity"`
	Overview            string  `json:"overview"`
	OriginalLanguage    string  `json:"original_language"`
	ProductionCompanies []struct {
		ID            int    `json:"id"`
		LogoPath      string `json:"logo_path"`
		Name          string `json:"name"`
		OriginCountry string `json:"origin_country"`
	} `json:"production_companies"`
	ProductionCountries []struct {
		Iso31661 string `json:"iso_3166_1"`
		Name     string `json:"name"`
	} `json:"production_countries"`
	Status          string  `json:"status"`
	Tagline         string  `json:"tagline"`
	VoteAverage     float32 `json:"vote_average"`
	VoteCount       uint32  `json:"vote_count"`
	SpokenLanguages []struct {
		EnglishName string `json:"english_name"`
		Iso6391     string `json:"iso_639_1"`
		Name        string `json:"name"`
	} `json:"spoken_languages"`

	// Extra items because we use `append_to_response` on the request
	Videos TMDBContentVideos `json:"videos"`
	// Raw watched providers object from tmdb
	WatchProviders interface{} `json:"watch/providers"`
	// Watched providers but after we apply our transformations to it.
	WatchProvidersTransformed WatchProviders `json:"-"`
}

// Adds the base items to a Media struct, which can be used in the
// structs that embed TMDBSearchResult to simplify and reduce duplication.
func (t *TMDBContentDetails) AsMedia() domain.Media {
	m := domain.Media{
		IDs: domain.MediaIDs{
			TMDB: t.ID,
		},
		Summary:         t.Overview,
		ExtPosterPath:   t.PosterPath,
		ExtBackdropPath: t.BackdropPath,
		Rating:          uint(t.VoteAverage * 10),
		RatingCount:     uint(t.VoteCount),
		Homepage:        t.Homepage,
	}
	// Genres
	for _, g := range t.Genres {
		m.Genres = append(m.Genres, domain.MediaGenre{
			ID:   uint(g.ID),
			Name: g.Name,
		})
	}
	// Videos
	for i := range t.Videos.Results {
		v := &t.Videos.Results[i]
		// Currently we only care about trailers
		if strings.ToLower(v.Type) != "trailer" {
			continue
		}
		// Is best?
		isBest := false
		if v.Official && strings.ToLower(v.Name) == "official trailer" {
			isBest = true
		}
		m.Videos = append(m.Videos, domain.MediaVideo{
			ID:   v.Key,
			Name: v.Name,
			// Currently we only care about trailers
			Type: domain.MediaVideoTypeTrailer,
			Best: isBest,
		})
	}
	// Watch providers
	for _, v := range t.WatchProvidersTransformed.Free {
		m.Providers = append(m.Providers, domain.MediaProvider{
			Name: v.ProviderName,
			Type: domain.MediaProviderTypeFree,
		})
	}
	for _, v := range t.WatchProvidersTransformed.Flatrate {
		m.Providers = append(m.Providers, domain.MediaProvider{
			Name: v.ProviderName,
			Type: domain.MediaProviderTypeSub,
		})
	}
	m.ProvidersFullListLink = t.WatchProvidersTransformed.Link
	return m
}

//
// Movie Details
//

type TMDBMovieDetails struct {
	TMDBContentDetails
	Adult               bool   `json:"adult"`
	BelongsToCollection any    `json:"belongs_to_collection"`
	Budget              uint32 `json:"budget"`
	ImdbID              string `json:"imdb_id"`
	OriginalTitle       string `json:"original_title"`
	ReleaseDate         string `json:"release_date"`
	Revenue             uint32 `json:"revenue"`
	Runtime             uint32 `json:"runtime"`
	Title               string `json:"title"`
	Video               bool   `json:"video"`

	// Extra items because we use `append_to_response` on the request
	ExternalIds TMDBExternalIdsMovie `json:"external_ids"`
	Similar     TMDBMovieSimilar     `json:"similar"`
}

func (t *TMDBMovieDetails) AsMedia() domain.Media {
	m := t.TMDBContentDetails.AsMedia()
	m.Type = domain.MediaTypeTMDBMovie
	m.Name = t.Title
	m.Runtime = uint(t.Runtime)
	if releaseDate, err := time.Parse("2006-01-02", t.ReleaseDate); err == nil {
		m.ReleaseDate = releaseDate
	} else {
		slog.Error("AsMedia: Failed to parse release date", "name", m.Name, "error", err)
	}
	// IDS
	m.IDs.IMDB = t.ExternalIds.ImdbID
	m.IDs.Wikidata = t.ExternalIds.WikidataID
	// Convert similar items to media too.
	for i := range t.Similar.Results {
		m.Similar = append(m.Similar, t.Similar.Results[i].AsMedia())
	}
	return m
}

//
// Movie Details Similar
//

type TMDBMovieSimilar struct {
	TMDBSearchResponse[TMDBMovieSimilarResult]
}

type TMDBMovieSimilarResult struct {
	ID               int     `json:"id"`
	Title            string  `json:"title"`
	Adult            bool    `json:"adult"`
	BackdropPath     string  `json:"backdrop_path"`
	GenreIds         []int   `json:"genre_ids"`
	OriginalLanguage string  `json:"original_language"`
	OriginalTitle    string  `json:"original_title"`
	Overview         string  `json:"overview"`
	Popularity       float64 `json:"popularity"`
	PosterPath       string  `json:"poster_path"`
	ReleaseDate      string  `json:"release_date"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        uint32  `json:"vote_count"`
}

func (t *TMDBMovieSimilarResult) AsMedia() domain.Media {
	m := domain.Media{
		IDs: domain.MediaIDs{
			TMDB: t.ID,
		},
		Type:          domain.MediaTypeTMDBMovie,
		Name:          t.Title,
		Summary:       t.Overview,
		ExtPosterPath: t.PosterPath,
		Rating:        uint(t.VoteAverage * 10),
		RatingCount:   uint(t.VoteCount),
	}
	if releaseDate, err := time.Parse("2006-01-02", t.ReleaseDate); err == nil {
		m.ReleaseDate = releaseDate
	} else {
		slog.Error("AsMedia: Failed to parse release date", "name", m.Name, "error", err)
	}
	return m
}

//
// Show Details
//

type TMDBShowDetails struct {
	TMDBContentDetails
	CreatedBy []struct {
		ID          int    `json:"id"`
		CreditID    string `json:"credit_id"`
		Name        string `json:"name"`
		Gender      int    `json:"gender"`
		ProfilePath string `json:"profile_path"`
	} `json:"created_by"`
	EpisodeRunTime   []int    `json:"episode_run_time"`
	FirstAirDate     string   `json:"first_air_date"`
	InProduction     bool     `json:"in_production"`
	Languages        []string `json:"languages"`
	LastAirDate      string   `json:"last_air_date"`
	LastEpisodeToAir struct {
		AirDate        string  `json:"air_date"`
		EpisodeNumber  int     `json:"episode_number"`
		ID             int     `json:"id"`
		Name           string  `json:"name"`
		Overview       string  `json:"overview"`
		ProductionCode string  `json:"production_code"`
		SeasonNumber   int     `json:"season_number"`
		StillPath      string  `json:"still_path"`
		VoteAverage    float64 `json:"vote_average"`
		VoteCount      uint32  `json:"vote_count"`
	} `json:"last_episode_to_air"`
	Name             string `json:"name"`
	NextEpisodeToAir any    `json:"next_episode_to_air"`
	Networks         []struct {
		Name          string `json:"name"`
		ID            int    `json:"id"`
		LogoPath      string `json:"logo_path"`
		OriginCountry string `json:"origin_country"`
	} `json:"networks"`
	NumberOfEpisodes uint32   `json:"number_of_episodes"`
	NumberOfSeasons  uint32   `json:"number_of_seasons"`
	OriginCountry    []string `json:"origin_country"`
	OriginalName     string   `json:"original_name"`
	Seasons          []struct {
		AirDate      string `json:"air_date"`
		EpisodeCount int    `json:"episode_count"`
		ID           int    `json:"id"`
		Name         string `json:"name"`
		Overview     string `json:"overview"`
		PosterPath   string `json:"poster_path"`
		SeasonNumber int    `json:"season_number"`
	} `json:"seasons"`
	Type string `json:"type"`

	// Extra items because we use `append_to_response` on the request
	ExternalIds TMDBExternalIdsShow `json:"external_ids"`
	Keywords    TMDBKeywords        `json:"keywords"`
	Similar     TMDBShowSimilar     `json:"similar"`
}

func (t *TMDBShowDetails) AsMedia() domain.Media {
	m := t.TMDBContentDetails.AsMedia()
	m.Type = domain.MediaTypeTMDBShow
	m.Name = t.Name
	if releaseDate, err := time.Parse("2006-01-02", t.FirstAirDate); err == nil {
		m.ReleaseDate = releaseDate
	} else {
		slog.Error("AsMedia: Failed to parse release date", "name", m.Name, "error", err)
	}
	// IDS
	m.IDs.IMDB = t.ExternalIds.ImdbID
	m.IDs.Wikidata = t.ExternalIds.WikidataID
	m.IDs.TVDB = t.ExternalIds.TvdbID
	// Convert similar items to media too.
	for i := range t.Similar.Results {
		m.Similar = append(m.Similar, t.Similar.Results[i].AsMedia())
	}
	// Seasons
	for _, v := range t.Seasons {
		ms := domain.MediaSeason{
			Number:       v.SeasonNumber,
			Name:         v.Name,
			EpisodeCount: v.EpisodeCount,
		}
		if releaseDate, err := time.Parse("2006-01-02", v.AirDate); err == nil {
			ms.ReleaseDate = releaseDate
		} else {
			slog.Error("AsMedia: Failed to parse release date", "name", m.Name, "error", err)
		}
		m.Seasons = append(m.Seasons, ms)
	}
	// Is show anime
	for _, v := range t.Keywords.Results {
		// 210024 is the id of "anime" keyword on tmdb.
		if v.ID == 210024 {
			m.IsShowAnime = true
			break
		}
	}
	return m
}

type TMDBSeasonDetails struct {
	ID       string `json:"_id"`
	AirDate  string `json:"air_date"`
	Episodes []struct {
		AirDate        string  `json:"air_date"`
		EpisodeNumber  int     `json:"episode_number"`
		ID             int     `json:"id"`
		Name           string  `json:"name"`
		Overview       string  `json:"overview"`
		ProductionCode string  `json:"production_code"`
		Runtime        int     `json:"runtime"`
		SeasonNumber   int     `json:"season_number"`
		ShowID         int     `json:"show_id"`
		StillPath      string  `json:"still_path"`
		VoteAverage    float64 `json:"vote_average"`
		VoteCount      int     `json:"vote_count"`
		Crew           []struct {
			Department         string  `json:"department"`
			Job                string  `json:"job"`
			CreditID           string  `json:"credit_id"`
			Adult              bool    `json:"adult"`
			Gender             int     `json:"gender"`
			ID                 int     `json:"id"`
			KnownForDepartment string  `json:"known_for_department"`
			Name               string  `json:"name"`
			OriginalName       string  `json:"original_name"`
			Popularity         float64 `json:"popularity"`
			ProfilePath        string  `json:"profile_path"`
		} `json:"crew"`
		GuestStars []struct {
			Character          string  `json:"character"`
			CreditID           string  `json:"credit_id"`
			Order              int     `json:"order"`
			Adult              bool    `json:"adult"`
			Gender             int     `json:"gender"`
			ID                 int     `json:"id"`
			KnownForDepartment string  `json:"known_for_department"`
			Name               string  `json:"name"`
			OriginalName       string  `json:"original_name"`
			Popularity         float64 `json:"popularity"`
			ProfilePath        string  `json:"profile_path"`
		} `json:"guest_stars"`
	} `json:"episodes"`
	Name         string `json:"name"`
	Overview     string `json:"overview"`
	ID0          int    `json:"id"`
	PosterPath   string `json:"poster_path"`
	SeasonNumber int    `json:"season_number"`
}

//
// Show Details Similar
//

type TMDBShowSimilar struct {
	TMDBSearchResponse[TMDBShowSimilarResult]
}

type TMDBShowSimilarResult struct {
	ID               int      `json:"id"`
	Name             string   `json:"name"`
	Adult            bool     `json:"adult"`
	BackdropPath     string   `json:"backdrop_path"`
	GenreIds         []int    `json:"genre_ids"`
	OriginCountry    []string `json:"origin_country"`
	OriginalLanguage string   `json:"original_language"`
	OriginalName     string   `json:"original_name"`
	Overview         string   `json:"overview"`
	Popularity       float64  `json:"popularity"`
	PosterPath       string   `json:"poster_path"`
	FirstAirDate     string   `json:"first_air_date"`
	VoteAverage      float64  `json:"vote_average"`
	VoteCount        uint32   `json:"vote_count"`
}

func (t *TMDBShowSimilarResult) AsMedia() domain.Media {
	m := domain.Media{
		IDs: domain.MediaIDs{
			TMDB: t.ID,
		},
		Type:          domain.MediaTypeTMDBShow,
		Name:          t.Name,
		Summary:       t.Overview,
		ExtPosterPath: t.PosterPath,
		Rating:        uint(t.VoteAverage * 10),
		RatingCount:   uint(t.VoteCount),
	}
	if releaseDate, err := time.Parse("2006-01-02", t.FirstAirDate); err == nil {
		m.ReleaseDate = releaseDate
	} else {
		slog.Error("AsMedia: Failed to parse release date", "name", m.Name, "error", err)
	}
	return m
}

//
// Person Details
//

type TMDBPersonDetails struct {
	ID                 int      `json:"id"`
	Name               string   `json:"name"`
	Birthday           string   `json:"birthday"`
	Deathday           string   `json:"deathday"`
	PlaceOfBirth       string   `json:"place_of_birth"`
	KnownForDepartment string   `json:"known_for_department"`
	Biography          string   `json:"biography"`
	AlsoKnownAs        []string `json:"also_known_as"`
	Popularity         float32  `json:"popularity"`
	ProfilePath        string   `json:"profile_path"`
	ImdbID             string   `json:"imdb_id"`
	Homepage           string   `json:"homepage"`
}

func (t *TMDBPersonDetails) AsPersonDetailsResponse() domain.PersonDetailsResponse {
	m := domain.PersonDetailsResponse{
		Name:               t.Name,
		PlaceOfBirth:       t.PlaceOfBirth,
		KnownForDepartment: t.KnownForDepartment,
		Biography:          t.Biography,
		ExtPosterPath:      t.ProfilePath,
		Homepage:           t.Homepage,
	}
	// Dates
	if date, err := time.Parse("2006-01-02", t.Birthday); err == nil {
		m.Birthday = date
	} else {
		slog.Error("AsPersonDetailsResponse: Failed to parse birthdate", "name", m.Name, "error", err)
	}
	if date, err := time.Parse("2006-01-02", t.Deathday); err == nil {
		m.Deathday = date
	} else {
		slog.Error("AsPersonDetailsResponse: Failed to parse birthdate", "name", m.Name, "error", err)
	}
	// Age
	m.Age = util.GetAge(m.Birthday, m.Deathday)
	return m
}

//
// Person Combined Credits
//

type TMDBPersonCombinedCredits struct {
	ID   int                                   `json:"id"`
	Cast []TMDBPersonCombinedCreditsCastResult `json:"cast"`
	// crew TMDBPersonCombinedCreditsCrew
}

type TMDBPersonCombinedCreditsCastResult struct {
	ID               int      `json:"id"`
	OriginalLanguage string   `json:"original_language"`
	EpisodeCount     int      `json:"episode_count"`
	Overview         string   `json:"overview"`
	OriginCountry    []string `json:"origin_country"`
	OriginalName     string   `json:"original_name"`
	GenreIDs         []int    `json:"genre_ids"`
	Name             string   `json:"name"`
	MediaType        string   `json:"media_type"`
	PosterPath       string   `json:"poster_path"`
	FirstAirDate     string   `json:"first_air_date"`
	VoteAverage      float64  `json:"vote_average"`
	VoteCount        uint32   `json:"vote_count"`
	Character        string   `json:"character"`
	BackdropPath     string   `json:"backdrop_path"`
	Popularity       float64  `json:"popularity"`
	CreditID         string   `json:"credit_id"`
	OriginalTitle    string   `json:"original_title"`
	Video            bool     `json:"video"`
	ReleaseDate      string   `json:"release_date"`
	Title            string   `json:"title"`
	Adult            bool     `json:"adult"`
}

func (t *TMDBPersonCombinedCreditsCastResult) AsMedia() domain.Media {
	m := domain.Media{
		IDs: domain.MediaIDs{
			TMDB: t.ID,
		},
		Summary:         t.Overview,
		ExtPosterPath:   t.PosterPath,
		ExtBackdropPath: t.BackdropPath,
		Rating:          uint(t.VoteAverage * 10),
		RatingCount:     uint(t.VoteCount),
	}

	m.Name = t.Title
	if t.Name != "" {
		m.Name = t.Name
	}

	var tmdbReleaseDate string
	switch t.MediaType {
	case "movie":
		m.Type = domain.MediaTypeTMDBMovie
		tmdbReleaseDate = t.ReleaseDate
	case "tv":
		m.Type = domain.MediaTypeTMDBShow
		tmdbReleaseDate = t.FirstAirDate
	}
	if releaseDate, err := time.Parse("2006-01-02", tmdbReleaseDate); err == nil {
		m.ReleaseDate = releaseDate
	} else {
		slog.Error("AsMedia: Failed to parse release date", "name", m.Name, "error", err)
	}
	return m
}

//
// Content Credits
//

type TMDBContentCredits struct {
	ID   int `json:"id"`
	Cast []struct {
		Adult              bool    `json:"adult"`
		Gender             int     `json:"gender"`
		ID                 int     `json:"id"`
		KnownForDepartment string  `json:"known_for_department"`
		Name               string  `json:"name"`
		OriginalName       string  `json:"original_name"`
		Popularity         float64 `json:"popularity"`
		ProfilePath        string  `json:"profile_path"`
		CastID             int     `json:"cast_id"`
		Character          string  `json:"character"`
		CreditID           string  `json:"credit_id"`
		Order              int     `json:"order"`
	} `json:"cast"`
	Crew []struct {
		Adult              bool    `json:"adult"`
		Gender             int     `json:"gender"`
		ID                 int     `json:"id"`
		KnownForDepartment string  `json:"known_for_department"`
		Name               string  `json:"name"`
		OriginalName       string  `json:"original_name"`
		Popularity         float64 `json:"popularity"`
		ProfilePath        string  `json:"profile_path"`
		CreditID           string  `json:"credit_id"`
		Department         string  `json:"department"`
		Job                string  `json:"job"`
	} `json:"crew"`
}

//
// Discover All Trending
//

type TrendingType string

const (
	TrendingTypeAll    TrendingType = "all"
	TrendingTypeMovie  TrendingType = "movie"
	TrendingTypeShow   TrendingType = "tv"
	TrendingTypePerson TrendingType = "person"
)

type TMDBTrendingCombined struct {
	TMDBSearchResponse[TMDBTrendingCombinedResult]
}

type TMDBTrendingCombinedResult struct {
	TMDBSearchResult
	Adult            bool     `json:"adult"`
	BackdropPath     string   `json:"backdrop_path"`
	Title            string   `json:"title,omitempty"`
	Name             string   `json:"name,omitempty"`
	OriginalLanguage string   `json:"original_language"`
	OriginalTitle    string   `json:"original_title,omitempty"`
	GenreIds         []int    `json:"genre_ids"`
	Popularity       float64  `json:"popularity"`
	ReleaseDate      string   `json:"release_date,omitempty"`
	Video            bool     `json:"video,omitempty"`
	OriginalName     string   `json:"original_name,omitempty"`
	FirstAirDate     string   `json:"first_air_date,omitempty"`
	OriginCountry    []string `json:"origin_country,omitempty"`
	ProfilePath      string   `json:"profile_path"`
}

func (t *TMDBTrendingCombinedResult) AsMedia() domain.Media {
	m := t.TMDBSearchResult.AsMedia()

	m.Name = t.Title
	if t.Name != "" {
		m.Name = t.Name
	}

	var tmdbReleaseDate string
	switch t.MediaType {
	case "movie":
		tmdbReleaseDate = t.ReleaseDate
	case "tv":
		tmdbReleaseDate = t.FirstAirDate
	case "person":
		m.ExtPosterPath = t.ProfilePath
	}
	if tmdbReleaseDate != "" {
		if releaseDate, err := time.Parse("2006-01-02", tmdbReleaseDate); err == nil {
			m.ReleaseDate = releaseDate
		} else {
			slog.Error("AsMedia: Failed to parse release date", "name", m.Name, "error", err)
		}
	}
	return m
}

//
// Discover
//

type DiscoverOptions struct {
	// Release date greater than.
	ReleaseDateMin time.Time
	// Release date less than.
	ReleaseDateMax time.Time
	// With release type.
	// Release types are listed on this page:
	// https://developer.themoviedb.org/reference/movie-release-dates
	WithReleaseType string
}

//
// Discover Movies
//

type TMDBDiscoverMovies struct {
	TMDBSearchResponse[TMDBDiscoverMoviesResult]
}

type TMDBDiscoverMoviesResult struct {
	Adult            bool    `json:"adult"`
	BackdropPath     string  `json:"backdrop_path"`
	GenreIds         []int   `json:"genre_ids"`
	ID               int     `json:"id"`
	OriginalLanguage string  `json:"original_language"`
	OriginalTitle    string  `json:"original_title"`
	Overview         string  `json:"overview"`
	Popularity       float64 `json:"popularity"`
	PosterPath       string  `json:"poster_path"`
	ReleaseDate      string  `json:"release_date"`
	Title            string  `json:"title"`
	Video            bool    `json:"video"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
}

func (t *TMDBDiscoverMoviesResult) AsMedia() domain.Media {
	m := domain.Media{
		IDs: domain.MediaIDs{
			TMDB: t.ID,
		},
		Type:          domain.MediaTypeTMDBMovie,
		Name:          t.Title,
		Summary:       t.Overview,
		ExtPosterPath: t.PosterPath,
		Rating:        uint(t.VoteAverage * 10),
		RatingCount:   uint(t.VoteCount),
	}
	if releaseDate, err := time.Parse("2006-01-02", t.ReleaseDate); err == nil {
		m.ReleaseDate = releaseDate
	} else {
		slog.Error("AsMedia: Failed to parse release date", "name", m.Name, "error", err)
	}
	return m
}

//
// Discover Shows
//

type TMDBDiscoverShows struct {
	TMDBSearchResponse[TMDBDiscoverShowsResult]
}

type TMDBDiscoverShowsResult struct {
	BackdropPath     string   `json:"backdrop_path"`
	FirstAirDate     string   `json:"first_air_date"`
	GenreIds         []int    `json:"genre_ids"`
	ID               int      `json:"id"`
	Name             string   `json:"name"`
	OriginCountry    []string `json:"origin_country"`
	OriginalLanguage string   `json:"original_language"`
	OriginalName     string   `json:"original_name"`
	Overview         string   `json:"overview"`
	Popularity       float64  `json:"popularity"`
	PosterPath       string   `json:"poster_path"`
	VoteAverage      float32  `json:"vote_average"`
	VoteCount        int      `json:"vote_count"`
}

func (t *TMDBDiscoverShowsResult) AsMedia() domain.Media {
	m := domain.Media{
		IDs: domain.MediaIDs{
			TMDB: t.ID,
		},
		Type:          domain.MediaTypeTMDBShow,
		Name:          t.Name,
		Summary:       t.Overview,
		ExtPosterPath: t.PosterPath,
		Rating:        uint(t.VoteAverage * 10),
		RatingCount:   uint(t.VoteCount),
	}
	if releaseDate, err := time.Parse("2006-01-02", t.FirstAirDate); err == nil {
		m.ReleaseDate = releaseDate
	} else {
		slog.Error("AsMedia: Failed to parse release date", "name", m.Name, "error", err)
	}
	return m
}

//
// Discover Shows
//

type TMDBPopularPeople struct {
	TMDBSearchResponse[TMDBPopularPeopleResult]
}

type TMDBPopularPeopleResult struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	ProfilePath string `json:"profile_path"`
}

func (t *TMDBPopularPeopleResult) AsMedia() domain.Media {
	m := domain.Media{
		IDs: domain.MediaIDs{
			TMDB: t.ID,
		},
		Type:          domain.MediaTypeTMDBPerson,
		Name:          t.Name,
		ExtPosterPath: t.ProfilePath,
	}
	return m
}

//
//
//

// Watch providers in one country
type WatchProviders struct {
	// Subscription services
	Flatrate []WatchProvider `json:"flatrate"`
	// Free providers
	Free []WatchProvider `json:"free"`
	// Link to view all streaming options on tmdb
	Link string `json:"link"`
}

type WatchProvider struct {
	ProviderID      int    `json:"provider_id"`
	ProviderName    string `json:"provider_name"`
	DisplayPriority int    `json:"display_priority"`
}

type TMDBContentVideos struct {
	ID      int `json:"id"`
	Results []struct {
		Iso6391     string    `json:"iso_639_1"`
		Iso31661    string    `json:"iso_3166_1"`
		Name        string    `json:"name"`
		Key         string    `json:"key"`
		Site        string    `json:"site"`
		Size        int       `json:"size"`
		Type        string    `json:"type"`
		Official    bool      `json:"official"`
		PublishedAt time.Time `json:"published_at"`
		ID          string    `json:"id"`
	} `json:"results"`
}

type TMDBExternalIds struct {
	ImdbID      string `json:"imdb_id"`
	WikidataID  string `json:"wikidata_id"`
	FacebookID  string `json:"facebook_id"`
	InstagramID string `json:"instagram_id"`
	TwitterID   string `json:"twitter_id"`
}

type TMDBExternalIdsMovie struct {
	TMDBExternalIds
}

type TMDBExternalIdsShow struct {
	TMDBExternalIds
	FreebaseMid string `json:"freebase_mid"`
	FreebaseID  string `json:"freebase_id"`
	TvdbID      int    `json:"tvdb_id"`
	TvrageID    int    `json:"tvrage_id"`
}

type TMDBKeywords struct {
	// ID      int `json:"id"`
	Results []struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
	} `json:"results"`
}

type TMDBRegions struct {
	Results []struct {
		ISO3166_1    string `json:"iso_3166_1"`
		English_Name string `json:"english_name"`
		Native_Name  string `json:"native_name"`
	} `json:"results"`
}
