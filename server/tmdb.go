package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
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
	Popularity       float64  `json:"popularity"`
	ReleaseDate      string   `json:"release_date,omitempty"`
	Video            bool     `json:"video,omitempty"`
	VoteAverage      float64  `json:"vote_average"`
	VoteCount        int      `json:"vote_count"`
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
	Popularity          float64 `json:"popularity"`
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
	VoteAverage     float64 `json:"vote_average"`
	VoteCount       int     `json:"vote_count"`
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
	Budget              int    `json:"budget"`
	ImdbID              string `json:"imdb_id"`
	OriginalTitle       string `json:"original_title"`
	ReleaseDate         string `json:"release_date"`
	Revenue             int    `json:"revenue"`
	Runtime             int    `json:"runtime"`
	Title               string `json:"title"`
	Video               bool   `json:"video"`
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
		VoteCount      int     `json:"vote_count"`
	} `json:"last_episode_to_air"`
	Name             string `json:"name"`
	NextEpisodeToAir any    `json:"next_episode_to_air"`
	Networks         []struct {
		Name          string `json:"name"`
		ID            int    `json:"id"`
		LogoPath      string `json:"logo_path"`
		OriginCountry string `json:"origin_country"`
	} `json:"networks"`
	NumberOfEpisodes int      `json:"number_of_episodes"`
	NumberOfSeasons  int      `json:"number_of_seasons"`
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
	VoteCount        int      `json:"vote_count"`
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

func tmdbAPIRequest(ep string, p map[string]string) ([]byte, error) {
	println("tmdbAPIRequest:", ep)
	base, err := url.Parse("https://api.themoviedb.org/3")
	if err != nil {
		return nil, errors.New("failed to parse api uri")
	}

	// Path params
	base.Path += ep

	// Query params
	params := url.Values{}
	params.Add("api_key", "d047fa61d926371f277e7a83c9c4ff2c")
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
		println("TMDB non 200 status code:", res.StatusCode)
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
