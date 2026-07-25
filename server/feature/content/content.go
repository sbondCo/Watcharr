package content

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	gocache "github.com/robfig/go-cache"
	"github.com/sbondCo/Watcharr/cache"
	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/media/tmdb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// inmemory content cache
var ContentStore = gocache.New(time.Hour*24, time.Minute)

// Download file over http (used for downloading poster images)
// url - The remote file url.
// outf - Where should we store the downloaded file.
// force - Should we overwrite an existing file? If false, existing files will be skipped.
func download(url string, outf string, force bool) (err error) {
	slog.Debug("download: Attempting to download file", "url", url, "outf", outf, "force", force)
	// If not forced, skip call if file already exists to save unnecessary requests.
	if !force {
		if _, err := os.Stat(outf); !errors.Is(err, os.ErrNotExist) {
			slog.Debug("download: Skipping file, it already exists locally.", "outf", outf, "error", err)
			return nil
		} else {
			slog.Debug("download: Continuing to download file, it does not already exist.", "outf", outf, "error", err)
		}
	}

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		slog.Error("download: Failed to make request.", "outf", outf, "error", err)
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		slog.Error("download: Request failed. Non OK response.", "outf", outf, "status", resp.Status, "error", err)
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create the file
	out, err := os.Create(outf)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Warn("download: Failed to create out file, trying to recover by ensuring directories exist.", "outf", outf)
			err = os.MkdirAll(path.Dir(outf), 0764)
			if err != nil {
				slog.Error("download: Failed to create dir(s) in recovery attempt.", "outf", outf, "error", err)
				return err
			}
			// If dirs made, try making file again
			out, err = os.Create(outf)
			if err != nil {
				slog.Error("download: Failed to create out file again in recovery attempt.", "outf", outf, "error", err)
				return err
			}
			slog.Info("download: recovered by creating dir(s).", "outf", outf)
		} else {
			slog.Error("download: Failed to create out file. No known recovery path possible.", "outf", outf, "error", err)
			return err
		}
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		slog.Error("download: Failed to write file to our file.", "outf", outf, "error", err)
		return err
	}

	slog.Debug("download: Successfully downloaded file", "outf", outf)
	return nil
}

type Service struct {
	db   *gorm.DB
	tmdb *tmdb.TMDB
}

func NewService(db *gorm.DB, tmdb *tmdb.TMDB) *Service {
	return &Service{
		db,
		tmdb,
	}
}

