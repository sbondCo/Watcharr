package main

import (
	"errors"
	"log/slog"
	"strconv"

	"gorm.io/gorm"
)

type JellyfinSeriesSeasonsResponse struct {
	Items []JellyfinSeriesSeasonItem `json:"Items"`
}

type JellyfinSeriesSeasonItem struct {
	JellyfinItems
	// aka the season number
	IndexNumber int `json:"IndexNumber"`
}

type JellyfinSeriesEpisodesResponse struct {
	Items []JellyfinSeriesEpisodeItem `json:"Items"`
}

type JellyfinSeriesEpisodeItem struct {
	JellyfinItems
	// the episode number
	IndexNumber int `json:"IndexNumber"`
	// the episodes season number
	ParentIndexNumber int `json:"ParentIndexNumber"`
}

type JellyfinSyncResponse struct {
	JobId string `json:"jobId"`
}

// Perform the jellyfin sync.
// Gets each type of media separately from jellyfin and attempts to import them.
// Errors are added silently to the job.
func startJellyfinSync(
	db *gorm.DB,
	jobId string,
	userId uint,
	username string,
	userThirdPartyId string,
	userThirdPartyAuth string,
) {
	// Get played movies
	updateJobCurrentTask(jobId, userId, "syncing movies")
	playedMovies := new(JellyfinItemSearchResponse)
	err := jellyfinAPIRequest(
		"GET",
		"/Users/"+userThirdPartyId+"/Items",
		map[string]string{
			"Filters":          "IsPlayed",
			"IncludeItemTypes": "Movie",
			"Fields":           "ProviderIds",
			"Recursive":        "true",
		},
		username,
		userThirdPartyAuth,
		&playedMovies,
	)
	if err != nil {
		slog.Error("jellyfinSyncWatched: Jellyfin API request failed", "error", err)
		addJobError(jobId, userId, "failed to get jellyfin response for movies")
	} else {
		if len(playedMovies.Items) <= 0 {
			slog.Info("jellyfinSyncWatched: User has no played movies.", "user_id", userId)
		} else {
			for _, v := range playedMovies.Items {
				slog.Info("jellyfinSyncWatched: Importing played movie.", "movie_name", v.Name, "user_id", userId)
				slog.Debug("jellyfinSyncWatched: Importing played movie.", "full_item", v, "user_id", userId)

				// 1. Ensure we have a tmdbId
				if v.ProviderIds.Tmdb == "" {
					slog.Error("jellyfinSyncWatched: Movie to import does not have a tmdb id.", "movie_name", v.Name, "movie_ids", v.ProviderIds, "user_id", userId)
					addJobError(jobId, userId, "movie could not be imported (no tmdbId present): "+v.Name)
					continue
				}
				tmdbId, err := strconv.Atoi(v.ProviderIds.Tmdb)
				if err != nil {
					slog.Error("jellyfinSyncWatched: Movie to import does not have a parseable (to int) tmdb id.", "movie_name", v.Name, "movie_ids", v.ProviderIds, "user_id", userId)
					addJobError(jobId, userId, "movie could not be imported (tmdbId was not parseable): "+v.Name)
					continue
				}

				updateJobCurrentTask(jobId, userId, "syncing "+v.Name)

				// 2. Imported watched movie
				w, err := addWatched(db, userId, WatchedAddRequest{
					Status:      FINISHED,
					ContentID:   tmdbId,
					ContentType: MOVIE,
					WatchedDate: v.UserData.LastPlayedDate,
				}, IMPORTED_WATCHED_JF)
				if err != nil {
					if err.Error() == "content already on watched list" {
						slog.Error("jellyfinSyncWatched: Unique constraint hit.. content must already be on watch list.", "movie_name", v.Name, "movie_ids", v.ProviderIds, "user_id", userId)
					} else {
						slog.Error("jellyfinSyncWatched: Movie failed to import.", "movie_name", v.Name, "movie_ids", v.ProviderIds, "user_id", userId)
						addJobError(jobId, userId, "movie could not be imported (failed when adding to watched list): "+v.Name)
					}
				} else {
					// 3. Add IMPORTED_ADDED_WATCHED_JF activity
					if !v.UserData.LastPlayedDate.IsZero() {
						_, err := addActivity(db, userId, ActivityAddRequest{WatchedID: w.ID, Type: IMPORTED_ADDED_WATCHED_JF, CustomDate: &v.UserData.LastPlayedDate})
						if err != nil {
							slog.Error("jellyfinSyncWatched: Failed to add dateswatched activity.", "movie_name", v.Name,
								"movie_ids", v.ProviderIds, "user_id", userId, "date", v.UserData.LastPlayedDate, "error", err)
						}
					}
				}
			}
		}
	}

	// Get played series
	// Can't rely on IsPlayed filter, since we want to get partially played series too.
	updateJobCurrentTask(jobId, userId, "syncing series")
	allSeries := new(JellyfinItemSearchResponse)
	err = jellyfinAPIRequest(
		"GET",
		"/Users/"+userThirdPartyId+"/Items",
		map[string]string{
			"IncludeItemTypes": "Series",
			"Fields":           "ProviderIds,RecursiveItemCount",
			"Recursive":        "true",
			"IsPlaceHolder":    "false",
		},
		username,
		userThirdPartyAuth,
		&allSeries,
	)
	if err != nil {
		slog.Error("jellyfinSyncWatched: Jellyfin API request failed", "error", err)
		addJobError(jobId, userId, "failed to get jellyfin response for series")
	} else {
		if len(allSeries.Items) <= 0 {
			slog.Info("jellyfinSyncWatched: No series found.", "user_id", userId)
		} else {
			// Import series
			for _, v := range allSeries.Items {
				slog.Info("jellyfinSyncWatched: Processing series.", "series_name", v.Name, "user_id", userId)
				slog.Debug("jellyfinSyncWatched: Processing series.", "full_item", v, "user_id", userId)

				// 1. Make sure show is watched or at least partially watched
				if !v.UserData.Played && v.UserData.PlayedPercentage <= 0 && v.RecursiveItemCount == v.UserData.UnplayedItemCount {
					slog.Debug("jellyfinSyncWatched: Skipping unwatched series:", "series_name", v.Name, "user_id", userId)
					continue
				}

				// 1.1. Ensure we have a tmdbId
				if v.ProviderIds.Tmdb == "" {
					slog.Error("jellyfinSyncWatched: Series to import does not have a tmdb id.", "series_name", v.Name, "series_ids", v.ProviderIds, "user_id", userId)
					addJobError(jobId, userId, "series could not be imported (no tmdbId present): "+v.Name)
					continue
				}
				tmdbId, err := strconv.Atoi(v.ProviderIds.Tmdb)
				if err != nil {
					slog.Error("jellyfinSyncWatched: Series to import does not have a parseable (to int) tmdb id.", "series_name", v.Name, "series_ids", v.ProviderIds, "user_id", userId)
					addJobError(jobId, userId, "series could not be imported (tmdbId was not parseable): "+v.Name)
					continue
				}

				updateJobCurrentTask(jobId, userId, "syncing serie "+v.Name)

				// 2. Imported watched series
				w, err := addWatched(db, userId, WatchedAddRequest{
					Status:      FINISHED,
					ContentID:   tmdbId,
					ContentType: SHOW,
					WatchedDate: v.UserData.LastPlayedDate,
				}, IMPORTED_WATCHED_JF)
				if err != nil {
					if err.Error() == "content already on watched list" {
						slog.Info("jellyfinSyncWatched: Unique constraint hit.. content must already be on watch list.",
							"series_name", v.Name, "series_ids", v.ProviderIds, "user_id", userId, "watched_id", w.ID)
					} else {
						slog.Error("jellyfinSyncWatched: Series failed to import.", "series_name", v.Name, "series_ids", v.ProviderIds, "user_id", userId)
						addJobError(jobId, userId, "series could not be imported (failed when adding to watched list): "+v.Name)
					}
				} else {
					// 3. Add IMPORTED_ADDED_WATCHED activity (only if no err above, show also must not have already been on our list)
					if !v.UserData.LastPlayedDate.IsZero() {
						_, err := addActivity(db, userId, ActivityAddRequest{WatchedID: w.ID, Type: IMPORTED_ADDED_WATCHED_JF, CustomDate: &v.UserData.LastPlayedDate})
						if err != nil {
							slog.Error("jellyfinSyncWatched: Failed to add dateswatched activity.", "series_name", v.Name,
								"series_ids", v.ProviderIds, "user_id", userId, "date", v.UserData.LastPlayedDate, "error", err)
						}
					}
				}

				// 4. Import watched seasons for this serie
				// Get all show seasons (filtering isPlayed doesn't seem to be a thing, so we will have to do that ourselves)
				seriesSeasons := new(JellyfinSeriesSeasonsResponse)
				err = jellyfinAPIRequest(
					"GET",
					"/Shows/"+v.Id+"/Seasons",
					map[string]string{
						"UserId":        userThirdPartyId,
						"Fields":        "ProviderIds",
						"IsPlaceHolder": "false",
					},
					username,
					userThirdPartyAuth,
					&seriesSeasons,
				)
				if err != nil {
					slog.Error("jellyfinSyncWatched: Failed to fetch series seasons.", "series_name", v.Name, "series_ids", v.ProviderIds, "user_id", userId, "error", err)
					addJobError(jobId, userId, "series seasons could not be imported (request failed): "+v.Name)
				} else if len(seriesSeasons.Items) <= 0 {
					slog.Info("jellyfinSyncWatched: Series has no seasons.", "series_name", v.Name, "series_ids", v.ProviderIds, "user_id", userId)
				} else {
					for _, vs := range seriesSeasons.Items {
						slog.Debug("jellyfinSyncWatched: Processing a season.", "full_item", v, "user_id", userId)
						if !vs.UserData.Played {
							slog.Debug("jellyfinSyncWatched: Skipping import of unplayed season.", "series_name", v.Name, "season_num", vs.IndexNumber, "user_id", userId)
							continue
						}
						updateJobCurrentTask(jobId, userId, "syncing "+v.Name+" season "+strconv.Itoa(vs.IndexNumber))
						_, err = addWatchedSeason(db, userId, WatchedSeasonAddRequest{
							WatchedID:       w.ID,
							SeasonNumber:    vs.IndexNumber,
							Status:          FINISHED,
							addActivity:     SEASON_ADDED_JF,
							addActivityDate: vs.UserData.LastPlayedDate,
						})
						if err != nil {
							slog.Error("jellyfinSyncWatched: Failed to fetch series seasons.", "series_name", v.Name, "series_ids", v.ProviderIds, "user_id", userId, "error", err)
							addJobError(jobId, userId, "series season could not be imported (addWatchedSeason request failed): "+v.Name+" season "+strconv.Itoa(vs.IndexNumber))
						}
					}
				}

				// 5. Import watched episodes for this serie
				// Gets all show episodes (filtering isPlayed doesn't seem to be a thing, so we will have to do that ourselves)
				seriesEpisodes := new(JellyfinSeriesEpisodesResponse)
				err = jellyfinAPIRequest(
					"GET",
					"/Shows/"+v.Id+"/Episodes",
					map[string]string{
						"UserId":        userThirdPartyId,
						"Fields":        "ProviderIds",
						"IsPlaceHolder": "false",
					},
					username,
					userThirdPartyAuth,
					&seriesEpisodes,
				)
				if err != nil {
					slog.Error("jellyfinSyncWatched: Failed to fetch series episodes.", "series_name", v.Name, "series_ids", v.ProviderIds, "user_id", userId, "error", err)
					addJobError(jobId, userId, "series episodes could not be imported (request failed): "+v.Name)
				} else if len(seriesEpisodes.Items) <= 0 {
					slog.Info("jellyfinSyncWatched: Series has no episodes.", "series_name", v.Name, "series_ids", v.ProviderIds, "user_id", userId)
				} else {
					for _, vs := range seriesEpisodes.Items {
						slog.Debug("jellyfinSyncWatched: Processing an episode.", "full_item", v, "user_id", userId)
						if !vs.UserData.Played {
							slog.Debug("jellyfinSyncWatched: Skipping import of unplayed episode.", "series_name", v.Name, "season_num", vs.ParentIndexNumber, "episode_num", vs.IndexNumber, "user_id", userId)
							continue
						}
						updateJobCurrentTask(jobId, userId, "syncing "+v.Name+" season "+strconv.Itoa(vs.ParentIndexNumber)+" episode "+strconv.Itoa(vs.IndexNumber))
						_, err = addWatchedEpisodes(db, userId, WatchedEpisodeAddRequest{
							WatchedID:       w.ID,
							SeasonNumber:    vs.ParentIndexNumber,
							EpisodeNumber:   vs.IndexNumber,
							Status:          FINISHED,
							addActivity:     EPISODE_ADDED_JF,
							addActivityDate: vs.UserData.LastPlayedDate,
						})
						if err != nil {
							slog.Error("jellyfinSyncWatched: Failed to import series episode.", "series_name", v.Name, "season_num", vs.ParentIndexNumber, "episode_num", vs.IndexNumber, "user_id", userId, "error", err)
							addJobError(jobId, userId, "series episode could not be imported (addWatchedEpisode request failed): "+v.Name+" "+vs.Name)
						}
					}
				}
			}
		}
	}

	updateJobStatus(jobId, userId, JOB_DONE)
}

func jellyfinSyncWatched(
	db *gorm.DB,
	userId uint,
	userType UserType,
	username string,
	userThirdPartyId string,
	userThirdPartyAuth string,
) (JellyfinSyncResponse, error) {
	jobId, err := addJob("jf_sync", userId)
	if err != nil {
		slog.Error("jellyfinSyncWatched: Failed to create a job", "error", err)
		return JellyfinSyncResponse{}, errors.New("failed to create job")
	}

	updateJobStatus(jobId, userId, JOB_RUNNING)

	go startJellyfinSync(
		db,
		jobId,
		userId,
		username,
		userThirdPartyId,
		userThirdPartyAuth,
	)

	return JellyfinSyncResponse{JobId: jobId}, nil
}
