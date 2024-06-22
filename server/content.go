package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"path"
	"strconv"
	"time"

	"github.com/gin-contrib/cache/persistence"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ContentType string

const (
	MOVIE ContentType = "movie"
	SHOW  ContentType = "tv"
)

var ContentStore = persistence.NewInMemoryStore(time.Hour * 24)

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

// onlyUpdate - If we should only update existing row if exists, or false to create/update if not exist.
func saveContent(db *gorm.DB, c *Content, onlyUpdate bool) error {
	slog.Info("Saving content to db", "id", c.TmdbID, "title", c.Title)
	if c.TmdbID == 0 || c.Title == "" || c.Type == "" {
		slog.Error("saveContent: content missing id, title or type!", "id", c.TmdbID, "title", c.Title, "type", c.Type)
		return errors.New("content missing id or title")
	}
	var res *gorm.DB
	if onlyUpdate {
		// We only want to update an existing row, if it exists.
		res = db.Model(&Content{}).Where("type = ? AND tmdb_id = ?", c.Type, c.TmdbID).Updates(c)
		if res.Error != nil {
			slog.Error("saveContent: Error updating content in database", "error", res.Error.Error())
			return errors.New("failed to update cached content in database")
		}
	} else {
		// On conflict, update existing row with details incase any were updated/missing.
		res = db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "tmdb_id"}, {Name: "type"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"title",
				"poster_path",
				"overview",
				"release_date",
				"popularity",
				"vote_average",
				"vote_count",
				"imdb_id",
				"status",
				"budget",
				"revenue",
				"runtime",
				"number_of_episodes",
				"number_of_seasons",
			}),
		}).Create(&c)
		if res.Error != nil {
			// Error if anything but unique contraint error
			if res.Error != gorm.ErrDuplicatedKey {
				slog.Error("saveContent: Error creating content in database", "error", res.Error.Error())
				return errors.New("failed to cache content in database")
			}
		}
	}
	// If row created, download the image
	if res.RowsAffected > 0 {
		slog.Debug("saveContent: Downloading poster.")
		err := download("https://image.tmdb.org/t/p/w500"+c.PosterPath, path.Join(DataPath, "img", c.PosterPath))
		if err != nil {
			slog.Error("saveContent: Failed to download content image!", "error", err.Error())
		}
	}
	return nil
}

func cacheContentTv(db *gorm.DB, content TMDBShowDetails, onlyUpdate bool) (Content, error) {
	slog.Debug("cacheContentTv", "content", content)
	var (
		releaseDate time.Time
		runtime     uint32
	)
	var dateFormat = "2006-01-02"
	releaseDate, err := time.Parse(dateFormat, content.FirstAirDate)
	if err != nil {
		slog.Error("Failed to parse tv release date", "error", err)
	}
	if len(content.EpisodeRunTime) > 0 {
		runtime = uint32(content.EpisodeRunTime[0])
	}

	c := Content{
		TmdbID:           content.ID,
		Title:            content.Name,
		Overview:         content.Overview,
		PosterPath:       content.PosterPath,
		Type:             SHOW,
		ReleaseDate:      &releaseDate,
		Popularity:       content.Popularity,
		VoteAverage:      content.VoteAverage,
		VoteCount:        content.VoteCount,
		Status:           content.Status,
		Runtime:          runtime,
		NumberOfEpisodes: content.NumberOfEpisodes,
		NumberOfSeasons:  content.NumberOfSeasons,
	}

	err = saveContent(db, &c, onlyUpdate)
	if err != nil {
		slog.Error("cacheContentTv: Failed to save content!", "error", err)
		return Content{}, errors.New("failed to save content")
	}

	return c, nil
}

