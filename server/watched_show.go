package main

import (
    "errors"
    "log/slog"
    "strconv"
    "time"

    "gorm.io/gorm"
)

// WatchedShowAddRequest is used by the /watched/show endpoint to add (and optionally
// create) a show together with one of its episodes.
//
// Either TMDBID (preferred) or TMDBEpisodeID MUST be provided. SeasonNumber and
// EpisodeNumber are always required.
// All rating fields are expected to be 0–10 (inclusive). Status values must be one
// of the WatchedStatus enum constants.
//
// swagger:model
type WatchedShowAddRequest struct {
    // OPTIONAL – if caller only has the episode-level TMDB id.
    TMDBEpisodeID int `json:"tmdbEpisodeId,omitempty"`
    // OPTIONAL – show-level TMDB id (preferred).
    TMDBID        int `json:"tmdbId,omitempty"`
    // OPTIONAL – TVDB episode id if webhook only supplies that.
    TvdbID        int `json:"tvdbId,omitempty"`

    SeasonNumber  int           `json:"seasonNumber"  binding:"required"`
    EpisodeNumber int           `json:"episodeNumber" binding:"required"`

    // Episode attributes
    Status        WatchedStatus `json:"status"`
    Rating        int8          `json:"rating" binding:"max=10"`

    // Show attributes (applied only when show is newly created)
    ShowStatus    WatchedStatus `json:"showStatus"`
    ShowRating    float64       `json:"showRating" binding:"max=10"`

    // Optional custom creation date when the show entry is created.
    WatchedDate   time.Time     `json:"watchedDate"`
}

// addWatchedShowAndEpisode ensures a Watched row exists for the requested TV show and
// then adds / updates the specific episode row.
func addWatchedShowAndEpisode(db *gorm.DB, userId uint, req WatchedShowAddRequest) (Watched, WatchedEpisodeAddResponse, error) {
    slog.Debug("addWatchedShowAndEpisode: started", "userId", userId, "req", req)

    // 1. Figure out the show TMDB id.
    showTmdbID := req.TMDBID
    // First, if we only have a TVDB id, translate it via TMDB find.
    if showTmdbID == 0 && req.TMDBEpisodeID == 0 && req.TvdbID != 0 {
        epId, shId, err := lookupTMDBIdsFromTVDB(req.TvdbID)
        if err != nil {
            return Watched{}, WatchedEpisodeAddResponse{}, err
        }
        // Prefer a direct show match when available – this avoids the extra episode→show
        // translation step and prevents 404s when the episode isn't yet indexed by TMDB.
        if shId != 0 {
            showTmdbID = shId
        } else if epId != 0 {
            req.TMDBEpisodeID = epId
        }
    }

    if showTmdbID == 0 {
        if req.TMDBEpisodeID == 0 {
            slog.Error("addWatchedShowAndEpisode: no usable id supplied")
            return Watched{}, WatchedEpisodeAddResponse{}, errors.New("provide tmdbId, tmdbEpisodeId or tvdbId")
        }
        var err error
        showTmdbID, err = lookupShowIdFromEpisode(req.TMDBEpisodeID)
        if err != nil {
            return Watched{}, WatchedEpisodeAddResponse{}, err
        }
    }

    // 2. Ensure Content cache exists (will create or update if needed).
    if _, err := getOrCacheContent(db, SHOW, showTmdbID); err != nil {
        return Watched{}, WatchedEpisodeAddResponse{}, err
    }

    // 3. Try obtain existing Watched row.
    w, err := getWatchedItemByTmdbId(db, userId, uint(showTmdbID), SHOW)
    if err != nil || w.ID == 0 {
        // If not found (or any error), create a new Watched row.
        wr := WatchedAddRequest{
            Status:      defaultShowStatus(req.ShowStatus),
            Rating:      req.ShowRating,
            ContentID:   showTmdbID,
            ContentType: SHOW,
            WatchedDate: req.WatchedDate,
        }
        w, err = addWatched(db, userId, wr, ADDED_WATCHED)
        if err != nil {
            return Watched{}, WatchedEpisodeAddResponse{}, err
        }
    }

    // 4. Always add / update the episode entry.
    er := WatchedEpisodeAddRequest{
        WatchedID:     w.ID,
        SeasonNumber:  req.SeasonNumber,
        EpisodeNumber: req.EpisodeNumber,
        Status:        req.Status,
        Rating:        req.Rating,
    }
    epResp, err := addWatchedEpisodes(db, userId, er)
    if err != nil {
        return Watched{}, WatchedEpisodeAddResponse{}, err
    }

    slog.Debug("addWatchedShowAndEpisode: finished", "watchedId", w.ID)
    return w, epResp, nil
}

