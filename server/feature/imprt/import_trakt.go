// Trakt.tv importer.

package imprt

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sbondCo/Watcharr/database/dbmodel"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/job"
)

type TraktImportRequest struct {
	// Username of public trakt user to import from.
	Username string `json:"username" binding:"required"`
	// An optional custom api key to use for the requests.
	ApiKey string `json:"apiKey"`
}

type TraktUser struct {
	Username string `json:"username"`
	Private  bool   `json:"private"`
	IDs      struct {
		Slug string `json:"slug"`
	} `json:"ids"`
}

type TraktHistory struct {
	ID        int64            `json:"id"`
	WatchedAt time.Time        `json:"watched_at"`
	Action    string           `json:"action"`
	Type      string           `json:"type"`
	Show      TraktListShow    `json:"show,omitempty"`
	Episode   TraktListEpisode `json:"episode,omitempty"`
	Movie     TraktListMovie   `json:"movie,omitempty"`
}

type TraktWatchlist struct {
	Rank     int              `json:"rank"`
	ID       int              `json:"id"`
	ListedAt time.Time        `json:"listed_at"`
	Notes    string           `json:"notes"`
	Type     string           `json:"type"`
	Show     TraktListShow    `json:"show,omitempty"`
	Episode  TraktListEpisode `json:"episode,omitempty"`
	Movie    TraktListMovie   `json:"movie,omitempty"`
}

type TraktRating struct {
	Rating  int              `json:"rating"`
	Type    string           `json:"type"`
	Show    TraktListShow    `json:"show,omitempty"`
	Episode TraktListEpisode `json:"episode,omitempty"`
	Movie   TraktListMovie   `json:"movie,omitempty"`
}

type TraktListShow struct {
	Title string `json:"title"`
	Year  int    `json:"year"`
	Ids   struct {
		Trakt int    `json:"trakt"`
		Slug  string `json:"slug"`
		Tmdb  int    `json:"tmdb"`
	} `json:"ids"`
}

