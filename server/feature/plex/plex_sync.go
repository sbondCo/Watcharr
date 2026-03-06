package plex

import (
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/watched/episode"
	"github.com/sbondCo/Watcharr/feature/watched/season"
	"github.com/sbondCo/Watcharr/job"
	"github.com/sbondCo/Watcharr/util"
)

type PlexSyncResponse struct {
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
	s                *Service
	wp               WatchedProvider
	wsp              WatchedSeasonProvider
	wep              WatchedEpisodeProvider
	activityProvider domain.ActivityAddProvider
}

func NewSyncService(
	s *Service,
	wp WatchedProvider,
	wsp WatchedSeasonProvider,
	wep WatchedEpisodeProvider,
	activityProvider domain.ActivityAddProvider,
) *SyncService {
	return &SyncService{
		s,
		wp,
		wsp,
		wep,
		activityProvider,
	}
}

// Perform a Plex sync.
// Errors are added silently to the job.
func (s *SyncService) startPlexSync(
	jobId string,
	userId uint,
	userPlexLocalAuth string,
) {
	job.UpdateJobCurrentTask(jobId, userId, "fetching libraries")
	libraries, err := s.s.GetPlexLibraries(userPlexLocalAuth)
	if err != nil {
		slog.Error("plexSyncWatched: Failed to fetch libraries", "user_id", userId, "error", err)
		job.AddJobError(jobId, userId, "failed to get plex libraries")
		job.UpdateJobStatus(jobId, userId, job.JOB_DONE)
		return
	}
	for _, library := range libraries.MediaContainer.Directory {
		slog.Debug("plexSyncWatched: Processing a library", "library_title", library.Title, "library_type", library.Type, "user_id", userId)
		if library.Type == "movie" {
			job.UpdateJobCurrentTask(jobId, userId, "importing movies from "+library.Title)
			movies, err := s.s.GetPlexLibraryItems(userPlexLocalAuth, library.Key)
			if err != nil {
				slog.Error("plexSyncWatched: Failed to fetch movies from library", "library", library.Key, "user_id", userId, "error", err)
				job.AddJobError(jobId, userId, "failed to fetch movies from library "+library.Key)
				continue
			}
			for _, movie := range movies.MediaContainer.Metadata {
				if movie.ViewCount == 0 {
					// Not viewed and not rated, skip importing
					slog.Debug("plexSyncWatched: Skipping unwatched movie:", "movie_name", movie.Title, "user_id", userId)
					continue
				}
				job.UpdateJobCurrentTask(jobId, userId, "importing movie "+movie.Title)
				slog.Info("plexSyncWatched: Importing movie.", "movie_name", movie.Title, "user_id", userId)

				// Find tmdb id
				if len(movie.Guid) <= 0 {
					slog.Error("plexSyncWatched: Movie to import does not have any external guids.", "movie_name", movie.Title, "movie_id", movie.GUID, "user_id", userId)
					job.AddJobError(jobId, userId, "movie could not be imported (no external ids present): "+movie.Title)
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
					job.AddJobError(jobId, userId, "movie could not be imported (no tmdbId present): "+movie.Title)
					continue
				}
				tmdbId, err := strconv.Atoi(tmdbIdStr)
				if err != nil {
					slog.Error("plexSyncWatched: Movie to import does not have a parseable (to int) tmdb id.", "movie_name", movie.Title, "tmdb_id_str", tmdbIdStr, "movie_id", movie.GUID, "user_id", userId)
					job.AddJobError(jobId, userId, "movie could not be imported (tmdbId was not parseable): "+movie.Title)
					continue
				}

				lastViewedAt := time.Unix(movie.LastViewedAt, 0)
				w, err := s.wp.AddWatched(
					userId,
					domain.WatchedAddRequest{
						Status:      entity.FINISHED,
						TMDBID:      tmdbId,
						ContentType: util.SupportedMediaMovie,
						Rating:      float64(movie.UserRating),
						WatchedDate: lastViewedAt,
					},
					domain.WatchedAddExtraProps{
						ActivityType: entity.IMPORTED_WATCHED_PLEX,
						DontRestore:  true,
					})
				if err != nil {
					if errors.Is(err, domain.ErrWatchedExists) {
						slog.Info("plexSyncWatched: Movie already exists on list.")
						continue
					}
					if errors.Is(err, domain.ErrWatchedExistsSoftDeleted) {
						// Entry was manually soft deleted by user at another point,
						// we DontRestore these automatically.
						slog.Warn("plexSyncWatched: Movie exists on list soft deleted.")
						job.AddJobError(
							jobId,
							userId,
							"failed to add movie "+
								movie.Title+
								". You have previously deleted it from your list!")
						continue
					}
					slog.Error("plexSyncWatched: Failed to add movie as watched", "error", err)
					job.AddJobError(jobId, userId, "failed to add movie "+movie.Title)
				} else {
					// 3. Add IMPORTED_ADDED_WATCHED_PLEX activity
					if !lastViewedAt.IsZero() {
						_, err := s.activityProvider.AddActivity(userId, domain.ActivityAddRequest{
							WatchedID:  w.ID,
							Type:       entity.IMPORTED_ADDED_WATCHED_PLEX,
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
			job.UpdateJobCurrentTask(jobId, userId, "importing tv shows from "+library.Title)
			shows, err := s.s.GetPlexLibraryItems(userPlexLocalAuth, library.Key)
			if err != nil {
				slog.Error("plexSyncWatched: Failed to fetch shows from library", "library", library.Key, "error", err)
				job.AddJobError(jobId, userId, "failed to fetch shows from library "+library.Key)
				continue
			}
			for _, show := range shows.MediaContainer.Metadata {
				if show.ViewedLeafCount != show.LeafCount {
					// Not viewed, skip importing
					// (could be improved to set status as watching when viewedLeafCount is higher than 0)
					slog.Debug("plexSyncWatched: Skipping unwatched show:", "show_name", show.Title, "leaf_count", show.LeafCount, "viewed_leaf_count", show.ViewedLeafCount, "user_id", userId)
					continue
				}
				job.UpdateJobCurrentTask(jobId, userId, "importing show "+show.Title)
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
					job.AddJobError(jobId, userId, "movie could not be imported (no tmdbId present): "+show.Title)
					continue
				}
				tmdbId, err := strconv.Atoi(tmdbIdStr)
				if err != nil {
					slog.Error("plexSyncWatched: Show to import does not have a parseable (to int) tmdb id.", "show_name", show.Title, "tmdb_id_str", tmdbIdStr, "show_id", show.GUID, "user_id", userId)
					job.AddJobError(jobId, userId, "show could not be imported (tmdbId was not parseable): "+show.Title)
					continue
				}

				lastViewedAt := time.Unix(show.LastViewedAt, 0)
				w, err := s.wp.AddWatched(
					userId,
					domain.WatchedAddRequest{
						Status:      entity.FINISHED,
						ContentType: util.SupportedMediaShow,
						TMDBID:      tmdbId,
						Rating:      float64(show.UserRating),
						WatchedDate: lastViewedAt,
					},
					domain.WatchedAddExtraProps{
						ActivityType: entity.IMPORTED_WATCHED_PLEX,
						DontRestore:  true,
					})
				if err != nil {
					if errors.Is(err, domain.ErrWatchedExists) {
						slog.Info("plexSyncWatched: Show already exists on list.")
						// In this case, we allow continuing below to start syncing seasons/episodes
					} else if errors.Is(err, domain.ErrWatchedExistsSoftDeleted) {
						// Entry was manually soft deleted by user at another point,
						// we DontRestore these automatically. We also don't continue
						// because eps/seasons cant be synced.
						slog.Warn("plexSyncWatched: Show exists on list soft deleted.")
						job.AddJobError(
							jobId,
							userId,
							"failed to add show "+
								show.Title+
								". You have previously deleted it from your list!")
						continue
					} else {
						slog.Error("plexSyncWatched: Failed to add show as watched", "error", err)
						job.AddJobError(jobId, userId, "failed to add show "+show.Title)
					}
				} else {
					// 3. Add IMPORTED_ADDED_WATCHED_PLEX activity
					if !lastViewedAt.IsZero() {
						_, err := s.activityProvider.AddActivity(userId, domain.ActivityAddRequest{
							WatchedID:  w.ID,
							Type:       entity.IMPORTED_ADDED_WATCHED_PLEX,
							CustomDate: &lastViewedAt,
						})
						if err != nil {
							slog.Error("plexSyncWatched: Failed to add dateswatched activity.", "movie_name", show.Title,
								"movie_id", show.GUID, "user_id", userId, "date", lastViewedAt, "unparsed_date", show.LastViewedAt, "error", err)
						}
					}
				}

				// Import watched seasons for this serie
				seriesSeasons, err := s.s.GetPlexLibraryItemSeasons(userPlexLocalAuth, show.RatingKey)
				if err != nil {
					slog.Error("plexSyncWatched: Failed to fetch series seasons.", "series_name", show.Title, "series_id", show.GUID, "user_id", userId, "error", err)
					job.AddJobError(jobId, userId, "series seasons could not be imported (request failed): "+show.Title)
				} else if len(seriesSeasons.MediaContainer.Metadata) <= 0 {
					slog.Info("plexSyncWatched: Series has no seasons.", "series_name", show.Title, "serie_ids", show.GUID, "user_id", userId)
				} else {
					for _, vs := range seriesSeasons.MediaContainer.Metadata {
						slog.Debug("plexSyncWatched: Processing a season.", "full_item", vs, "user_id", userId)
						if vs.ViewedLeafCount != vs.LeafCount {
							slog.Debug("plexSyncWatched: Skipping import of unplayed season.", "series_name", show.Title, "season_num", vs.Index, "user_id", userId)
							continue
						}
						job.UpdateJobCurrentTask(jobId, userId, "syncing "+show.Title+" season "+strconv.Itoa(vs.Index))
						var seasonLastViewedAt time.Time
						if vs.LastViewedAt != 0 {
							seasonLastViewedAt = time.Unix(vs.LastViewedAt, 0)
						}
						_, err = s.wsp.AddWatchedSeason(userId, season.WatchedSeasonAddRequest{
							WatchedID:       w.ID,
							SeasonNumber:    vs.Index,
							Status:          entity.FINISHED,
							AddActivity:     entity.SEASON_ADDED_PLEX,
							AddActivityDate: seasonLastViewedAt,
						})
						if err != nil {
							slog.Error("plexSyncWatched: Failed to fetch series seasons.", "series_name", show.Title, "series_id", show.GUID, "user_id", userId, "error", err)
							job.AddJobError(jobId, userId, "series season could not be imported (addWatchedSeason request failed): "+show.Title+" season "+strconv.Itoa(vs.Index))
						}
					}
				}

				// Import watched episodes for this serie
				seriesEpisodes, err := s.s.GetPlexLibraryItemEpisodes(userPlexLocalAuth, show.RatingKey)
				if err != nil {
					slog.Error("plexSyncWatched: Failed to fetch series episodes.", "series_name", show.Title, "series_id", show.GUID, "user_id", userId, "error", err)
					job.AddJobError(jobId, userId, "series episodes could not be imported (request failed): "+show.Title)
				} else if len(seriesEpisodes.MediaContainer.Metadata) <= 0 {
					slog.Info("plexSyncWatched: Series has no episodes.", "series_name", show.Title, "series_id", show.GUID, "user_id", userId)
				} else {
					for _, vs := range seriesEpisodes.MediaContainer.Metadata {
						slog.Debug("plexSyncWatched: Processing an episode.", "full_item", vs, "user_id", userId)
						if vs.ViewCount <= 0 {
							slog.Debug("plexSyncWatched: Skipping import of unplayed episode.", "series_name", show.Title, "season_num", vs.ParentIndex, "episode_num", vs.Index, "user_id", userId)
							continue
						}
						job.UpdateJobCurrentTask(jobId, userId, "syncing "+show.Title+" season "+strconv.Itoa(vs.ParentIndex)+" episode "+strconv.Itoa(vs.Index))
						var episodeLastViewedAt time.Time
						if vs.LastViewedAt != 0 {
							episodeLastViewedAt = time.Unix(vs.LastViewedAt, 0)
						}
						_, err = s.wep.AddWatchedEpisodes(userId, episode.WatchedEpisodeAddRequest{
							WatchedID:       w.ID,
							SeasonNumber:    vs.ParentIndex,
							EpisodeNumber:   vs.Index,
							Status:          entity.FINISHED,
							AddActivity:     entity.EPISODE_ADDED_PLEX,
							AddActivityDate: episodeLastViewedAt,
						})
						if err != nil {
							slog.Error("plexSyncWatched: Failed to import series episode.", "series_name", show.Title, "season_num", vs.ParentIndex, "episode_num", vs.Index, "user_id", userId, "error", err)
							job.AddJobError(jobId, userId, "series episode could not be imported (addWatchedEpisode request failed): "+show.Title+" "+vs.Title)
						}
					}
				}
			}
		}
	}
	job.UpdateJobStatus(jobId, userId, job.JOB_DONE)
}

func (s *SyncService) PlexSyncWatched(
	userId uint,
	userPlexLocalAuth string,
) (PlexSyncResponse, error) {
	jobId, err := job.AddJob("plex_sync", userId)
	if err != nil {
		slog.Error("startPlexSync: Failed to create a job", "error", err)
		return PlexSyncResponse{}, errors.New("failed to create job")
	}

	job.UpdateJobStatus(jobId, userId, job.JOB_RUNNING)

	go s.startPlexSync(
		jobId,
		userId,
		userPlexLocalAuth,
	)

	return PlexSyncResponse{JobId: jobId}, nil
}
