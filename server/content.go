package main

import (
	"errors"
	"log/slog"
	"time"
)

type ContentType string

const (
	MOVIE ContentType = "movie"
	SHOW  ContentType = "tv"
)

// For storing cached content, so we can serve the basic local data for watched list to work
type Content struct {
	ID               int         `json:"id" gorm:"primaryKey;autoIncrement"`
	TmdbID           int         `json:"tmdbId" gorm:"uniqueIndex:contentidtotypeidx;not null"`
	Title            string      `json:"title"`
	PosterPath       string      `json:"poster_path"`
	Overview         string      `json:"overview"`
	Type             ContentType `json:"type" gorm:"uniqueIndex:contentidtotypeidx;not null"`
	ReleaseDate      *time.Time  `json:"release_date,omitempty"`
	Popularity       float32     `json:"popularity"`
	VoteAverage      float32     `json:"vote_average"`
	VoteCount        uint32      `json:"vote_count"`
	ImdbID           string      `json:"imdb_id"`
	Status           string      `json:"status"`
	Budget           uint32      `json:"budget"`
	Revenue          uint32      `json:"revenue"`
	Runtime          uint32      `json:"runtime"`
	NumberOfEpisodes uint32      `json:"numberOfEpisodes"`
	NumberOfSeasons  uint32      `json:"numberOfSeasons"`
}

// Getting only region needed from api is not a feature yet
// https://trello.com/c/75tR4cpF/106-add-watch-provider-region-filtering
// When it is, this can be removed for that instead.
func transformProviders(c *interface{}, country string) {
	slog.Debug("transformProviders called", "country", country)
	if cmap, ok := (*c).(map[string]interface{}); ok {
		if rmap, ok := cmap["results"].(map[string]interface{}); ok {
			if val, ok := rmap[country]; ok {
				slog.Debug("transformProviders: Found country.. overwriting whole object", "new_obj", val)
				if rvmap, ok := val.(map[string]interface{}); ok {
					rvmap["country"] = country
				}
				*c = val
			} else {
				slog.Warn("transformProviders: Couldn't find country..", "country", country)
			}
		} else {
			slog.Warn("transformProviders: Couldn't find results property..")
		}
	} else {
		slog.Error("transformProviders: Assertion failed")
	}
}

func searchContent(query string) (TMDBSearchMultiResponse, error) {
	resp := new(TMDBSearchMultiResponse)
	err := tmdbRequest("/search/multi", map[string]string{"query": query, "page": "1"}, &resp)
	if err != nil {
		slog.Error("Failed to complete multi search request!", "error", err.Error())
		return TMDBSearchMultiResponse{}, errors.New("failed to complete multi search request")
	}
	return *resp, nil
}

func movieDetails(id string, country string, rParams map[string]string) (TMDBMovieDetails, error) {
	resp := new(TMDBMovieDetails)
	err := tmdbRequest("/movie/"+id, rParams, &resp)
	if err != nil {
		slog.Error("Failed to complete movie details request!", "error", err.Error())
		return TMDBMovieDetails{}, errors.New("failed to complete movie details request")
	}
	transformProviders(&resp.WatchProviders, country)
	return *resp, nil
}

func movieCredits(id string) (TMDBContentCredits, error) {
	resp := new(TMDBContentCredits)
	err := tmdbRequest("/movie/"+id+"/credits", map[string]string{}, &resp)
	if err != nil {
		slog.Error("Failed to complete movie cast request!", "error", err.Error())
		return TMDBContentCredits{}, errors.New("failed to complete movie cast request")
	}
	return *resp, nil
}

func tvDetails(id string, country string, rParams map[string]string) (TMDBShowDetails, error) {
	resp := new(TMDBShowDetails)
	err := tmdbRequest("/tv/"+id, rParams, &resp)
	if err != nil {
		slog.Error("Failed to complete tv details request!", "error", err.Error())
		return TMDBShowDetails{}, errors.New("failed to complete tv details request")
	}
	transformProviders(&resp.WatchProviders, country)
	return *resp, nil
}

