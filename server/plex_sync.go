package main

import (
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type PlexSyncResponse struct {
	JobId string `json:"jobId"`
}

// Perform a Plex sync.
// Errors are added silently to the job.
func startPlexSync(
	db *gorm.DB,
	jobId string,
	userId uint,
	userThirdPartyAuth string,
) {
	updateJobCurrentTask(jobId, userId, "fetching libraries")
	libraries, err := getPlexLibraries(userThirdPartyAuth)
	if err != nil {
		slog.Error("plexSyncWatched: Failed to fetch libraries", "user_id", userId, "error", err)
		addJobError(jobId, userId, "failed to get plex libraries")
		updateJobStatus(jobId, userId, JOB_DONE)
		return
	}
	for _, library := range libraries.MediaContainer.Directory {
		slog.Debug("plexSyncWatched: Processing a library", "library_title", library.Title, "library_type", library.Type, "user_id", userId)
		if library.Type == "movie" {
			updateJobCurrentTask(jobId, userId, "importing movies from "+library.Title)
			movies, err := getPlexLibraryItems(userThirdPartyAuth, library.Key)
			if err != nil {
				slog.Error("plexSyncWatched: Failed to fetch movies from library", "library", library.Key, "user_id", userId, "error", err)
				addJobError(jobId, userId, "failed to fetch movies from library "+library.Key)
				continue
			}
			for _, movie := range movies.MediaContainer.Metadata {
				if movie.ViewCount == 0 {
					// Not viewed and not rated, skip importing
					slog.Debug("plexSyncWatched: Skipping unwatched movie:", "movie_name", movie.Title, "user_id", userId)
					continue
				}
				updateJobCurrentTask(jobId, userId, "importing movie "+movie.Title)
				slog.Info("plexSyncWatched: Importing movie.", "movie_name", movie.Title, "user_id", userId)

				// Find tmdb id
				if len(movie.Guid) <= 0 {
					slog.Error("plexSyncWatched: Movie to import does not have any external guids.", "movie_name", movie.Title, "movie_id", movie.GUID, "user_id", userId)
					addJobError(jobId, userId, "movie could not be imported (no external ids present): "+movie.Title)
					continue
				}
				tmdbIdStr := ""
				for _, v := range movie.Guid {
					if strings.HasPrefix(v.ID, "tmdb://") {
						tmdbIdStr = v.ID[7:]
						break
					}
				}
				if tmdbIdStr == "" {
					slog.Error("plexSyncWatched: Movie to import does not have a tmdb id.", "movie_name", movie.Title, "tmdb_id_str", tmdbIdStr, "movie_id", movie.GUID, "user_id", userId)
					addJobError(jobId, userId, "movie could not be imported (no tmdbId present): "+movie.Title)
					continue
				}
				tmdbId, err := strconv.Atoi(tmdbIdStr)
				if err != nil {
					slog.Error("plexSyncWatched: Movie to import does not have a parseable (to int) tmdb id.", "movie_name", movie.Title, "tmdb_id_str", tmdbIdStr, "movie_id", movie.GUID, "user_id", userId)
					addJobError(jobId, userId, "movie could not be imported (tmdbId was not parseable): "+movie.Title)
					continue
				}

				lastViewedAt := time.Unix(movie.LastViewedAt, 0)
				w, err := addWatched(db, userId, WatchedAddRequest{
					Status:      FINISHED,
					ContentID:   tmdbId,
					ContentType: MOVIE,
					Rating:      int8(movie.UserRating),
					WatchedDate: lastViewedAt,
				}, IMPORTED_WATCHED_PLEX)
				if err != nil {
					if err.Error() == "content already on watched list" {
						slog.Error("plexSyncWatched: unique constraint hit. movie must already be on watch list", "error", err)
						continue
					}
					slog.Error("plexSyncWatched: Failed to add movie as watched", "error", err)
					addJobError(jobId, userId, "failed to add movie "+movie.Title)
				} else {
					// 3. Add IMPORTED_ADDED_WATCHED_PLEX activity
					if !lastViewedAt.IsZero() {
						_, err := addActivity(db, userId, ActivityAddRequest{
							WatchedID:  w.ID,
							Type:       IMPORTED_ADDED_WATCHED_PLEX,
							CustomDate: &lastViewedAt,
						})
						if err != nil {
							slog.Error("plexSyncWatched: Failed to add dateswatched activity.", "movie_name", movie.Title,
								"movie_id", movie.GUID, "user_id", userId, "date", lastViewedAt, "unparsed_date", movie.LastViewedAt, "error", err)
						}
					}
				}
			}
		} else if library.Type == "show" {
			updateJobCurrentTask(jobId, userId, "importing tv shows from "+library.Title)
			shows, err := getPlexLibraryItems(userThirdPartyAuth, library.Key)
			if err != nil {
				slog.Error("plexSyncWatched: Failed to fetch shows from library", "library", library.Key, "error", err)
				addJobError(jobId, userId, "failed to fetch shows from library "+library.Key)
				continue
			}
			for _, show := range shows.MediaContainer.Metadata {
				if show.ViewedLeafCount != show.LeafCount {
					// Not viewed, skip importing
					// (could be improved to set status as watching when viewedLeafCount is higher than 0)
					slog.Debug("plexSyncWatched: Skipping unwatched show:", "show_name", show.Title, "leaf_count", show.LeafCount, "viewed_leaf_count", show.ViewedLeafCount, "user_id", userId)
					continue
				}
				updateJobCurrentTask(jobId, userId, "importing show "+show.Title)
				slog.Info("plexSyncWatched: Importing show.", "show_name", show.Title, "user_id", userId)

				tmdbIdStr := ""
				for _, v := range show.Guid {
					if strings.HasPrefix(v.ID, "tmdb://") {
						tmdbIdStr = v.ID[7:]
						break
					}
				}
				if tmdbIdStr == "" {
					slog.Error("plexSyncWatched: Show to import does not have a tmdb id.", "show_name", show.Title, "tmdb_id_str", tmdbIdStr, "show_id", show.GUID, "user_id", userId)
					addJobError(jobId, userId, "movie could not be imported (no tmdbId present): "+show.Title)
					continue
				}
				tmdbId, err := strconv.Atoi(tmdbIdStr)
				if err != nil {
					slog.Error("plexSyncWatched: Show to import does not have a parseable (to int) tmdb id.", "show_name", show.Title, "tmdb_id_str", tmdbIdStr, "show_id", show.GUID, "user_id", userId)
					addJobError(jobId, userId, "show could not be imported (tmdbId was not parseable): "+show.Title)
					continue
				}

				lastViewedAt := time.Unix(show.LastViewedAt, 0)
				w, err := addWatched(db, userId, WatchedAddRequest{
					Status:      FINISHED,
					ContentID:   tmdbId,
					ContentType: SHOW,
					Rating:      int8(show.UserRating),
					WatchedDate: lastViewedAt,
				}, IMPORTED_WATCHED_PLEX)
				if err != nil {
					if err.Error() == "content already on watched list" {
						slog.Error("plexSyncWatched: unique constraint hit. show must already be on watch list", "error", err)
						continue
					}
					slog.Error("plexSyncWatched: Failed to add show as watched", "error", err)
					addJobError(jobId, userId, "failed to add show "+show.Title)
				} else {
					// 3. Add IMPORTED_ADDED_WATCHED_PLEX activity
					if !lastViewedAt.IsZero() {
						_, err := addActivity(db, userId, ActivityAddRequest{
							WatchedID:  w.ID,
							Type:       IMPORTED_ADDED_WATCHED_PLEX,
							CustomDate: &lastViewedAt,
						})
						if err != nil {
							slog.Error("plexSyncWatched: Failed to add dateswatched activity.", "movie_name", show.Title,
								"movie_id", show.GUID, "user_id", userId, "date", lastViewedAt, "unparsed_date", show.LastViewedAt, "error", err)
						}
					}
				}
			}
		}
	}
	updateJobStatus(jobId, userId, JOB_DONE)
}

func plexSyncWatched(
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

	go startPlexSync(
		db,
		jobId,
		userId,
		userThirdPartyAuth,
	)

	return PlexSyncResponse{JobId: jobId}, nil
}