type TraktListEpisode struct {
	Season int    `json:"season"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	Ids    struct {
		Trakt int    `json:"trakt"`
		Slug  string `json:"slug"`
		Tmdb  int    `json:"tmdb"`
	} `json:"ids"`
}

type TraktListMovie struct {
	Title string `json:"title"`
	Year  int    `json:"year"`
	Ids   struct {
		Trakt int    `json:"trakt"`
		Slug  string `json:"slug"`
		Tmdb  int    `json:"tmdb"`
	} `json:"ids"`
}

type TraktImportResponse struct {
	JobId string `json:"jobId"`
}

type TraktService struct {
	s *Service
}

func NewTraktService(s *Service) *TraktService {
	return &TraktService{
		s,
	}
}

type trackImporter struct {
	toImport     map[string]domain.ImportRequest
	reportError  func(err string) error
	reportUpdate func(update string) error
}

func (i *trackImporter) addHistory(t *TraktService, v TraktHistory) error {
	var collectingText string
	switch v.Type {
	case "episode":
		collectingText = fmt.Sprintf("%s S%dE%d", v.Show.Title, v.Episode.Season, v.Episode.Number)
	case "show":
		collectingText = v.Show.Title
	case "movie":
		collectingText = v.Movie.Title
	}
	if collectingText != "" {
		i.reportUpdate("collecting " + collectingText)
	}
	return t.processTraktHistoryItem(v, i.toImport)
}

func (i *trackImporter) addWatchlist(t *TraktService, v TraktWatchlist) {
	var (
		title       string
		contentType domain.ImportContentType
		tmdbId      int
	)
	switch v.Type {
	case "show", "episode":
		title = v.Show.Title
		tmdbId = v.Show.Ids.Tmdb
		contentType = domain.ImportContentTypeShow
		if v.Type == "episode" {
			title = v.Episode.Title
		}
	case "movie":
		title = v.Movie.Title
		tmdbId = v.Movie.Ids.Tmdb
		contentType = domain.ImportContentTypeMovie
	}
	i.reportUpdate("setting status for " + title)
	mapKey := t.makeTraktMapKey(contentType, tmdbId)
	if mv, ok := i.toImport[mapKey]; ok {
		// If item already exists in toImport, set its status to planned.
		if v.Type == "episode" {
			// For episode entries, we have to find the WatchedEpisode to set its status to planned.
			weFound := false
			for i, we := range mv.WatchedEpisodes {
				if we.SeasonNumber == v.Episode.Season && we.EpisodeNumber == v.Episode.Number {
					we.Status = entity.PLANNED
					mv.WatchedEpisodes[i] = we
					weFound = true
					break
				}
			}
			if !weFound {
				mv.WatchedEpisodes = append(mv.WatchedEpisodes, entity.WatchedEpisode{
					SeasonNumber:  v.Episode.Season,
					EpisodeNumber: v.Episode.Number,
					Status:        entity.PLANNED,
					GormModelNoDel: dbmodel.GormModelNoDel{
						CreatedAt: v.ListedAt,
					},
				})
			}
			i.toImport[mapKey] = mv
		} else {
			mv.Status = entity.PLANNED
			if v.Notes != "" {
				// episodes dont support notes in watcharr
				mv.Thoughts = v.Notes
			}
			i.toImport[mapKey] = mv
		}
	} else {
		// If the item does not exist in toImport, create it and set it to planned.
		ti := domain.ImportRequest{
			Type:   contentType,
			TmdbID: tmdbId,
			Status: entity.PLANNED,
		}
		if v.Type == "episode" {
			ti.WatchedEpisodes = []entity.WatchedEpisode{{
				SeasonNumber:  v.Episode.Season,
				EpisodeNumber: v.Episode.Number,
				Status:        entity.PLANNED,
				GormModelNoDel: dbmodel.GormModelNoDel{
					CreatedAt: v.ListedAt,
				},
			}}
		} else {
			// episodes dont support notes in watcharr
			ti.Thoughts = v.Notes
		}
		i.toImport[mapKey] = ti
	}
}

func (i *trackImporter) addRating(t *TraktService, v TraktRating) {
	var (
		title       string
		contentType domain.ImportContentType
		tmdbId      int
		traktSlug   string
	)
	switch v.Type {
	case "show", "episode":
		title = v.Show.Title
		tmdbId = v.Show.Ids.Tmdb
		traktSlug = v.Show.Ids.Slug
		contentType = domain.ImportContentTypeShow
		if v.Type == "episode" {
			title = v.Episode.Title
			traktSlug = v.Episode.Ids.Slug
		}
	case "movie":
		title = v.Movie.Title
		tmdbId = v.Movie.Ids.Tmdb
		contentType = domain.ImportContentTypeMovie
		traktSlug = v.Movie.Ids.Slug
	}
	i.reportUpdate(fmt.Sprintf("setting rating of %d for %s", v.Rating, title))
	mapKey := t.makeTraktMapKey(contentType, tmdbId)
	if mv, ok := i.toImport[mapKey]; ok {
		if v.Type == "episode" {
			// For episode entries, we have to find the WatchedEpisode to set its rating.
			epFound := false
			for i, we := range mv.WatchedEpisodes {
				if we.SeasonNumber == v.Episode.Season && we.EpisodeNumber == v.Episode.Number {
					we.Rating = int8(v.Rating)
					mv.WatchedEpisodes[i] = we
					epFound = true
					break
				}
			}
			i.toImport[mapKey] = mv
			if !epFound {
				i.reportError(fmt.Sprintf("episode rating of %d for %s not imported. The episode does not exist in your history or watchlist.", v.Rating, title))
			}
		} else {
			mv.Rating = float64(v.Rating)
			i.toImport[mapKey] = mv
		}
	} else {
		// Item should be in toImport by now (from history or watchlist) if it has a rating, otherwise we won't import it
		i.reportError(fmt.Sprintf("cannot import rating of %d for %s. The main content does not exist in your history or watchlist. type: %s traktSlug: %s", v.Rating, title, v.Type, traktSlug))
	}
}

func (i *trackImporter) persist(t *TraktService, userId uint) {
	for _, v := range i.toImport {
		_, err := t.s.ImportContent(userId, v)
		if err != nil {
			slog.Error("startTraktImport: Failed to do import on content!", "error", err, "import_obj", v)
			i.reportError(fmt.Sprintf("Failed to import %s as %s. tmdbId: %d", v.Type, v.Status, v.TmdbID))
		}
	}
}

// TODO we could support trakt list imports when we support a similar feature (tags will function as custom lists when done #199)
func (t *TraktService) startTraktImport(jobId string, userId uint, req TraktImportRequest) {
	importer := trackImporter{
		toImport: map[string]domain.ImportRequest{},
		reportError: func(err string) error {
			return job.AddJobError(jobId, userId, err)
		},
		reportUpdate: func(update string) error {
			return job.UpdateJobCurrentTask(jobId, userId, update)
		},
	}

	// Get trakt user. We want to get their profile `slug` for use in
	// next requests and we can check their profile isn't private while here.
	var traktUser TraktUser
	_, err := t.traktAPIRequest(
		"users/"+req.Username,
		map[string]string{},
		&traktUser,
		req.ApiKey)
	if err != nil {
		slog.Error("startTraktImport: Failed to get users profile",
			"error", err,
			"trakt_user", traktUser)
		job.AddJobError(jobId, userId, "failed to request trakt profile from api")
		job.UpdateJobStatus(jobId, userId, job.JOB_CANCELLED)
		return
	}
	if traktUser.Private {
		slog.Error("startTraktImport: Users profile is private. Cannot continue with import.")
		job.AddJobError(jobId, userId, "trakt profile is private")
		job.UpdateJobStatus(jobId, userId, job.JOB_CANCELLED)
		return
	}
	userSlug := traktUser.IDs.Slug
	// Process all history for this user (in chunks of 1000).
	var history []TraktHistory
	slog.Debug("startTraktImport: Getting first history page")
	historyHeaders, err := t.traktAPIRequest(
		"users/"+userSlug+"/history",
		map[string]string{"limit": "1000"},
		&history,
		req.ApiKey)
	if err != nil {
		// FATAL if we can't get the users history, we probably shouldn't continue
		// (to ratings/watchlist below).
		slog.Error("startTraktImport: Failed to get users history", "error", err)
		job.AddJobError(jobId, userId, "failed to get your history")
		return
	} else {
		pageCount := historyHeaders.Get("x-pagination-page-count")
		slog.Debug("startTraktImport: Got first history page", "page_count", pageCount)
		if pageCount == "" {
			slog.Error("startTraktImport: Failed to get history page count!", "page_count", pageCount)
			job.AddJobError(jobId, userId, "Failed to get history page count")
			return
		}
		pageCountNum, err := strconv.Atoi(pageCount)
		if err != nil {
			slog.Error("startTraktImport: Failed to parse history page count into an int!", "error", err)
			job.AddJobError(jobId, userId, "Failed to parse history page count: "+pageCount)
			return
		}
		rProc := func(v TraktHistory) {
			if err := importer.addHistory(t, v); err != nil {
				job.AddJobError(jobId, userId, err.Error())
			}
		}
		// Process first page of history (next pages processed below)
		for _, v := range history {
			rProc(v)
		}
		for i := range pageCountNum {
			slog.Debug("startTraktImport: Getting history page", "page_num", i)
			_, err := t.traktAPIRequest(
				"users/"+userSlug+"/history",
				map[string]string{"limit": "1000", "page": strconv.Itoa(i)},
				&history,
				req.ApiKey)
			if err != nil {
				slog.Error("startTraktImport: Failed to get a history page", "page_num", i, "error", err)
				job.AddJobError(jobId, userId, "Failed to get history page: "+strconv.Itoa(i))
			} else {
				for _, v := range history {
					rProc(v)
				}
			}
		}
		slog.Info("startTraktImport: Finished processing all history")
		history = nil // clear whatever is lingering in the history slice
	}
	// Get watchlist for PLANNED items
	slog.Info("startTraktImport: Getting whole watchlist")
	var watchlist []TraktWatchlist
	_, err = t.traktAPIRequest(
		"users/"+userSlug+"/watchlist",
		map[string]string{},
		&watchlist,
		req.ApiKey)
	if err != nil {
		slog.Error("startTraktImport: Failed to get users watchlist! Cannot import planned content.", "error", err)
		job.AddJobError(jobId, userId, "failed to get your watchlist (planned items cannot be imported)")
	} else {
		slog.Debug("startTraktImport: Successfully got whole watchlist")
		for _, v := range watchlist {
			slog.Debug("startTraktImport: Processing watchlist item", "item", v)
			importer.addWatchlist(t, v)
		}
	}
	// Process ratings
	slog.Info("startTraktImport: Getting all ratings")
	var ratings []TraktRating
	_, err = t.traktAPIRequest(
		"users/"+userSlug+"/ratings",
		map[string]string{},
		&ratings,
		req.ApiKey)
	if err != nil {
		slog.Error("startTraktImport: Failed to get users ratings!", "error", err)
		job.AddJobError(jobId, userId, "failed to get your ratings (content ratings cannot be imported)")
	} else {
		slog.Debug("startTraktImport: Successfully got all ratings")
		for _, v := range ratings {
			slog.Debug("startTraktImport: Processing rating item", "item", v)
			importer.addRating(t, v)
		}
	}

	importer.persist(t, userId)

	// We are donezo
	job.UpdateJobStatus(jobId, userId, job.JOB_DONE)
}

func (t *TraktService) processTraktHistoryItem(v TraktHistory, toImport map[string]domain.ImportRequest) error {
	var (
		title          string
		traktId        int
		tmdbId         int
		contentType    domain.ImportContentType
		watchedEpisode entity.WatchedEpisode
	)
	switch v.Type {
	case "show", "episode":
		title = v.Show.Title
		traktId = v.Show.Ids.Trakt
		tmdbId = v.Show.Ids.Tmdb
		contentType = domain.ImportContentTypeShow
		if v.Type == "episode" {
			traktId = v.Episode.Ids.Trakt
			watchedEpisode = entity.WatchedEpisode{
				SeasonNumber:  v.Episode.Season,
				EpisodeNumber: v.Episode.Number,
				Status:        entity.FINISHED,
				// Rating: ,
				GormModelNoDel: dbmodel.GormModelNoDel{
					CreatedAt: v.WatchedAt,
				},
			}
			slog.Debug("processTraktHistoryItem: Processing an episode.", "showTitle", title, "season", v.Episode.Season, "episode", v.Episode.Number)
		} else {
			slog.Debug("processTraktHistoryItem: Processing a show.", "contentTitle", title, "contentTmdbId", tmdbId)
		}
	case "movie":
		title = v.Movie.Title
		traktId = v.Movie.Ids.Trakt
		tmdbId = v.Movie.Ids.Tmdb
		contentType = domain.ImportContentTypeMovie
		slog.Debug("processTraktHistoryItem: Processing a movie.", "contentTitle", title, "contentTmdbId", tmdbId)
	}
	if tmdbId == 0 {
		slog.Debug("processTraktHistoryItem: Item had no tmdbId. Cannot process.")
		return errors.New("Failed to process history: " + title + " type:" + v.Type + " trakt id:" + strconv.Itoa(traktId) + " tmdb id:" + strconv.Itoa(tmdbId) + " error:" + "item had no tmdb id")
	}
	mapKey := t.makeTraktMapKey(contentType, tmdbId)
	if e, ok := toImport[mapKey]; ok {
		e.WatchedEpisodes = append(toImport[mapKey].WatchedEpisodes, watchedEpisode)
		toImport[mapKey] = e
	} else {
		toImport[mapKey] = domain.ImportRequest{
			Type:            contentType,
			TmdbID:          tmdbId,
			Status:          entity.FINISHED,
			DatesWatched:    []time.Time{v.WatchedAt},
			WatchedEpisodes: []entity.WatchedEpisode{watchedEpisode},
		}
	}
	return nil
}

// `tmdbId` is for the movie or show (not for episodes).
func (t *TraktService) makeTraktMapKey(ct domain.ImportContentType, tmdbId int) string {
	return string(ct) + strconv.Itoa(tmdbId)
}

func (t *TraktService) traktAPIRequest(
	ep string,
	p map[string]string,
	resp interface{},
	// If not provided, a default key is used
	apiKey string,
) (http.Header, error) {
	base, err := url.Parse("https://api.trakt.tv")
	if err != nil {
		return map[string][]string{}, errors.New("failed to parse api uri")
	}
	base.Path += ep
	if len(p) > 0 {
		params := url.Values{}
		for k, v := range p {
			params.Add(k, v)
		}
		base.RawQuery = params.Encode()
	}
	slog.Debug("traktAPIRequest", "request_url", base.String())
	req, err := http.NewRequest("GET", base.String(), nil)
	if err != nil {
		return map[string][]string{}, err
	}
	traktApiKey := "1309b3a473f6718ba0586eddf4b8caccd8733ec74ac47aa6eadc23331d9c7ab4"
	if apiKey != "" {
		traktApiKey = apiKey
	}
	// trakt-api-key is your Trakt Apps 'Client ID'.
	req.Header.Add("trakt-api-key", traktApiKey)
	req.Header.Add("trakt-api-version", "2")
	req.Header.Add("Content-type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return map[string][]string{}, err
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return map[string][]string{}, err
	}
	if !(res.StatusCode >= 200 && res.StatusCode <= 299) {
		slog.Error("traktAPIRequest: non 2xx status code:", "status_code", res.StatusCode)
		return map[string][]string{}, errors.New("non success status code")
	}
	err = json.Unmarshal([]byte(body), &resp)
	if err != nil {
		return map[string][]string{}, err
	}
	return res.Header, nil
}

func (t *TraktService) TraktImportWatched(
	userId uint,
	req TraktImportRequest,
) (TraktImportResponse, error) {
	jobId, err := job.AddUniqueJob("trakt_import", userId)
	if err != nil {
		slog.Error("TraktImportWatched: Failed to create a job",
			"error", err)
		return TraktImportResponse{}, err
	}

	job.UpdateJobStatus(jobId, userId, job.JOB_RUNNING)

	if req.ApiKey != "" {
		slog.Info("TraktImportWatched: A custom api key was provided for this import.")
	}

	go t.startTraktImport(
		jobId,
		userId,
		req,
	)

	return TraktImportResponse{JobId: jobId}, nil
}

func (t *TraktService) TraktImportWatchedFromZip(
	userId uint,
	file io.ReaderAt,
	size int64,
) error {
	errors := make([]string, 0)

	importer := trackImporter{
		toImport: map[string]domain.ImportRequest{},
		reportError: func(err string) error {
			slog.Error("error while importing trakt data", slog.String("error", err), slog.Uint64("userId", uint64(userId)))
			errors = append(errors, err)
			return nil
		},
		reportUpdate: func(update string) error {
			slog.Info("trakt data import update", slog.String("update", update), slog.Uint64("userId", uint64(userId)))
			return nil
		},
	}

	reader, err := zip.NewReader(file, size)
	if err != nil {
		return fmt.Errorf("failed opening ZIP file: %w", err)
	}

	var watchedHistoryItems []TraktHistory
	var watchlistItems []TraktWatchlist
	var ratingItems []TraktRating

	for _, file := range reader.File {
		if strings.HasPrefix(file.Name, "watched-history-") {
			f, err := reader.Open(file.Name)
			if err != nil {
				return fmt.Errorf("failed opening '%s': %w", file.Name, err)
			}
			data, err := io.ReadAll(f)
			if err != nil {
				return fmt.Errorf("failed reading '%s': %w", file.Name, err)
			}

			var items []TraktHistory
			if err := json.Unmarshal(data, &items); err != nil {
				return fmt.Errorf("failed parsing JSON for '%s': %w", file.Name, err)
			}

			for _, item := range items {
				watchedHistoryItems = append(watchedHistoryItems, item)
			}
		}

		if file.Name == "lists-watchlist.json" {
			f, err := reader.Open(file.Name)
			if err != nil {
				return fmt.Errorf("failed opening '%s': %w", f, err)
			}
			data, err := io.ReadAll(f)
			if err != nil {
				return fmt.Errorf("failed reading '%s': %w", f, err)
			}

			var items []TraktWatchlist
			if err := json.Unmarshal(data, &items); err != nil {
				return fmt.Errorf("failed parsing JSON for '%s': %w", f, err)
			}

			for _, item := range items {
				watchlistItems = append(watchlistItems, item)
			}
		}

		if strings.HasPrefix(file.Name, "ratings-") {
			f, err := reader.Open(file.Name)
			if err != nil {
				return fmt.Errorf("failed opening '%s': %w", f, err)
			}
			data, err := io.ReadAll(f)
			if err != nil {
				return fmt.Errorf("failed reading '%s': %w", f, err)
			}

			var items []TraktRating
			if err := json.Unmarshal(data, &items); err != nil {
				return fmt.Errorf("failed parsing JSON for '%s': %w", f, err)
			}

			for _, item := range items {
				ratingItems = append(ratingItems, item)
			}
		}
	}

	for _, item := range watchedHistoryItems {
		importer.addHistory(t, item)
	}

	for _, item := range watchlistItems {
		importer.addWatchlist(t, item)
	}

	for _, item := range ratingItems {
		importer.addRating(t, item)
	}

	importer.persist(t, userId)

	return nil
}