func tvCredits(id string) (TMDBContentCredits, error) {
	resp := new(TMDBContentCredits)
	err := tmdbRequest("/tv/"+id+"/credits", map[string]string{}, &resp)
	if err != nil {
		slog.Error("Failed to complete tv cast request!", err.Error())
		return TMDBContentCredits{}, errors.New("failed to complete tv cast request")
	}
	return *resp, nil
}

func seasonDetails(tvId string, seasonNumber string) (TMDBSeasonDetails, error) {
	resp := new(TMDBSeasonDetails)
	err := tmdbRequest("/tv/"+tvId+"/season/"+seasonNumber, map[string]string{}, &resp)
	if err != nil {
		slog.Error("Failed to complete season details request!", "error", err.Error())
		return TMDBSeasonDetails{}, errors.New("failed to complete season details request")
	}
	return *resp, nil
}

func personDetails(id string) (TMDBPersonDetails, error) {
	resp := new(TMDBPersonDetails)
	err := tmdbRequest("/person/"+id, map[string]string{}, &resp)
	if err != nil {
		slog.Error("Failed to complete person details request!", "error", err.Error())
		return TMDBPersonDetails{}, errors.New("failed to complete person details request")
	}
	return *resp, nil
}

func personCredits(id string) (TMDBPersonCombinedCredits, error) {
	resp := new(TMDBPersonCombinedCredits)
	err := tmdbRequest("/person/"+id+"/combined_credits", map[string]string{}, &resp)
	if err != nil {
		slog.Error("Failed to complete person details request!", "error", err.Error())
		return TMDBPersonCombinedCredits{}, errors.New("failed to complete person details request")
	}
	return *resp, nil
}

func discoverMovies() (TMDBDiscoverMovies, error) {
	resp := new(TMDBDiscoverMovies)
	err := tmdbRequest("/discover/movie/", map[string]string{"page": "1"}, &resp)
	if err != nil {
		slog.Error("Failed to complete discover movies request!", "error", err.Error())
		return TMDBDiscoverMovies{}, errors.New("failed to complete discover movies request")
	}
	return *resp, nil
}

func discoverTv() (TMDBDiscoverShows, error) {
	resp := new(TMDBDiscoverShows)
	err := tmdbRequest("/discover/tv/", map[string]string{"page": "1"}, &resp)
	if err != nil {
		slog.Error("Failed to complete discover tv request!", "error", err.Error())
		return TMDBDiscoverShows{}, errors.New("failed to complete discover tv request")
	}
	return *resp, nil
}

func allTrending() (TMDBTrendingAll, error) {
	resp := new(TMDBTrendingAll)
	err := tmdbRequest("/trending/all/day", map[string]string{}, &resp)
	if err != nil {
		slog.Error("Failed to complete all trending request!", "error", err.Error())
		return TMDBTrendingAll{}, errors.New("failed to complete all trending request")
	}
	return *resp, nil
}

func upcomingMovies() (TMDBUpcomingMovies, error) {
	resp := new(TMDBUpcomingMovies)
	err := tmdbRequest("/movie/upcoming", map[string]string{"page": "1"}, &resp)
	if err != nil {
		slog.Error("Failed to complete upcoming movies request!", "error", err.Error())
		return TMDBUpcomingMovies{}, errors.New("failed to complete upcoming movies request")
	}
	return *resp, nil
}

// Theres no upcoming endpoint for tv ;( - using discover with future dates
func upcomingTv() (TMDBUpcomingShows, error) {
	resp := new(TMDBUpcomingShows)
	dFmt := "2006-01-02"
	mind := time.Now().Format(dFmt)
	maxd := time.Now().AddDate(0, 0, 15).Format(dFmt)
	err := tmdbRequest("/discover/tv", map[string]string{"page": "1", "first_air_date.gte": mind, "first_air_date.lte": maxd, "sort_by": "popularity.desc", "with_type": "2|3"}, &resp)
	if err != nil {
		slog.Error("Failed to complete upcoming tv request!", "error", err.Error())
		return TMDBUpcomingShows{}, errors.New("failed to complete upcoming tv request")
	}
	return *resp, nil
}
