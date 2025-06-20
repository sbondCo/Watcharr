package main

import (
    "errors"
    "sort"
    "strconv"
)

// probably the most difficult part to solve for automatic mapping of episodes from tvdb format to tmdb format, this implementation is based on filebot sonnarr/radarr basically, and it uses 
// the absolute number of the episode to map it to the correct season/episode in tmdb when tvdb disagrees about season labeling for whatever reason for live tv shows, anime, reality tv, etc.
// MapTVDBEpisodeToTMDB converts a TVDB episode id into the corresponding
// TMDB show-id, season and episode numbers using the simplified 3-request
// algorithm of get tvdb series id from episode id, find tmdb show id from tvdb
// series id, and then get tmdb show details to know season layout and mapping between episodes and absolute numbers.
// this prevents errors when translating episodes from emby's tvdb format to watcharr's tmdb format.
func MapTVDBEpisodeToTMDB(tvdbEpID int) (tmdbShowID int, tmdbSeason int, tmdbEpisode int, err error) {
    // 1. TVDB episode → (seriesID, absoluteNumber)
    seriesID, absNum, err := getTVDBEpisodeInfo(tvdbEpID)
    if err != nil {
        return 0, 0, 0, err
    }

    // 2. TVDB series → TMDB show id via /find
    tmdbShowID, err = findTMDBShowIDByTVDBSeriesID(seriesID)
    if err != nil {
        return 0, 0, 0, err
    }

    // 3. Grab TMDB show details to know season layout
    var details TMDBShowDetails
    if err = tmdbRequest("/tv/"+strconv.Itoa(tmdbShowID), nil, &details); err != nil {
        return 0, 0, 0, err
    }

    if uint32(absNum) > details.NumberOfEpisodes {
        return 0, 0, 0, errors.New("tvdb absoluteNumber exceeds tmdb total episodes")
    }

    // Ensure seasons sorted ascending.
    seasons := details.Seasons
    sort.Slice(seasons, func(i, j int) bool { return seasons[i].SeasonNumber < seasons[j].SeasonNumber })

    remaining := absNum
    for _, s := range seasons {
        if remaining <= s.EpisodeCount {
            return tmdbShowID, s.SeasonNumber, remaining, nil
        }
        remaining -= s.EpisodeCount
    }

    return 0, 0, 0, errors.New("failed to map absoluteNumber to tmdb season/episode")
}