// onlyUpdate - If we should only update existing row if exists, or false to create/update if not exist.
func (s *Service) saveContent(c *entity.Content, onlyUpdate bool) error {
	slog.Info("Saving content to db", "id", c.TmdbID, "title", c.Title)
	if c.TmdbID == 0 || c.Title == "" || c.Type == "" {
		slog.Error("saveContent: content missing id, title or type!", "id", c.TmdbID, "title", c.Title, "type", c.Type)
		return errors.New("content missing id or title")
	}
	var res *gorm.DB
	if onlyUpdate {
		// We only want to update an existing row, if it exists.
		res = s.db.Model(&entity.Content{}).Where("type = ? AND tmdb_id = ?", c.Type, c.TmdbID).Updates(c)
		if res.Error != nil {
			slog.Error("saveContent: Error updating content in database", "error", res.Error.Error())
			return errors.New("failed to update cached content in database")
		}
	} else {
		// On conflict, update existing row with details incase any were updated/missing.
		res = s.db.Clauses(clause.OnConflict{
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
		err := download(
			"https://image.tmdb.org/t/p/w500"+c.PosterPath,
			path.Join(config.DataPath, "img", c.PosterPath),
			false,
		)
		if err != nil {
			slog.Error("saveContent: Failed to download content image!", "error", err.Error())
		}
	}
	return nil
}

func (s *Service) cacheContentTv(content tmdb.TMDBShowDetails, onlyUpdate bool) (entity.Content, error) {
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

	c := entity.Content{
		TmdbID:           content.ID,
		Title:            content.Name,
		Overview:         content.Overview,
		PosterPath:       content.PosterPath,
		Type:             entity.SHOW,
		ReleaseDate:      &releaseDate,
		Popularity:       content.Popularity,
		VoteAverage:      content.VoteAverage,
		VoteCount:        content.VoteCount,
		Status:           content.Status,
		Runtime:          runtime,
		NumberOfEpisodes: content.NumberOfEpisodes,
		NumberOfSeasons:  content.NumberOfSeasons,
	}

	err = s.saveContent(&c, onlyUpdate)
	if err != nil {
		slog.Error("cacheContentTv: Failed to save content!", "error", err)
		return entity.Content{}, errors.New("failed to save content")
	}

	return c, nil
}

func (s *Service) cacheContentMovie(content tmdb.TMDBMovieDetails, onlyUpdate bool) (entity.Content, error) {
	var (
		releaseDate time.Time
	)
	var dateFormat = "2006-01-02"
	// Get details from movie/show response and fill out needed vars
	releaseDate, err := time.Parse(dateFormat, content.ReleaseDate)
	if err != nil {
		slog.Error("Failed to parse movie release date", "error", err)
	}

	c := entity.Content{
		TmdbID:      content.ID,
		Title:       content.Title,
		Overview:    content.Overview,
		PosterPath:  content.PosterPath,
		Type:        entity.MOVIE,
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

	err = s.saveContent(&c, onlyUpdate)
	if err != nil {
		slog.Error("cacheContentMovie: Failed to save content!", "error", err)
		return entity.Content{}, errors.New("failed to save content")
	}

	return c, nil
}

// Get content from our db cache, or cache it if it doesn't exist.
// fillTvFallback backfills empty text/poster fields from en-US when the
// configured language has no translation (TMDB returns empty fields in that case).
func (s *Service) fillTvFallback(id string, c *tmdb.TMDBShowDetails) {
	if s.tmdb.GetLang() == "en-US" || (c.Overview != "" && c.Name != "" && c.PosterPath != "") {
		return
	}
	en := new(tmdb.TMDBShowDetails)
	if err := s.tmdb.Request("/tv/"+id, map[string]string{"language": "en-US"}, &en); err != nil {
		return
	}
	if c.Overview == "" {
		c.Overview = en.Overview
	}
	if c.Name == "" {
		c.Name = en.Name
	}
	if c.PosterPath == "" {
		c.PosterPath = en.PosterPath
	}
}

// fillMovieFallback: same as fillTvFallback but for movies.
func (s *Service) fillMovieFallback(id string, c *tmdb.TMDBMovieDetails) {
	if s.tmdb.GetLang() == "en-US" || (c.Overview != "" && c.Title != "" && c.PosterPath != "") {
		return
	}
	en := new(tmdb.TMDBMovieDetails)
	if err := s.tmdb.Request("/movie/"+id, map[string]string{"language": "en-US"}, &en); err != nil {
		return
	}
	if c.Overview == "" {
		c.Overview = en.Overview
	}
	if c.Title == "" {
		c.Title = en.Title
	}
	if c.PosterPath == "" {
		c.PosterPath = en.PosterPath
	}
}

func (s *Service) GetOrCacheContent(contentType entity.ContentType, tmdbId int) (entity.Content, error) {
	var content entity.Content
	// Look in db for content.
	s.db.Where("type = ? AND tmdb_id = ?", contentType, tmdbId).Find(&content)
	// Create content if not found from our db.
	if content == (entity.Content{}) {
		slog.Debug("Content not in db, fetching...", "type", contentType, "tmdbId", tmdbId)

		resp, err := s.tmdb.APIRequest("/"+string(contentType)+"/"+strconv.Itoa(tmdbId), map[string]string{})
		if err != nil {
			slog.Error("GetOrCacheContent: content tmdb api request failed", "error", err)
			return entity.Content{}, errors.New("failed to find requested media")
		}

		if contentType == "movie" {
			c := new(tmdb.TMDBMovieDetails)
			err := json.Unmarshal([]byte(resp), &c)
			if err != nil {
				slog.Error("Failed to unmarshal movie details", "error", err)
				return entity.Content{}, errors.New("failed to process movie details response")
			}
			s.fillMovieFallback(strconv.Itoa(tmdbId), c)
			content, err = s.cacheContentMovie(*c, false)
			if err != nil {
				slog.Error("GetOrCacheContent: failed to cache movie content", "type", contentType, "content_id", tmdbId, "err", err)
				return entity.Content{}, errors.New("failed to cache content")
			}
		} else {
			c := new(tmdb.TMDBShowDetails)
			err := json.Unmarshal(resp, &c)
			if err != nil {
				slog.Error("Failed to unmarshal tv details", "error", err)
				return entity.Content{}, errors.New("failed to process tv details response")
			}
			s.fillTvFallback(strconv.Itoa(tmdbId), c)
			content, err = s.cacheContentTv(*c, false)
			if err != nil {
				slog.Error("GetOrCacheContent: failed to cache tv content", "type", contentType, "content_id", tmdbId, "err", err)
				return entity.Content{}, errors.New("failed to cache content")
			}
		}
	}
	return content, nil
}

// TMDB Multi Search.
func (s *Service) SearchContent(query string, pageNum int) (tmdb.TMDBSearchMultiResponse, error) {
	resp := new(tmdb.TMDBSearchMultiResponse)
	if pageNum == 0 {
		pageNum = 1
	}
	cacheKey := cache.CreateCacheKey("SearchContent", query, pageNum)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("SearchContent: Returning cache.")
		return *resp, nil
	}
	err := s.tmdb.Request("/search/multi", map[string]string{"query": query, "page": strconv.Itoa(pageNum)}, &resp)
	if err != nil {
		slog.Error("Failed to complete multi search request!", "error", err.Error())
		return tmdb.TMDBSearchMultiResponse{}, errors.New("failed to complete multi search request")
	}
	// English fallback for search overviews: TMDB returns empty overviews for
	// entries with no translation in the configured language. Do a single extra
	// en-US search and backfill by id, only when something is actually missing.
	if s.tmdb.GetLang() != "en-US" {
		missing := false
		for i := range resp.Results {
			if resp.Results[i].Overview == "" {
				missing = true
				break
			}
		}
		if missing {
			en := new(tmdb.TMDBSearchMultiResponse)
			if e := s.tmdb.Request("/search/multi", map[string]string{
				"query": query, "page": strconv.Itoa(pageNum), "language": "en-US",
			}, &en); e == nil {
				enOverview := make(map[int]string, len(en.Results))
				for i := range en.Results {
					enOverview[en.Results[i].ID] = en.Results[i].Overview
				}
				for i := range resp.Results {
					if resp.Results[i].Overview == "" {
						if ov := enOverview[resp.Results[i].ID]; ov != "" {
							resp.Results[i].Overview = ov
						}
					}
				}
			}
		}
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

func (s *Service) SearchMovies(query string, pageNum int) (tmdb.TMDBSearchMoviesResponse, error) {
	resp := new(tmdb.TMDBSearchMoviesResponse)
	if pageNum == 0 {
		pageNum = 1
	}
	cacheKey := cache.CreateCacheKey("SearchMovies", query, pageNum)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("SearchMovies: Returning cache.")
		return *resp, nil
	}
	err := s.tmdb.Request("/search/movie", map[string]string{"query": query, "page": strconv.Itoa(pageNum)}, &resp)
	if err != nil {
		slog.Error("Failed to complete movie search request!", "error", err.Error())
		return tmdb.TMDBSearchMoviesResponse{}, errors.New("failed to complete movie search request")
	}
	for i := range resp.Results {
		resp.Results[i].MediaType = "movie"
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

func (s *Service) SearchTv(query string, pageNum int) (tmdb.TMDBSearchShowsResponse, error) {
	resp := new(tmdb.TMDBSearchShowsResponse)
	if pageNum == 0 {
		pageNum = 1
	}
	cacheKey := cache.CreateCacheKey("SearchTv", query, pageNum)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("SearchTv: Returning cache.")
		return *resp, nil
	}
	err := s.tmdb.Request("/search/tv", map[string]string{"query": query, "page": strconv.Itoa(pageNum)}, &resp)
	if err != nil {
		slog.Error("Failed to complete tv search request!", "error", err.Error())
		return tmdb.TMDBSearchShowsResponse{}, errors.New("failed to complete tv search request")
	}
	for i := range resp.Results {
		resp.Results[i].MediaType = "tv"
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

func (s *Service) SearchPeople(query string, pageNum int) (tmdb.TMDBSearchPeopleResponse, error) {
	resp := new(tmdb.TMDBSearchPeopleResponse)
	if pageNum == 0 {
		pageNum = 1
	}
	cacheKey := cache.CreateCacheKey("SearchPeople", query, pageNum)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("SearchPeople: Returning cache.")
		return *resp, nil
	}
	err := s.tmdb.Request("/search/person", map[string]string{
		"query": query,
		"page":  strconv.Itoa(pageNum),
	}, &resp)
	if err != nil {
		slog.Error("Failed to complete people search request!", "error", err.Error())
		return tmdb.TMDBSearchPeopleResponse{}, errors.New("failed to complete people search request")
	}
	for i := range resp.Results {
		resp.Results[i].MediaType = "person"
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

// Search for content by an external id (imdb, etc).
// Defaults to imdb if no source if provided (probably most common).
func (s *Service) SearchByExternalId(id string, source string) (tmdb.TMDBSearchMultiResponse, error) {
	resp := new(tmdb.TMDBFindByExternalIdResponse)
	if source == "" {
		source = "imdb"
	}
	cacheKey := cache.CreateCacheKey("SearchByExternalId", id, source)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("SearchByExternalId: Got cache.")
	} else {
		// If not found in cache, request data from tmdb.
		err := s.tmdb.Request("/find/"+id, map[string]string{"external_source": source + "_id"}, &resp)
		if err != nil {
			slog.Error("Failed to complete find/external_id request!", "error", err.Error())
			return tmdb.TMDBSearchMultiResponse{}, errors.New("failed to complete find/external_id request")
		}
		ContentStore.Set(cacheKey, resp, time.Hour*24)
	}
	comb := []tmdb.TMDBSearchMultiResult{}
	comb = append(comb, resp.MovieResults...)
	comb = append(comb, resp.TvResults...)
	comb = append(comb, resp.PersonResults...)
	comb = append(comb, resp.TvSeasonResults...)
	comb = append(comb, resp.TvEpisodeResults...)
	return tmdb.TMDBSearchMultiResponse{TMDBSearchResponse: tmdb.TMDBSearchResponse[tmdb.TMDBSearchMultiResult]{
		Results: comb,
		TMDBPageFields: tmdb.TMDBPageFields{
			TotalResults: len(comb),
			// Just providing these so we don't break frontend pagination logic.
			TotalPages: 1,
			Page:       1,
		},
	}}, nil
}

func (s *Service) MovieDetails(
	id string,
	country string,
	rParams map[string]string,
) (tmdb.TMDBMovieDetails, error) {
	resp := new(tmdb.TMDBMovieDetails)
	cacheKey := cache.CreateCacheKey("MovieDetails", id, country, rParams)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("MovieDetails: Returning cache.")
		return *resp, nil
	}
	err := s.tmdb.Request("/movie/"+id, rParams, &resp)
	if err != nil {
		slog.Error("Failed to complete movie details request!",
			"error", err.Error())
		return tmdb.TMDBMovieDetails{},
			errors.New("failed to complete movie details request")
	}
	s.fillMovieFallback(id, resp)
	resp.WatchProvidersTransformed = transformProviders(&resp.WatchProviders, country)
	resp.WatchProviders = nil // We don't want this to linger around (in cache) since we have the transformed version now..
	go s.cacheContentMovie(*resp, true)
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

func (s *Service) MovieCredits(id string) (tmdb.TMDBContentCredits, error) {
	resp := new(tmdb.TMDBContentCredits)
	err := s.tmdb.Request("/movie/"+id+"/credits", map[string]string{}, &resp)
	if err != nil {
		slog.Error("Failed to complete movie cast request!", "error", err.Error())
		return tmdb.TMDBContentCredits{}, errors.New("failed to complete movie cast request")
	}
	return *resp, nil
}

func (s *Service) TvDetails(
	id string,
	country string,
	rParams map[string]string,
) (tmdb.TMDBShowDetails, error) {
	cacheKey := cache.CreateCacheKey("TvDetails", id, country, rParams)
	resp := new(tmdb.TMDBShowDetails)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("TvDetails: Returning cache.")
		return *resp, nil
	}
	err := s.tmdb.Request("/tv/"+id, rParams, &resp)
	if err != nil {
		slog.Error("Failed to complete tv details request!", "error", err.Error())
		return tmdb.TMDBShowDetails{}, errors.New("failed to complete tv details request")
	}
	s.fillTvFallback(id, resp)
	resp.WatchProvidersTransformed = transformProviders(&resp.WatchProviders, country)
	resp.WatchProviders = nil // We don't want this to linger around (in cache) since we have the transformed version now..
	go s.cacheContentTv(*resp, true)
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

func (s *Service) TvCredits(id string) (tmdb.TMDBContentCredits, error) {
	resp := new(tmdb.TMDBContentCredits)
	err := s.tmdb.Request("/tv/"+id+"/credits", map[string]string{}, &resp)
	if err != nil {
		slog.Error("Failed to complete tv cast request!", "error", err.Error())
		return tmdb.TMDBContentCredits{}, errors.New("failed to complete tv cast request")
	}
	return *resp, nil
}

// This method is manually cached, so it can be easily used in other places (on the server) with cache benefits
func (s *Service) SeasonDetails(tvId string, seasonNumber string) (tmdb.TMDBSeasonDetails, error) {
	cacheKey := cache.CreateCacheKey("SeasonDetails", tvId, seasonNumber)
	resp := new(tmdb.TMDBSeasonDetails)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("SeasonDetails: Returning cache.")
		return *resp, nil
	}
	err := s.tmdb.Request("/tv/"+tvId+"/season/"+seasonNumber, map[string]string{}, &resp)
	if err != nil {
		slog.Error("SeasonDetails: Failed to complete season details request!", "error", err.Error())
		return tmdb.TMDBSeasonDetails{}, errors.New("failed to complete season details request")
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

func (s *Service) PersonDetails(id string) (tmdb.TMDBPersonDetails, error) {
	resp := new(tmdb.TMDBPersonDetails)
	err := s.tmdb.Request("/person/"+id, map[string]string{}, &resp)
	if err != nil {
		slog.Error("Failed to complete person details request!", "error", err.Error())
		return tmdb.TMDBPersonDetails{}, errors.New("failed to complete person details request")
	}
	return *resp, nil
}

func (s *Service) PersonCredits(id string) (tmdb.TMDBPersonCombinedCredits, error) {
	cacheKey := cache.CreateCacheKey("PersonCredits", id)
	resp := new(tmdb.TMDBPersonCombinedCredits)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("PersonCredits: Returning cache.")
		return *resp, nil
	}
	err := s.tmdb.Request("/person/"+id+"/combined_credits", map[string]string{}, &resp)
	if err != nil {
		slog.Error("Failed to complete person details request!", "error", err.Error())
		return tmdb.TMDBPersonCombinedCredits{}, errors.New("failed to complete person details request")
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

func (s *Service) Trending(t tmdb.TrendingType, pageNum int, region string) (tmdb.TMDBTrendingCombined, error) {
	resp := new(tmdb.TMDBTrendingCombined)
	if t != tmdb.TrendingTypeAll &&
		t != tmdb.TrendingTypeMovie &&
		t != tmdb.TrendingTypeShow &&
		t != tmdb.TrendingTypePerson {
		slog.Error("Trending: Invalid type provided", "provided_t", t)
		return *resp, errors.New("invalid type")
	}
	if pageNum <= 0 {
		pageNum = 1
	}
	cacheKey := cache.CreateCacheKey("Trending", string(t), region, pageNum)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("Trending: Returning cache.")
		return *resp, nil
	}
	err := s.tmdb.Request("/trending/"+string(t)+"/day", map[string]string{
		"page":   strconv.Itoa(pageNum),
		"region": region,
	}, &resp)
	if err != nil {
		slog.Error("Failed to complete all trending request!", "error", err.Error())
		return *resp, errors.New("failed to complete all trending request")
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

func (s *Service) DiscoverMovies(
	o tmdb.DiscoverOptions,
	pageNum int,
	region string,
) (tmdb.TMDBDiscoverMovies, error) {
	resp := new(tmdb.TMDBDiscoverMovies)
	reqParams := map[string]string{
		"page":   strconv.Itoa(pageNum),
		"region": region,
	}
	s.applyDiscoverOptionsToMap(true, o, reqParams)
	cacheKey := cache.CreateCacheKey("DiscoverMovies", pageNum, reqParams)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("DiscoverMovies: Returning cache.")
		return *resp, nil
	}
	err := s.tmdb.Request("/discover/movie", reqParams, &resp)
	if err != nil {
		slog.Error("DiscoverMovies: Failed to complete request!", "error", err.Error())
		return tmdb.TMDBDiscoverMovies{}, errors.New("failed to complete discover movies request")
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

func (s *Service) DiscoverTv(
	o tmdb.DiscoverOptions,
	pageNum int,
	region string,
) (tmdb.TMDBDiscoverShows, error) {
	resp := new(tmdb.TMDBDiscoverShows)
	reqParams := map[string]string{
		"page":   strconv.Itoa(pageNum),
		"region": region,
	}
	s.applyDiscoverOptionsToMap(false, o, reqParams)
	cacheKey := cache.CreateCacheKey("DiscoverTv", pageNum, reqParams)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("DiscoverTv: Returning cache.")
		return *resp, nil
	}
	err := s.tmdb.Request("/discover/tv", reqParams, &resp)
	if err != nil {
		slog.Error("DiscoverTv: Failed to complete request!", "error", err.Error())
		return tmdb.TMDBDiscoverShows{}, errors.New("failed to complete discover tv request")
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

func (s *Service) applyDiscoverOptionsToMap(
	// Some properties are named differently for sorting the same thing as far
	// as we care, so we need to differenciate to name them properly.
	forMovie bool,
	o tmdb.DiscoverOptions,
	m map[string]string,
) {
	releaseDateMinKey := "release_date.gte"
	releaseDateMaxKey := "release_date.lte"
	withReleaseTypeKey := "with_release_type"
	if !forMovie {
		// Replace with names for equivalent tv filters
		releaseDateMinKey = "first_air_date.gte"
		releaseDateMaxKey = "first_air_date.lte"
		withReleaseTypeKey = "with_type"
	}
	if !o.ReleaseDateMin.IsZero() {
		m[releaseDateMinKey] = o.ReleaseDateMin.Format("2006-01-02")
	}
	if !o.ReleaseDateMax.IsZero() {
		m[releaseDateMaxKey] = o.ReleaseDateMax.Format("2006-01-02")
	}
	if o.WithReleaseType != "" {
		m[withReleaseTypeKey] = o.WithReleaseType
	}
}

func (s *Service) PopularPeople(pageNum int) (tmdb.TMDBPopularPeople, error) {
	cacheKey := cache.CreateCacheKey("PopularPeople", pageNum)
	resp := new(tmdb.TMDBPopularPeople)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("PopularPeople: Returning cache.")
		return *resp, nil
	}
	err := s.tmdb.Request("/person/popular",
		map[string]string{"page": strconv.Itoa(pageNum)},
		&resp)
	if err != nil {
		slog.Error("PopularPeople: Failed to complete request!", "error", err.Error())
		return tmdb.TMDBPopularPeople{}, errors.New("failed to complete request")
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

func (s *Service) Regions() (tmdb.TMDBRegions, error) {
	resp := new(tmdb.TMDBRegions)
	err := s.tmdb.Request("/watch/providers/regions", map[string]string{}, &resp)
	if err != nil {
		slog.Error("Failed to complete regions request!", "error", err.Error())
		return tmdb.TMDBRegions{}, errors.New("failed to complete regions request")
	}
	return *resp, nil
}