func cacheContentMovie(db *gorm.DB, content TMDBMovieDetails, onlyUpdate bool) (Content, error) {
	var (
		releaseDate time.Time
	)
	var dateFormat = "2006-01-02"
	// Get details from movie/show response and fill out needed vars
	releaseDate, err := time.Parse(dateFormat, content.ReleaseDate)
	if err != nil {
		slog.Error("Failed to parse movie release date", "error", err)
	}

	c := Content{
		TmdbID:      content.ID,
		Title:       content.Title,
		Overview:    content.Overview,
		PosterPath:  content.PosterPath,
		Type:        MOVIE,
		ReleaseDate: &releaseDate,
		Popularity:  content.Popularity,
		VoteAverage: content.VoteAverage,
		VoteCount:   content.VoteCount,
		ImdbID:      content.ImdbID,
		Status:      content.Status,
		Budget:      content.Budget,
		Revenue:     content.Revenue,
		Runtime:     content.Runtime,
	}

	err = saveContent(db, &c, onlyUpdate)
	if err != nil {
		slog.Error("cacheContentMovie: Failed to save content!", "error", err)
		return Content{}, errors.New("failed to save content")
	}

	return c, nil
}

// Get content from our cache, or cache it if it doesn't exist.
func getOrCacheContent(db *gorm.DB, contentType ContentType, tmdbId int) (Content, error) {
	var content Content
	// Look in db for content.
	db.Where("type = ? AND tmdb_id = ?", contentType, tmdbId).Find(&content)
	// Create content if not found from our db.
	if content == (Content{}) {
		slog.Debug("Content not in db, fetching...", "type", contentType, "tmdbId", tmdbId)

		resp, err := tmdbAPIRequest("/"+string(contentType)+"/"+strconv.Itoa(tmdbId), map[string]string{})
		if err != nil {
			slog.Error("getOrCacheContent: content tmdb api request failed", "error", err)
			return Content{}, errors.New("failed to find requested media")
		}

		if contentType == "movie" {
			c := new(TMDBMovieDetails)
			err := json.Unmarshal([]byte(resp), &c)
			if err != nil {
				slog.Error("Failed to unmarshal movie details", "error", err)
				return Content{}, errors.New("failed to process movie details response")
			}
			content, err = cacheContentMovie(db, *c, false)
			if err != nil {
				slog.Error("getOrCacheContent: failed to cache movie content", "type", contentType, "content_id", tmdbId, "err", err)
				return Content{}, errors.New("failed to cache content")
			}
		} else {
			c := new(TMDBShowDetails)
			err := json.Unmarshal(resp, &c)
			if err != nil {
				slog.Error("Failed to unmarshal tv details", "error", err)
				return Content{}, errors.New("failed to process tv details response")
			}
			content, err = cacheContentTv(db, *c, false)
			if err != nil {
				slog.Error("getOrCacheContent: failed to cache tv content", "type", contentType, "content_id", tmdbId, "err", err)
				return Content{}, errors.New("failed to cache content")
			}
		}
	}
	return content, nil
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

func searchContent(query string, pageNum int) (TMDBSearchMultiResponse, error) {
	resp := new(TMDBSearchMultiResponse)
	if pageNum == 0 {
		pageNum = 1
	}
	err := tmdbRequest("/search/multi", map[string]string{"query": query, "page": strconv.Itoa(pageNum)}, &resp)
	if err != nil {
		slog.Error("Failed to complete multi search request!", "error", err.Error())
		return TMDBSearchMultiResponse{}, errors.New("failed to complete multi search request")
	}
	return *resp, nil
}

func searchMovies(query string, pageNum int) (TMDBSearchMoviesResponse, error) {
	resp := new(TMDBSearchMoviesResponse)
	if pageNum == 0 {
		pageNum = 1
	}
	err := tmdbRequest("/search/movie", map[string]string{"query": query, "page": strconv.Itoa(pageNum)}, &resp)
	if err != nil {
		slog.Error("Failed to complete movie search request!", "error", err.Error())
		return TMDBSearchMoviesResponse{}, errors.New("failed to complete movie search request")
	}
	for i := range resp.Results {
		resp.Results[i].MediaType = "movie"
	}
	return *resp, nil
}

func searchTv(query string, pageNum int) (TMDBSearchShowsResponse, error) {
	resp := new(TMDBSearchShowsResponse)
	if pageNum == 0 {
		pageNum = 1
	}
	err := tmdbRequest("/search/tv", map[string]string{"query": query, "page": strconv.Itoa(pageNum)}, &resp)
	if err != nil {
		slog.Error("Failed to complete tv search request!", "error", err.Error())
		return TMDBSearchShowsResponse{}, errors.New("failed to complete tv search request")
	}
	for i := range resp.Results {
		resp.Results[i].MediaType = "tv"
	}
	return *resp, nil
}

func searchPeople(query string, pageNum int) (TMDBSearchPeopleResponse, error) {
	resp := new(TMDBSearchPeopleResponse)
	if pageNum == 0 {
		pageNum = 1
	}
	err := tmdbRequest("/search/person", map[string]string{"query": query, "page": strconv.Itoa(pageNum)}, &resp)
	if err != nil {
		slog.Error("Failed to complete people search request!", "error", err.Error())
		return TMDBSearchPeopleResponse{}, errors.New("failed to complete people search request")
	}
	for i := range resp.Results {
		resp.Results[i].MediaType = "person"
	}
	return *resp, nil
}

func movieDetails(db *gorm.DB, id string, country string, rParams map[string]string) (TMDBMovieDetails, error) {
	resp := new(TMDBMovieDetails)
	err := tmdbRequest("/movie/"+id, rParams, &resp)
	if err != nil {
		slog.Error("Failed to complete movie details request!", "error", err.Error())
		return TMDBMovieDetails{}, errors.New("failed to complete movie details request")
	}
	transformProviders(&resp.WatchProviders, country)
	go cacheContentMovie(db, *resp, true)
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

func tvDetails(db *gorm.DB, id string, country string, rParams map[string]string) (TMDBShowDetails, error) {
	resp := new(TMDBShowDetails)
	err := tmdbRequest("/tv/"+id, rParams, &resp)
	if err != nil {
		slog.Error("Failed to complete tv details request!", "error", err.Error())
		return TMDBShowDetails{}, errors.New("failed to complete tv details request")
	}
	transformProviders(&resp.WatchProviders, country)
	go cacheContentTv(db, *resp, true)
	return *resp, nil
}

func tvCredits(id string) (TMDBContentCredits, error) {
	resp := new(TMDBContentCredits)
	err := tmdbRequest("/tv/"+id+"/credits", map[string]string{}, &resp)
	if err != nil {
		slog.Error("Failed to complete tv cast request!", "error", err.Error())
		return TMDBContentCredits{}, errors.New("failed to complete tv cast request")
	}
	return *resp, nil
}

// This method is manually cached, so it can be easily used in other places (on the server) with cache benefits
func seasonDetails(tvId string, seasonNumber string) (TMDBSeasonDetails, error) {
	var cacheKey = "contentstore-seasondetails-" + tvId + "-" + seasonNumber
	resp := new(TMDBSeasonDetails)
	if err := ContentStore.Get(cacheKey, &resp); err != nil {
		if err != persistence.ErrCacheMiss {
			slog.Error("seasonDetails: Cache failed for some reason", "error", err)
		}
	} else {
		slog.Debug("seasonDetails: Returning cache.")
		return *resp, nil
	}
	err := tmdbRequest("/tv/"+tvId+"/season/"+seasonNumber, map[string]string{}, &resp)
	if err != nil {
		slog.Error("seasonDetails: Failed to complete season details request!", "error", err.Error())
		return TMDBSeasonDetails{}, errors.New("failed to complete season details request")
	}
	if err := ContentStore.Set(cacheKey, resp, time.Hour*24); err != nil {
		slog.Error("seasonDetails: Failed to set cache!", "error", err)
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
	err := tmdbRequest("/discover/movie", map[string]string{"page": "1"}, &resp)
	if err != nil {
		slog.Error("Failed to complete discover movies request!", "error", err.Error())
		return TMDBDiscoverMovies{}, errors.New("failed to complete discover movies request")
	}
	return *resp, nil
}

func discoverTv() (TMDBDiscoverShows, error) {
	resp := new(TMDBDiscoverShows)
	err := tmdbRequest("/discover/tv", map[string]string{"page": "1"}, &resp)
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

func regions() (TMDBRegions, error) {
	resp := new(TMDBRegions)
	err := tmdbRequest("/watch/providers/regions", map[string]string{}, &resp)
	if err != nil {
		slog.Error("Failed to complete regions request!", "error", err.Error())
		return TMDBRegions{}, errors.New("failed to complete regions request")
	}
	return *resp, nil
}
