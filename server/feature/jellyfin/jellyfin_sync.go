package jellyfin

import (
	"errors"
	"log/slog"
	"strconv"

	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/watched/episode"
	"github.com/sbondCo/Watcharr/feature/watched/season"
	"github.com/sbondCo/Watcharr/job"
	"github.com/sbondCo/Watcharr/util"
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

type WatchedProvider interface {
	AddWatched(userId uint, ar domain.WatchedAddRequest, extraProps domain.WatchedAddExtraProps) (entity.Watched, error)
}

type WatchedSeasonProvider interface {
	AddWatchedSeason(userId uint, ar season.WatchedSeasonAddRequest) (season.WatchedSeasonAddResponse, error)
}

type WatchedEpisodeProvider interface {
	AddWatchedEpisodes(userId uint, ar episode.WatchedEpisodeAddRequest) (episode.WatchedEpisodeAddResponse, error)
}

type SyncService struct {
	cfg              *config.ServerConfig
	service          *Service
	wp               WatchedProvider
	wsp              WatchedSeasonProvider
	wep              WatchedEpisodeProvider
	activityProvider domain.ActivityAddProvider
}

func NewSyncService(
	cfg *config.ServerConfig,
	service *Service,
	wp WatchedProvider,
	wsp WatchedSeasonProvider,
	wep WatchedEpisodeProvider,
	activityProvider domain.ActivityAddProvider,
) *SyncService {
	return &SyncService{
		cfg,
		service,
		wp,
		wsp,
		wep,
		activityProvider,
	}
}

// Perform the jellyfin sync.
// Gets each type of media separately from jellyfin and attempts to import them.
// Errors are added silently to the job.
func (s *SyncService) startJellyfinSync(
	jobId string,
	userId uint,
	username string,
	userThirdPartyId string,
	userThirdPartyAuth string,
) {
	// Get played movies
	job.UpdateJobCurrentTask(jobId, userId, "syncing movies")
	playedMovies := new(JellyfinItemSearchResponse)
	err := s.service.JellyfinAPIRequest(
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
		job.AddJobError(jobId, userId, "failed to get jellyfin response for movies")
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
					job.AddJobError(jobId, userId, "movie could not be imported (no tmdbId present): "+v.Name)
					continue
				}
				tmdbId, err := strconv.Atoi(v.ProviderIds.Tmdb)
				if err != nil {
					slog.Error("jellyfinSyncWatched: Movie to import does not have a parseable (to int) tmdb id.", "movie_name", v.Name, "movie_ids", v.ProviderIds, "user_id", userId)
					job.AddJobError(jobId, userId, "movie could not be imported (tmdbId was not parseable): "+v.Name)
					continue
				}

				job.UpdateJobCurrentTask(jobId, userId, "syncing "+v.Name)

				// 2. Imported watched movie
				w, err := s.wp.AddWatched(
					userId,
					domain.WatchedAddRequest{
						Status:      entity.FINISHED,
						ContentType: util.SupportedMediaMovie,
						TMDBID:      tmdbId,
						WatchedDate: v.UserData.LastPlayedDate,
					}, domain.WatchedAddExtraProps{
						ActivityType: entity.IMPORTED_WATCHED_JF,
						DontRestore:  true,
					})
				if err != nil {
					if errors.Is(err, domain.ErrWatchedExists) {
						slog.Info("jellyfinSyncWatched: Content already exists on list.",
							"movie_name", v.Name,
							"movie_ids", v.ProviderIds,
							"user_id", userId)
					} else if errors.Is(err, domain.ErrWatchedExistsSoftDeleted) {
						slog.Warn("jellyfinSyncWatched: Movie exists on list soft deleted.")
						job.AddJobError(
							jobId,
							userId,
							"failed to add movie "+
								v.Name+
								". You have previously deleted it from your list!")
						// We don't continue as it was manually removed as is still
						// soft deleted. We don't want to re-add it (user should un-delete themselves).
						continue
					} else {
						slog.Error("jellyfinSyncWatched: Movie failed to import.",
							"movie_name", v.Name,
							"movie_ids", v.ProviderIds,
							"user_id", userId,
							"error", err)
						job.AddJobError(
							jobId,
							userId,
							"movie could not be imported (failed when adding to watched list): "+v.Name)
					}
				} else {
					// 3. Add IMPORTED_ADDED_WATCHED_JF activity
					if !v.UserData.LastPlayedDate.IsZero() {
						_, err := s.activityProvider.AddActivity(userId, domain.ActivityAddRequest{WatchedID: w.ID, Type: entity.IMPORTED_ADDED_WATCHED_JF, CustomDate: &v.UserData.LastPlayedDate})
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
	job.UpdateJobCurrentTask(jobId, userId, "syncing series")
	allSeries := new(JellyfinItemSearchResponse)
	err = s.service.JellyfinAPIRequest(
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
		job.AddJobError(jobId, userId, "failed to get jellyfin response for series")
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
					job.AddJobError(jobId, userId, "series could not be imported (no tmdbId present): "+v.Name)
					continue
				}
				tmdbId, err := strconv.Atoi(v.ProviderIds.Tmdb)
				if err != nil {
					slog.Error("jellyfinSyncWatched: Series to import does not have a parseable (to int) tmdb id.", "series_name", v.Name, "series_ids", v.ProviderIds, "user_id", userId)
					job.AddJobError(jobId, userId, "series could not be imported (tmdbId was not parseable): "+v.Name)
					continue
				}

				job.UpdateJobCurrentTask(jobId, userId, "syncing serie "+v.Name)

				// 2. Imported watched series
				w, err := s.wp.AddWatched(
					userId,
					domain.WatchedAddRequest{
						Status:      entity.FINISHED,
						ContentType: util.SupportedMediaShow,
						TMDBID:      tmdbId,
						WatchedDate: v.UserData.LastPlayedDate,
					}, domain.WatchedAddExtraProps{
						ActivityType: entity.IMPORTED_WATCHED_JF,
						DontRestore:  true,
					})
				if err != nil {
					if errors.Is(err, domain.ErrWatchedExists) {
						slog.Info("jellyfinSyncWatched: Content already exists on list.",
							"series_name", v.Name,
							"series_ids", v.ProviderIds,
							"user_id", userId,
							"watched_id", w.ID)
						// In this case, we allow continuing below to start syncing seasons/episodes
					} else if errors.Is(err, domain.ErrWatchedExistsSoftDeleted) {
						slog.Warn("jellyfinSyncWatched: Show exists on list soft deleted.")
						job.AddJobError(
							jobId,
							userId,
							"failed to add show "+
								v.Name+
								". You have previously deleted it from your list!")
						// We don't continue as it was manually removed as is still
						// soft deleted. We don't want to re-add it (user should un-delete themselves).
						continue
					} else {
						slog.Error("jellyfinSyncWatched: Series failed to import.",
							"series_name", v.Name,
							"series_ids", v.ProviderIds,
							"user_id", userId,
							"error", err)
						job.AddJobError(jobId, userId, "series could not be imported (failed when adding to watched list): "+v.Name)
					}
				} else {
					// 3. Add IMPORTED_ADDED_WATCHED activity (only if no err above, show also must not have already been on our list)
					if !v.UserData.LastPlayedDate.IsZero() {
						_, err := s.activityProvider.AddActivity(userId, domain.ActivityAddRequest{WatchedID: w.ID, Type: entity.IMPORTED_ADDED_WATCHED_JF, CustomDate: &v.UserData.LastPlayedDate})
						if err != nil {
							slog.Error("jellyfinSyncWatched: Failed to add dateswatched activity.", "series_name", v.Name,
								"series_ids", v.ProviderIds, "user_id", userId, "date", v.UserData.LastPlayedDate, "error", err)
						}
					}
				}

				// 4. Import watched seasons for this serie
				// Get all show seasons (filtering isPlayed doesn't seem to be a thing, so we will have to do that ourselves)
				seriesSeasons := new(JellyfinSeriesSeasonsResponse)
				err = s.service.JellyfinAPIRequest(
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
					job.AddJobError(jobId, userId, "series seasons could not be imported (request failed): "+v.Name)
				} else if len(seriesSeasons.Items) <= 0 {
					slog.Info("jellyfinSyncWatched: Series has no seasons.", "series_name", v.Name, "series_ids", v.ProviderIds, "user_id", userId)
				} else {
					for _, vs := range seriesSeasons.Items {
						slog.Debug("jellyfinSyncWatched: Processing a season.", "full_item", v, "user_id", userId)
						if !vs.UserData.Played {
							slog.Debug("jellyfinSyncWatched: Skipping import of unplayed season.", "series_name", v.Name, "season_num", vs.IndexNumber, "user_id", userId)
							continue
						}
						job.UpdateJobCurrentTask(jobId, userId, "syncing "+v.Name+" season "+strconv.Itoa(vs.IndexNumber))
						_, err = s.wsp.AddWatchedSeason(userId, season.WatchedSeasonAddRequest{
							WatchedID:       w.ID,
							SeasonNumber:    vs.IndexNumber,
							Status:          entity.FINISHED,
							AddActivity:     entity.SEASON_ADDED_JF,
							AddActivityDate: vs.UserData.LastPlayedDate,
						})
						if err != nil {
							slog.Error("jellyfinSyncWatched: Failed to fetch series seasons.", "series_name", v.Name, "series_ids", v.ProviderIds, "user_id", userId, "error", err)
							job.AddJobError(jobId, userId, "series season could not be imported (addWatchedSeason request failed): "+v.Name+" season "+strconv.Itoa(vs.IndexNumber))
						}
					}
				}

				// 5. Import watched episodes for this serie
				// Gets all show episodes (filtering isPlayed doesn't seem to be a thing, so we will have to do that ourselves)
				seriesEpisodes := new(JellyfinSeriesEpisodesResponse)
				err = s.service.JellyfinAPIRequest(
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
					job.AddJobError(jobId, userId, "series episodes could not be imported (request failed): "+v.Name)
				} else if len(seriesEpisodes.Items) <= 0 {
					slog.Info("jellyfinSyncWatched: Series has no episodes.", "series_name", v.Name, "series_ids", v.ProviderIds, "user_id", userId)
				} else {
					for _, vs := range seriesEpisodes.Items {
						slog.Debug("jellyfinSyncWatched: Processing an episode.", "full_item", v, "user_id", userId)
						if !vs.UserData.Played {
							slog.Debug("jellyfinSyncWatched: Skipping import of unplayed episode.", "series_name", v.Name, "season_num", vs.ParentIndexNumber, "episode_num", vs.IndexNumber, "user_id", userId)
							continue
						}
						job.UpdateJobCurrentTask(jobId, userId, "syncing "+v.Name+" season "+strconv.Itoa(vs.ParentIndexNumber)+" episode "+strconv.Itoa(vs.IndexNumber))
						_, err = s.wep.AddWatchedEpisodes(userId, episode.WatchedEpisodeAddRequest{
							WatchedID:       w.ID,
							SeasonNumber:    vs.ParentIndexNumber,
							EpisodeNumber:   vs.IndexNumber,
							Status:          entity.FINISHED,
							AddActivity:     entity.EPISODE_ADDED_JF,
							AddActivityDate: vs.UserData.LastPlayedDate,
						})
						if err != nil {
							slog.Error("jellyfinSyncWatched: Failed to import series episode.", "series_name", v.Name, "season_num", vs.ParentIndexNumber, "episode_num", vs.IndexNumber, "user_id", userId, "error", err)
							job.AddJobError(jobId, userId, "series episode could not be imported (addWatchedEpisode request failed): "+v.Name+" "+vs.Name)
						}
					}
				}
			}
		}
	}

	job.UpdateJobStatus(jobId, userId, job.JOB_DONE)
}

func (s *SyncService) jellyfinSyncWatched(
	userId uint,
	userType entity.UserType,
	username string,
	userThirdPartyId string,
	userThirdPartyAuth string,
) (JellyfinSyncResponse, error) {
	jobId, err := job.AddJob("jf_sync", userId)
	if err != nil {
		slog.Error("jellyfinSyncWatched: Failed to create a job", "error", err)
		return JellyfinSyncResponse{}, errors.New("failed to create job")
	}

	job.UpdateJobStatus(jobId, userId, job.JOB_RUNNING)

	go s.startJellyfinSync(
		jobId,
		userId,
		username,
		userThirdPartyId,
		userThirdPartyAuth,
	)

	return JellyfinSyncResponse{JobId: jobId}, nil
}