// defaultShowStatus returns WATCHING when blank, otherwise the provided value.
func defaultShowStatus(s WatchedStatus) WatchedStatus {
    if s == "" {
        return WATCHING
    }
    return s
}

// lookupShowIdFromEpisode queries TMDB for the parent show id of the given episode id.
// lookupTMDBIdsFromTVDB translates a TVDB episode id to TMDB ids via the TMDB
// external-id lookup endpoint. It returns (episodeTMDB, showTMDB).
func lookupTMDBIdsFromTVDB(tvdbId int) (int, int, error) {
    slog.Debug("lookupTMDBIdsFromTVDB: running", "tvdbId", tvdbId)
    type findResp struct {
        TvEpisodeResults []struct {
            ID     int `json:"id"`
            ShowID int `json:"show_id"`
        } `json:"tv_episode_results"`
        TvResults []struct{ ID int `json:"id"` } `json:"tv_results"`
    }
    var fr findResp
    ep := "find/" + strconv.Itoa(tvdbId)
    // Strategy:
    //   1) Ask TMDB to treat the TVDB id as a *series* id (external_source=tvdb_id).
    //      If we get a tv_results entry we are done.
    //   2) If no series result, ask TMDB to treat the id as an *episode* id
    //      (external_source=tvdb_episode_id) and capture the TMDB episode id.
    //   3) Caller can decide whether the returned episode id needs a further
    //      /episode/{id} lookup to reach the parent show.

        var epTmdbID, showTmdbID int

    // --- 1. Try treating id as a *series* id first (tvdb_id) -------------------
    if err := tmdbRequest(ep, map[string]string{"external_source": "tvdb_id"}, &fr); err != nil {
        return 0, 0, err
    }
    if len(fr.TvResults) > 0 {
        // Best-case: we already have the parent show – nothing else needed.
        showTmdbID = fr.TvResults[0].ID
    }
    // Capture any episode match from this response as well (tvdb_id may still
    // populate tv_episode_results).
    var firstEpisodeMatch int
    var firstEpisodeMatchShowID int
    if len(fr.TvEpisodeResults) > 0 {
        firstEpisodeMatch = fr.TvEpisodeResults[0].ID
        if fr.TvEpisodeResults[0].ShowID != 0 {
            firstEpisodeMatchShowID = fr.TvEpisodeResults[0].ShowID
            // If we still lacked a show id, use this.
            if showTmdbID == 0 {
                showTmdbID = firstEpisodeMatchShowID
            }
        }
    }

    // --- 2. If we still don't know the show, try treating id as an episode id --
    if showTmdbID == 0 {
        fr = findResp{}
        if err := tmdbRequest(ep, map[string]string{"external_source": "tvdb_episode_id"}, &fr); err != nil {
            return 0, 0, err
        }
        if len(fr.TvEpisodeResults) > 0 {
            epRes := fr.TvEpisodeResults[0]
            epTmdbID = epRes.ID
            if epRes.ShowID != 0 {
                showTmdbID = epRes.ShowID
            }
        } else if firstEpisodeMatch != 0 {
            epTmdbID = firstEpisodeMatch
            if showTmdbID == 0 && firstEpisodeMatchShowID != 0 {
                showTmdbID = firstEpisodeMatchShowID
            }
        }
    }

    // Final sanity check
    if showTmdbID == 0 && epTmdbID == 0 {
        return 0, 0, errors.New("no TMDB match for tvdb id")
    }
    return epTmdbID, showTmdbID, nil
}

func lookupShowIdFromEpisode(episodeId int) (int, error) {
    slog.Debug("lookupShowIdFromEpisode: running", "episodeId", episodeId)
    type tmdbEpisodeDetails struct {
        ShowID int `json:"show_id"`
    }
    var details tmdbEpisodeDetails
    ep := "episode/" + strconv.Itoa(episodeId)
    if err := tmdbRequest(ep, nil, &details); err != nil {
        slog.Error("lookupShowIdFromEpisode: tmdb request failed", "error", err)
        return 0, err
    }
    if details.ShowID == 0 {
        return 0, errors.New("failed to resolve parent show id from episode id")
    }
    slog.Debug("lookupShowIdFromEpisode: success", "showId", details.ShowID)
    return details.ShowID, nil
}
