package imprt

import (
	"log/slog"
	"strconv"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/watched/episode"
	"github.com/sbondCo/Watcharr/media/tmdb"
	"github.com/sbondCo/Watcharr/util"
)

// A single episode of a show, identified the way we store watched episodes.
type episodeRef struct {
	SeasonNumber  int
	EpisodeNumber int
}

// How many episodes a season has, which is all we need from a show to work
// out which episodes an absolute watched count covers.
type seasonEpisodeCount struct {
	Number       int
	EpisodeCount int
}

// Work out which episodes a watched count covers, within one season.
//
// Some sources only tell us how many episodes have been watched, not which
// ones. A MyAnimeList export is the case we have: my_watched_episodes counts
// from the start of the entry, and a MyAnimeList entry is a single season,
// since a sequel gets an entry of its own there.
//
// So the count is only ever counted against one season and must not run past
// the end of it. Carrying a remainder into the next season would be wrong
// twice over: that season is a different entry with its own count, and the
// episodes it marked would be the wrong ones.
//
// Season 0 is skipped. Specials are not part of the running order the count
// is counted against, so including them would shift every episode along.
//
// The count is clamped to the season's episode count, because MyAnimeList
// and tmdb do not always agree on how many episodes a season has. Clamping
// undercounts by the difference, which is better than marking episodes that
// the count never covered.
func episodesForCount(
	seasons []seasonEpisodeCount,
	count int,
) []episodeRef {
	if count <= 0 {
		return nil
	}
	// The season the count belongs to, which is the first real one. Nothing
	// guarantees tmdb gives us the seasons in order, so pick the lowest
	// numbered rather than trusting the response.
	var season seasonEpisodeCount
	for _, v := range seasons {
		if v.Number <= 0 || v.EpisodeCount <= 0 {
			continue
		}
		if season.Number == 0 || v.Number < season.Number {
			season = v
		}
	}
	if season.Number == 0 {
		slog.Warn("episodesForCount: Show has no season we can fill in.",
			"watched_count", count)
		return nil
	}
	take := count
	if take > season.EpisodeCount {
		slog.Warn("episodesForCount: Season has fewer episodes than the import"+
			" says were watched, filling in what we can.",
			"season", season.Number, "watched_count", count,
			"season_episodes", season.EpisodeCount,
			"not_filled_in", count-season.EpisodeCount)
		take = season.EpisodeCount
	}
	eps := make([]episodeRef, 0, take)
	for e := 1; e <= take; e++ {
		eps = append(eps, episodeRef{SeasonNumber: season.Number, EpisodeNumber: e})
	}
	return eps
}

// Get the seasons of a show from tmdb, to fill a watched count into.
func (s *Service) getShowSeasons(tmdbId int) []seasonEpisodeCount {
	details, err := s.tmdb.ShowDetails(tmdb.ShowDetailsOptions{
		ID: strconv.Itoa(tmdbId),
		// The show is cached by the import match that got us here, so there
		// is nothing for us to cache again.
		DontRunDBCache: true,
	})
	if err != nil {
		slog.Error("getShowSeasons: Failed to get show details from tmdb.",
			"tmdb_id", tmdbId, "error", err)
		return nil
	}
	seasons := make([]seasonEpisodeCount, 0, len(details.Seasons))
	for _, v := range details.Seasons {
		seasons = append(seasons, seasonEpisodeCount{
			Number:       v.SeasonNumber,
			EpisodeCount: v.EpisodeCount,
		})
	}
	return seasons
}

// Mark the episodes an absolute watched count covers as watched on w.
//
// Returns how many episodes were added.
func (s *Service) addEpisodesFromCount(
	userId uint,
	w *entity.Watched,
	tmdbId int,
	count int,
) int {
	eps := episodesForCount(s.getShowSeasons(tmdbId), count)
	if len(eps) <= 0 {
		slog.Warn("addEpisodesFromCount: Worked out no episodes to add.",
			"tmdb_id", tmdbId, "watched_count", count)
		return 0
	}
	return s.addEpisodes(userId, w, eps)
}

