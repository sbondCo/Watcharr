// Trakt.tv importer.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type TraktImportRequest struct {
	// Username of public trakt user to import from.
	Username string `json:"username" binding:"required"`
}

type TraktUser struct {
	Username string `json:"username"`
	Private  bool   `json:"private"`
	IDs      struct {
		Slug string `json:"slug"`
	} `json:"ids"`
}

type TraktHistory struct {
	ID        int64     `json:"id"`
	WatchedAt time.Time `json:"watched_at"`
	Action    string    `json:"action"`
	Type      string    `json:"type"`
	Movie     struct {
		Title string `json:"title"`
		Year  int    `json:"year"`
		Ids   struct {
			Trakt int    `json:"trakt"`
			Slug  string `json:"slug"`
			Tmdb  int    `json:"tmdb"`
		} `json:"ids"`
	} `json:"movie,omitempty"`
	Episode struct {
		Season int    `json:"season"`
		Number int    `json:"number"`
		Title  string `json:"title"`
		Ids    struct {
			Trakt int `json:"trakt"`
			Tmdb  int `json:"tmdb"`
		} `json:"ids"`
	} `json:"episode,omitempty"`
	Show struct {
		Title string `json:"title"`
		Year  int    `json:"year"`
		Ids   struct {
			Trakt int    `json:"trakt"`
			Slug  string `json:"slug"`
			Tmdb  int    `json:"tmdb"`
		} `json:"ids"`
	} `json:"show,omitempty"`
}

type TraktImportResponse struct {
	JobId string `json:"jobId"`
}

// TODO we could support trakt list imports when we support a similar feature (tags will function as custom lists when done #199)
func startTraktImport(db *gorm.DB, jobId string, userId uint, traktUsername string) {
	// Get trakt user. We want to get their profile `slug` for use in
	// next requests and we can check their profile isn't private while here.
	var traktUser TraktUser
	_, err := traktAPIRequest("users/"+traktUsername, map[string]string{}, &traktUser)
	if err != nil {
		slog.Error("startTraktImport: Failed to get users profile", "error", err, "trakt_user", traktUser)
		addJobError(jobId, userId, "failed to request trakt profile from api")
		updateJobStatus(jobId, userId, JOB_CANCELLED)
		return
	}
	if traktUser.Private {
		slog.Error("startTraktImport: Users profile is private. Cannot continue with import.")
		addJobError(jobId, userId, "trakt profile is private")
		updateJobStatus(jobId, userId, JOB_CANCELLED)
		return
	}
	userSlug := traktUser.IDs.Slug
	// Everything will be added to this map for importing at the end.
	toImport := map[string]ImportRequest{}
	// Process all history for this user (in chunks of 1000).
	var history []TraktHistory
	slog.Debug("startTraktImport: Getting first history page")
	historyHeaders, err := traktAPIRequest("users/"+userSlug+"/history", map[string]string{"limit": "1000"}, &history)
	if err != nil {
		slog.Error("startTraktImport: Failed to get users history", "error", err)
		addJobError(jobId, userId, "failed to get your history")
	} else {
		pageCount := historyHeaders.Get("x-pagination-page-count")
		slog.Debug("startTraktImport: Got first history page", "page_count", pageCount)
		if pageCount == "" {
			slog.Error("startTraktImport: Failed to get history page count!", "page_count", pageCount)
			addJobError(jobId, userId, "Failed to get history page count")
			return
		}
		pageCountNum, err := strconv.Atoi(pageCount)
		if err != nil {
			slog.Error("startTraktImport: Failed to parse history page count into an int!", "error", err)
			addJobError(jobId, userId, "Failed to parse history page count: "+pageCount)
			return
		}
		rProc := func(v TraktHistory) {
			var collectingText string
			if v.Type == "episode" {
				collectingText = fmt.Sprintf("%s S%dE%d", v.Show.Title, v.Episode.Season, v.Episode.Number)
			} else if v.Type == "show" {
				collectingText = v.Show.Title
			} else if v.Type == "movie" {
				collectingText = v.Movie.Title
			}
			if collectingText != "" {
				updateJobCurrentTask(jobId, userId, "collecting "+collectingText)
			}
			err = processTraktHistoryItem(v, toImport)
			if err != nil {
				addJobError(jobId, userId, err.Error())
			}
		}
		// Process first page of history (next pages processed below)
		for _, v := range history {
			rProc(v)
		}
		for i := range pageCountNum {
			slog.Debug("startTraktImport: Getting history page", "page_num", i)
			_, err := traktAPIRequest("users/"+userSlug+"/history", map[string]string{"limit": "1000", "page": strconv.Itoa(i)}, &history)
			if err != nil {
				slog.Error("startTraktImport: Failed to get a history page", "page_num", i, "error", err)
				addJobError(jobId, userId, "Failed to get history page: "+strconv.Itoa(i))
			} else {
				for _, v := range history {
					rProc(v)
				}
			}
		}
		slog.Debug("startTraktImport: toImport:", "toimport", toImport)
	}
	// Get watchlist for PLANNED items
	// Process ratings
}

