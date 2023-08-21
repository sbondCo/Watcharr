package main

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

type TMDBSearchMultiResponse struct {
	Page         int                      `json:"page"`
	Results      []TMDBSearchMultiResults `json:"results"`
	TotalPages   int                      `json:"total_pages"`
	TotalResults int                      `json:"total_results"`
}

type TMDBSearchMultiResults struct {
	Adult            bool     `json:"adult"`
	BackdropPath     string   `json:"backdrop_path"`
	ID               int      `json:"id"`
	Title            string   `json:"title,omitempty"`
	OriginalLanguage string   `json:"original_language"`
	OriginalTitle    string   `json:"original_title,omitempty"`
	Overview         string   `json:"overview"`
	PosterPath       string   `json:"poster_path"`
	ProfilePath      string   `json:"profile_path"`
	MediaType        string   `json:"media_type"`
	GenreIds         []int64  `json:"genre_ids"`
	Popularity       float32  `json:"popularity"`
	ReleaseDate      string   `json:"release_date,omitempty"`
	Video            bool     `json:"video,omitempty"`
	VoteAverage      float32  `json:"vote_average"`
	VoteCount        uint32   `json:"vote_count"`
	Name             string   `json:"name,omitempty"`
	OriginalName     string   `json:"original_name,omitempty"`
	FirstAirDate     string   `json:"first_air_date,omitempty"`
	OriginCountry    []string `json:"origin_country,omitempty"`
}

type TMDBContentDetails struct {
	ID           int    `json:"id"`
	BackdropPath string `json:"backdrop_path"`
	Genres       []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"genres"`
	PosterPath          string  `json:"poster_path"`
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
}

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
	Videos         TMDBContentVideos `json:"videos"`
	WatchProviders interface{}       `json:"watch/providers"`
}

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
	Videos         TMDBContentVideos `json:"videos"`
	WatchProviders interface{}       `json:"watch/providers"`
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

type TMDBPersonDetails struct {
	Birthday           string   `json:"birthday"`
	KnownForDepartment string   `json:"known_for_department"`
	Deathday           string   `json:"deathday"`
	ID                 int      `json:"id"`
	Name               string   `json:"name"`
	AlsoKnownAs        []string `json:"also_known_as"`
	Gender             int8     `json:"gender"`
	Biography          string   `json:"biography"`
	Popularity         float32  `json:"popularity"`
	PlaceOfBirth       string   `json:"place_of_birth"`
	ProfilePath        string   `json:"profile_path"`
	Adult              bool     `json:"adult"`
	ImdbID             string   `json:"imdb_id"`
	Homepage           string   `json:"homepage"`
}

type TMDBPersonCombinedCredits struct {
	ID   int                             `json:"id"`
	Cast []TMDBPersonCombinedCreditsCast `json:"cast"`
	// crew TMDBPersonCombinedCreditsCrew
}

type TMDBPersonCombinedCreditsCast struct {
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

type TMDBDiscoverMovies struct {
	Page    int `json:"page"`
	Results []struct {
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
	} `json:"results"`
	TotalPages   int `json:"total_pages"`
	TotalResults int `json:"total_results"`
}

type TMDBDiscoverShows struct {
	Page    int `json:"page"`
	Results []struct {
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
	} `json:"results"`
	TotalPages   int `json:"total_pages"`
	TotalResults int `json:"total_results"`
}

type TMDBTrendingAll struct {
	Page    int `json:"page"`
	Results []struct {
		Adult            bool     `json:"adult"`
		BackdropPath     string   `json:"backdrop_path"`
		ID               int      `json:"id"`
		Title            string   `json:"title,omitempty"`
		OriginalLanguage string   `json:"original_language"`
		OriginalTitle    string   `json:"original_title,omitempty"`
		Overview         string   `json:"overview"`
		PosterPath       string   `json:"poster_path"`
		MediaType        string   `json:"media_type"`
		GenreIds         []int    `json:"genre_ids"`
		Popularity       float64  `json:"popularity"`
		ReleaseDate      string   `json:"release_date,omitempty"`
		Video            bool     `json:"video,omitempty"`
		VoteAverage      float64  `json:"vote_average"`
		VoteCount        int      `json:"vote_count"`
		Name             string   `json:"name,omitempty"`
		OriginalName     string   `json:"original_name,omitempty"`
		FirstAirDate     string   `json:"first_air_date,omitempty"`
		OriginCountry    []string `json:"origin_country,omitempty"`
	} `json:"results"`
	TotalPages   int `json:"total_pages"`
	TotalResults int `json:"total_results"`
}

type TMDBUpcomingMovies struct {
	Dates struct {
		Maximum string `json:"maximum"`
		Minimum string `json:"minimum"`
	} `json:"dates"`
	Page    int `json:"page"`
	Results []struct {
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
		VoteAverage      float32 `json:"vote_average"`
		VoteCount        int     `json:"vote_count"`
	} `json:"results"`
	TotalPages   int `json:"total_pages"`
	TotalResults int `json:"total_results"`
}

type TMDBUpcomingShows struct {
	Page    int `json:"page"`
	Results []struct {
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
	} `json:"results"`
	TotalPages   int `json:"total_pages"`
	TotalResults int `json:"total_results"`
}

func tmdbAPIRequest(ep string, p map[string]string) ([]byte, error) {
	slog.Debug("tmdbAPIRequest", "endpoint", ep, "params", p)
	base, err := url.Parse("https://api.themoviedb.org/3")
	if err != nil {
		return nil, errors.New("failed to parse api uri")
	}

	// Path params
	base.Path += ep

	// Query params
	params := url.Values{}
	params.Add("api_key", TMDBKey)
	params.Add("language", "en-US")
	for k, v := range p {
		params.Add(k, v)
	}

	// Add params to url
	base.RawQuery = params.Encode()

	// Run get request
	res, err := http.Get(base.String())
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		slog.Error("TMDB non 200 status code:", "status_code", res.StatusCode)
		return nil, errors.New(string(body))
	}
	return body, nil
}

func tmdbRequest(ep string, p map[string]string, resp interface{}) error {
	body, err := tmdbAPIRequest(ep, p)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(body), &resp)
	if err != nil {
		return err
	}
	return nil
}