// Mark each of eps as watched on w.
//
// Episodes that already have a status are left alone, so this never
// overwrites progress the user set themselves and re-running an import adds
// nothing the second time. Returns how many episodes were added.
func (s *Service) addEpisodes(
	userId uint,
	w *entity.Watched,
	eps []episodeRef,
) int {
	added := 0
	for _, e := range eps {
		if hasWatchedEpisode(w.WatchedEpisodes, e) {
			continue
		}
		resp, err := s.wep.AddWatchedEpisodes(userId, episode.WatchedEpisodeAddRequest{
			WatchedID:     w.ID,
			SeasonNumber:  e.SeasonNumber,
			EpisodeNumber: e.EpisodeNumber,
			Status:        entity.FINISHED,
		})
		if err != nil {
			slog.Error("addEpisodes: Failed to add a watched episode.",
				"watched_id", w.ID, "season", e.SeasonNumber,
				"episode", e.EpisodeNumber, "error", err)
			continue
		}
		// The response carries the full episode list, so keeping it means the
		// already watched check above stays correct as we go.
		w.WatchedEpisodes = resp.WatchedEpisodes
		added++
	}
	slog.Info("addEpisodes: Added watched episodes.",
		"watched_id", w.ID, "asked_for", len(eps), "added", added)
	return added
}

// Whether an episode already has a watched status.
func hasWatchedEpisode(eps []entity.WatchedEpisode, e episodeRef) bool {
	for _, v := range eps {
		if v.SeasonNumber == e.SeasonNumber && v.EpisodeNumber == e.EpisodeNumber {
			return true
		}
	}
	return false
}

// Fill in episode progress on content that is already on the users list.
//
// An import can still be worth something when the content is already there.
// A show added by a Plex or Jellyfin sync has no episodes marked when the
// server holds none of them, so a MyAnimeList export that knows how far the
// user got can fill them in.
//
// Existing episodes are never touched, so this cannot clobber progress the
// user set themselves. When there is nothing to fill in, the import is still
// reported as IMPORT_EXISTS, exactly as before.
func (s *Service) fillInMissingEpisodes(
	userId uint,
	ar *domain.ImportRequest,
	props domain.SuccessfulImportProps,
) domain.ImportResponse {
	exists := domain.ImportResponse{Type: domain.IMPORT_EXISTS}
	if ar.WatchedEpisodesCount <= 0 {
		return exists
	}
	// Only shows have episodes, and only shows are looked up by tmdb id.
	if props.ContentType != util.SupportedMediaShow {
		return exists
	}
	existing, err := s.wp.GetWatchedItemByTmdbId(
		userId,
		uint(props.TmdbID),
		entity.ContentType(props.ContentType),
	)
	if err != nil {
		slog.Error("fillInMissingEpisodes: Failed to get the existing watched item",
			"tmdb_id", props.TmdbID, "error", err)
		return exists
	}
	if existing.ID == 0 {
		slog.Warn("fillInMissingEpisodes: No existing watched item found",
			"tmdb_id", props.TmdbID)
		return exists
	}
	if s.addEpisodesFromCount(userId, &existing, props.TmdbID, ar.WatchedEpisodesCount) <= 0 {
		slog.Debug("fillInMissingEpisodes: Existing item had nothing to fill in,"+
			" leaving it be", "watched_id", existing.ID)
		return exists
	}
	return domain.ImportResponse{
		Type:         domain.IMPORT_EPISODES_UPDATED,
		WatchedEntry: existing,
	}
}

// Fill in details on content that is already on the users list.
//
// Both a missing rating and missing episode progress are filled in when the
// import has them, since an import can carry either. The response can only
// name one of them, so episodes win when both happened, being the larger
// change of the two.
func (s *Service) fillInExistingContent(
	userId uint,
	ar *domain.ImportRequest,
	props domain.SuccessfulImportProps,
) domain.ImportResponse {
	// The rating goes first so the entry the episode fill reads back, and
	// returns to the user, already has it.
	ratingResp := s.fillInMissingRating(userId, ar, props)
	episodesResp := s.fillInMissingEpisodes(userId, ar, props)
	if episodesResp.Type == domain.IMPORT_EPISODES_UPDATED {
		return episodesResp
	}
	return ratingResp
}