func processTraktHistoryItem(v TraktHistory, toImport map[string]ImportRequest) error {
	var (
		title          string
		traktId        int
		tmdbId         int
		contentType    ContentType
		watchedEpisode WatchedEpisode
	)
	if v.Type == "show" || v.Type == "episode" {
		title = v.Show.Title
		traktId = v.Show.Ids.Trakt
		tmdbId = v.Show.Ids.Tmdb
		contentType = SHOW
		if v.Type == "episode" {
			traktId = v.Episode.Ids.Trakt
			watchedEpisode = WatchedEpisode{
				SeasonNumber:  v.Episode.Season,
				EpisodeNumber: v.Episode.Number,
				Status:        FINISHED,
				// Rating: ,
				GormModel: GormModel{
					CreatedAt: v.WatchedAt,
				},
			}
			slog.Debug("processTraktHistoryItem: Processing an episode.", "showTitle", title, "season", v.Episode.Season, "episode", v.Episode.Number)
		} else {
			slog.Debug("processTraktHistoryItem: Processing a show.", "contentTitle", title, "contentTmdbId", tmdbId)
		}
	} else if v.Type == "movie" {
		title = v.Movie.Title
		traktId = v.Movie.Ids.Trakt
		tmdbId = v.Movie.Ids.Tmdb
		contentType = MOVIE
		slog.Debug("processTraktHistoryItem: Processing a movie.", "contentTitle", title, "contentTmdbId", tmdbId)
	}
	if tmdbId == 0 {
		slog.Debug("processTraktHistoryItem: Item had no tmdbId. Cannot process.")
		return errors.New("Failed to process history: " + title + " type:" + v.Type + " trakt id:" + strconv.Itoa(traktId) + " tmdb id:" + strconv.Itoa(tmdbId) + " error:" + "item had no tmdb id")
	}
	mapKey := string(contentType) + strconv.Itoa(tmdbId)
	if e, ok := toImport[mapKey]; ok {
		e.WatchedEpisodes = append(toImport[mapKey].WatchedEpisodes, watchedEpisode)
		toImport[mapKey] = e
	} else {
		toImport[mapKey] = ImportRequest{
			Type:            contentType,
			TmdbID:          tmdbId,
			Status:          FINISHED,
			DatesWatched:    []time.Time{v.WatchedAt},
			WatchedEpisodes: []WatchedEpisode{watchedEpisode},
		}
	}
	return nil
}

func traktAPIRequest(ep string, p map[string]string, resp interface{}) (http.Header, error) {
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
	req.Header.Add("trakt-api-key", "c481cb044dcd58d83f3fde113741d1e28d19c1bef1bcbfcb9acedee222f3a673")
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

func traktImportWatched(
	db *gorm.DB,
	userId uint,
	traktUsername string,
) (TraktImportResponse, error) {
	jobId, err := addJob("trakt_import", userId)
	if err != nil {
		slog.Error("traktSyncWatched: Failed to create a job", "error", err)
		return TraktImportResponse{}, errors.New("failed to create job")
	}

	updateJobStatus(jobId, userId, JOB_RUNNING)

	go startTraktImport(
		db,
		jobId,
		userId,
		traktUsername,
	)

	return TraktImportResponse{JobId: jobId}, nil
}
