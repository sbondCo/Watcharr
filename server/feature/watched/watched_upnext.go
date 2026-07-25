package watched

import (
	"log/slog"
	"sort"
	"strconv"

	"github.com/sbondCo/Watcharr/database/entity"
)

// UpNextItem describes the next unwatched episode of an in-progress show,
// for the "Up Next" row on the overview page.
type UpNextItem struct {
	WatchedID     uint   `json:"watchedId"`
	TmdbID        int    `json:"tmdbId"`
	ShowTitle     string `json:"showTitle"`
	PosterPath    string `json:"posterPath"`
	SeasonNumber  int    `json:"seasonNumber"`
	EpisodeNumber int    `json:"episodeNumber"`
	// Total number of episodes in that season (for a "S6.E2/12" style display).
	SeasonEpisodeCount int    `json:"seasonEpisodeCount"`
	EpisodeName        string `json:"episodeName"`
	StillPath          string `json:"stillPath"`
	AirDate            string `json:"airDate"`
}

// UpNext returns, for each show the user is currently WATCHING, the next
// unwatched episode. Shows that are fully watched are omitted.
//
// Note: we deliberately do NOT filter on TMDB air_date. TMDB's air_date is the
// TV broadcast schedule, which does not reflect streaming/replay availability
// (a fully-streamable show can still have "future" broadcast dates). So we
// simply offer the next episode that exists and hasn't been watched.
func (s *Service) UpNext(userId uint) ([]UpNextItem, error) {
	var watched []entity.Watched
	res := s.db.
		Joins("Content").
		Preload("WatchedEpisodes").
		Where("watcheds.user_id = ? AND Content.type = ? AND watcheds.status = ?",
			userId, entity.SHOW, entity.WATCHING).
		Find(&watched)
	if res.Error != nil {
		slog.Error("UpNext: query failed", "error", res.Error)
		return nil, res.Error
	}

	items := []UpNextItem{}
	for i := range watched {
		if item, ok := s.nextEpisodeFor(&watched[i]); ok {
			items = append(items, item)
		}
	}
	return items, nil
}

// nextEpisodeFor computes the next unwatched episode for one show. It scans
// seasons starting from the highest season the user has already touched
// (episodes are usually watched contiguously), fetching season details from
// TMDB (cached) only as needed.
func (s *Service) nextEpisodeFor(w *entity.Watched) (UpNextItem, bool) {
	if w.Content == nil {
		return UpNextItem{}, false
	}
	// Set of already-watched (season, episode).
	seen := make(map[[2]int]bool, len(w.WatchedEpisodes))
	maxSeason := 1
	for _, we := range w.WatchedEpisodes {
		seen[[2]int{we.SeasonNumber, we.EpisodeNumber}] = true
		if we.SeasonNumber > maxSeason {
			maxSeason = we.SeasonNumber
		}
	}

	totalSeasons := int(w.Content.NumberOfSeasons)
	if totalSeasons < maxSeason {
		totalSeasons = maxSeason
	}
	tvId := strconv.Itoa(w.Content.TmdbID)

	for season := maxSeason; season <= totalSeasons; season++ {
		if season <= 0 { // skip specials (season 0)
			continue
		}
		sd, err := s.cp.SeasonDetails(tvId, strconv.Itoa(season))
		if err != nil {
			slog.Warn("UpNext: season details failed", "tmdbId", w.Content.TmdbID, "season", season, "error", err)
			continue
		}
		eps := sd.Episodes
		sort.Slice(eps, func(a, b int) bool { return eps[a].EpisodeNumber < eps[b].EpisodeNumber })
		for _, ep := range eps {
			if seen[[2]int{ep.SeasonNumber, ep.EpisodeNumber}] {
				continue
			}
			return UpNextItem{
				WatchedID:          w.ID,
				TmdbID:             w.Content.TmdbID,
				ShowTitle:          w.Content.Title,
				PosterPath:         w.Content.PosterPath,
				SeasonNumber:       ep.SeasonNumber,
				EpisodeNumber:      ep.EpisodeNumber,
				SeasonEpisodeCount: len(eps),
				EpisodeName:        ep.Name,
				StillPath:          ep.StillPath,
				AirDate:            ep.AirDate,
			}, true
		}
	}
	return UpNextItem{}, false
}
