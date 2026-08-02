package search

import (
	"log/slog"
	"net/url"
	"strings"

	"github.com/sbondCo/Watcharr/domain"
)

// Perform "special"  direct search if possible using search query.
// Eg: Search term is in provider:id format or is a supported url.
func (s *Service) searchExtProviderById(
	query string,
	resp *domain.SearchResponse,
) bool {
	queryLower := strings.ToLower(query)

	provider, providerID := s.getExtProviderFromQuery(queryLower)

	if provider == "" || providerID == "" {
		return false
	}

	slog.Debug("searchExtProviderById: Processing.",
		"provider", provider,
		"provider_id", providerID)

	switch provider {
	case "movie":
		if err := s.searchMovieById(providerID, resp); err == nil {
			return true
		}
	case "tv":
		if err := s.searchTvById(providerID, resp); err == nil {
			return true
		}
	case "igdb":
		if err := s.searchGameById(providerID, resp); err == nil {
			return true
		}
	case "igdb-slug":
		if err := s.searchGameBySlug(providerID, resp); err == nil {
			return true
		}
	default:
		// By default, if provider name isn't caught in above cases, just send
		// it to tmdb external id search.
		tmdbRes, err := s.tmdb.SearchByExternalId(
			providerID,
			provider,
		)
		if err != nil {
			slog.Error("searchExtProviderById: Failed to search tmdb!",
				"error", err)
			return false
		}
		resLen := len(tmdbRes.Results)
		if resLen <= 0 {
			return false
		}
		for _, v := range tmdbRes.Results {
			resp.Results = append(
				resp.Results,
				v.AsMedia(),
			)
		}
		resp.Page = 1
		resp.TotalPages = 1
		resp.TotalResults = int64(resLen)
		return true
	}

	return false
}

// Takes in query and returns (Provider, ProviderID) if found.
func (s *Service) getExtProviderFromQuery(queryLower string) (string, string) {
	var provider string

	// Before checking for provider:providerid format, check if query is
	// a supported url.
	if p, i := s.getExtProviderFromURL(queryLower); p != "" && i != "" {
		slog.Debug("getExtProviderFromQuery: Returning from parsed url.")
		return p, i
	}

	querySplit := strings.Split(queryLower, ":")

	if len(querySplit) != 2 {
		slog.Debug("getExtProviderFromQuery: querySplit len != 2")
		return "", ""
	}

	switch querySplit[0] {
	case "movie", // TMDB ID target
		"tv",   // TMDB ID target
		"igdb", // IGDB ID target
		// The rest below are sent as is to tmdbs find by (external) id api.
		"imdb",
		"tvdb",
		"youtube",
		"wikidata",
		"facebook",
		"instagram",
		"twitter",
		"tiktok":
		provider = querySplit[0]
		// Any aliases we want to support
	case "i":
	case "imd":
		provider = "imdb"
	case "wd":
	case "wdt":
		provider = "wikidata"
	case "yt":
		provider = "youtube"
	case "thetvdb":
		provider = "tvdb"
	case "game":
		provider = "igdb"
	case "series":
		provider = "tv"
	default:
		slog.Debug("getExtProviderFromQuery: No provider found.")
		return "", ""
	}

	return provider, querySplit[1]
}

// Takes in what may be a url. If it is and is a supported url
// Returns (Provider, ProviderID).
func (s *Service) getExtProviderFromURL(maybeaurl string) (string, string) {
	u, err := url.Parse(maybeaurl)
	if err != nil || u.Host == "" {
		slog.Debug("getExtProviderFromURL: Doesn't look like a url.")
		return "", ""
	}

	hostLower := strings.ToLower(u.Host)
	slog.Debug("getExtProviderFromURL: Looks like a url.",
		"host", hostLower,
		"path", u.Path)

	// Using HasSuffix so for ex: www.imdb.com AND imdb.com will match.
	if strings.HasSuffix(hostLower, "imdb.com") {
		return s.getExtProviderIDFromIMDBURL(u)
	} else if strings.HasSuffix(hostLower, "themoviedb.org") {
		return s.getExtProviderIDFromTMDBURL(u)
	} else if strings.HasSuffix(hostLower, "igdb.com") {
		return s.getExtProviderIDFromIGDBURL(u)
	}

	return "", " "
}

// Extract id from IMDB url.
// Returns (Provider, ProviderID).
func (s *Service) getExtProviderIDFromIMDBURL(u *url.URL) (string, string) {
	segments := strings.Split(
		// Trim start/end '/' to avoid empty items at start/end
		// of final slice.
		strings.Trim(u.Path, "/"),
		"/",
	)
	segmentsLen := len(segments)
	slog.Debug("getExtProviderIDFromIMDBURL: Parsing path.",
		"segments", segments,
		"segments_len", segmentsLen)

	if segmentsLen < 2 ||
		segments[0] != "title" ||
		!strings.HasPrefix(segments[1], "tt") {
		slog.Debug("getExtProviderIDFromIMDBURL: path provided not supported.")
		return "", ""
	}

	return "imdb", segments[1]
}

// Extract id from TMDB url.
// Returns (Provider, ProviderID).
func (s *Service) getExtProviderIDFromTMDBURL(u *url.URL) (string, string) {
	// Split path by '/'
	segments := strings.Split(
		// Trim start/end '/' to avoid empty items at start/end
		// of final slice.
		strings.Trim(u.Path, "/"),
		"/",
	)
	segmentsLen := len(segments)
	slog.Debug("getExtProviderIDFromTMDBURL: Parsing path.",
		"segments", segments,
		"segments_len", segmentsLen)

	// Check if segments of the path are valid as a tv/movie page.
	if segmentsLen < 2 ||
		(segments[0] != "tv" && segments[0] != "movie") ||
		segments[1] == "" {
		slog.Debug("getExtProviderIDFromTMDBURL: path provided not supported.")
		return "", ""
	}

	// Extract id from second segment.
	segs2 := strings.SplitN(segments[1], "-", 2)
	slog.Debug("getExtProviderIDFromTMDBURL: Parsing media path segment.",
		"segments", segs2)
	if len(segs2) != 2 {
		slog.Warn("getExtProviderIDFromTMDBURL: segs2 doesn't have len of 2.")
		return "", ""
	}

	return segments[0], segs2[0]
}

// Extract slug from IGDB url.
// Returns (Provider, ProviderID).
func (s *Service) getExtProviderIDFromIGDBURL(u *url.URL) (string, string) {
	segments := strings.Split(
		// Trim start/end '/' to avoid empty items at start/end
		// of final slice.
		strings.Trim(u.Path, "/"),
		"/",
	)
	segmentsLen := len(segments)
	slog.Debug("getExtProviderIDFromIMDBURL: Parsing path.",
		"segments", segments,
		"segments_len", segmentsLen)

	if segmentsLen < 2 ||
		segments[0] != "games" {
		slog.Debug("getExtProviderIDFromIMDBURL: path provided not supported.")
		return "", ""
	}

	return "igdb-slug", segments[1]
}
