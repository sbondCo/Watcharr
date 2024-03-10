package main

import (
	"errors"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

type PlexSyncResponse struct {
	JobId string `json:"jobId"`
}

// Perform a Plex sync.
// Errors are added silently to the job.
func plexSyncWatched(
	db *gorm.DB,
	jobId string,
	userId uint,
	userThirdPartyAuth string,
) {
	updateJobCurrentTask(jobId, userId, "fetching libraries")
	libraries, err := plexGetLibraries(userThirdPartyAuth)
	if err != nil {
		addJobError(jobId, userId, "failed to get plex libraries")
		updateJobStatus(jobId, userId, JOB_DONE)
		return
	}
	for _, library := range libraries {
		if library.Type == string(MOVIES) {
			updateJobCurrentTask(jobId, userId, "importing movies")
			movies, err := plexGetLibraryItems(userThirdPartyAuth, library.Key)
			if err != nil {
				slog.Error("Failed to fetch movies from library", "library", library.Key, "error", err)
				addJobError(jobId, userId, "failed to fetch movies from library "+library.Key)
				continue
			}
			for _, movie := range movies {
				if movie.ViewCount == 0 && movie.UserRating == 0 {
					// Not viewed and not rated, skip importing
					continue
				}
				updateJobCurrentTask(jobId, userId, "importing movie "+movie.Title)
				slog.Info("plexSyncWatched: Importing movie.", "movie_name", movie.Title, "user_id", userId)
				sr, err := searchMovie(movie.Title, string(movie.Year))
				if err != nil {
					slog.Error("Failed to search for movie", "movie", movie.Title, "error", err)
					addJobError(jobId, userId, "failed to search for movie "+movie.Title)
					continue
				}

				if len(sr.Results) == 0 {
					slog.Error("plexSyncWatched: could not find movie", "movie", movie.Title)
					addJobError(jobId, userId, "failed to add movie "+movie.Title)
					continue
				}
				// We have to assume it's the first result. Thank you Plex, for not giving better metadata
				// In theory, we could grab the IMDB ID and cross reference it with TMDB's /movie/external_ids API,
				// but that seems like a lot of overhead. Generally title + year is good enough.
				match := sr.Results[0]
				_, err = addWatched(db, userId, WatchedAddRequest{
					Status:      FINISHED,
					ContentID:   match.ID,
					ContentType: MOVIE,
					Rating:      int8(movie.UserRating),
					WatchedDate: time.Unix(int64(movie.LastViewedAt), 0),
				}, IMPORTED_WATCHED_PLEX)
				if err != nil {
					if err.Error() == "content already on watched list" {
						slog.Error("plexSyncWatched: unique constraint hit. movie must already be on watch list", "error", err)
						continue
					}
					slog.Error("plexSyncWatched: Failed to add movie as watched", "error", err)
					addJobError(jobId, userId, "failed to add movie "+movie.Title)
				}
			}
		} else if library.Type == string(TV_SHOWS) {
			updateJobCurrentTask(jobId, userId, "importing tv shows")
			shows, err := plexGetLibraryItems(userThirdPartyAuth, library.Key)
			if err != nil {
				slog.Error("Failed to fetch shows from library", "library", library.Key, "error", err)
				addJobError(jobId, userId, "failed to fetch shows from library "+library.Key)
				continue
			}
			for _, show := range shows {
				if show.ViewCount == 0 && show.UserRating == 0 {
					// Not viewed and not rated, skip importing
					continue
				}
				updateJobCurrentTask(jobId, userId, "importing show "+show.Title)
				slog.Info("plexSyncWatched: Importing show.", "show_name", show.Title, "user_id", userId)
				sr, err := searchTvShow(show.Title, string(show.Year))
				if err != nil {
					slog.Error("Failed to search for show", "show", show.Title, "error", err)
					addJobError(jobId, userId, "failed to search for show "+show.Title)
					continue
				}

				if len(sr.Results) == 0 {
					slog.Error("plexSyncWatched: could not find show", "show", show.Title)
					addJobError(jobId, userId, "failed to add show "+show.Title)
					continue
				}

				match := sr.Results[0]
				_, err = addWatched(db, userId, WatchedAddRequest{
					Status:      FINISHED,
					ContentID:   match.ID,
					ContentType: SHOW,
					Rating:      int8(show.UserRating),
					WatchedDate: time.Unix(int64(show.LastViewedAt), 0),
				}, IMPORTED_WATCHED_PLEX)
				if err != nil {
					if err.Error() == "content already on watched list" {
						slog.Error("plexSyncWatched: unique constraint hit. show must already be on watch list", "error", err)
						continue
					}
					slog.Error("plexSyncWatched: Failed to add show as watched", "error", err)
					addJobError(jobId, userId, "failed to add show "+show.Title)
				}
			}
		}
	}
	updateJobStatus(jobId, userId, JOB_DONE)
}

func startPlexSync(
	db *gorm.DB,
	userId uint,
	userThirdPartyAuth string,
) (PlexSyncResponse, error) {
	jobId, err := addJob("plex_sync", userId)
	if err != nil {
		slog.Error("startPlexSync: Failed to create a job", "error", err)
		return PlexSyncResponse{}, errors.New("failed to create job")
	}

	updateJobStatus(jobId, userId, JOB_RUNNING)

	go plexSyncWatched(
		db,
		jobId,
		userId,
		userThirdPartyAuth,
	)

	return PlexSyncResponse{JobId: jobId}, nil
}
