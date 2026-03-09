package imprt

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/watched/episode"
	"github.com/sbondCo/Watcharr/util"
)

// Same Service as import.go but here so that one file doesn't get too hugemongus.

var (
	ErrNoResult = errors.New("no result from specific data request")
)

// Import by name (and or type). We do a content search to find a result,
// then try to import. Multiple results can end in returning them all for the
// user to decide.
func (s *Service) importWithName(
	userId uint,
	ar *domain.ImportRequest,
) (domain.ImportResponse, error) {
	// Try to search for results by name (and or type).
	searchReq := domain.SearchRequest{
		Query: ar.Name,
		// Default to multi search.
		Type: domain.SearchTypeMulti,
	}
	if ar.Type != "" {
		searchType := domain.ImportContentTypeToSearchType(ar.Type)
		if searchType == "" {
			// Invalid.. `ar.Type` is unsupported.
			slog.Error("importWithName: Invalid ImportContentType provided!",
				"type", ar.Type)
			return domain.ImportResponse{},
				errors.New("invalid import content type provided")
		}
		searchReq.Type = searchType
	}
	searchResp, err := s.searchProvider.Search(
		searchReq,
		util.PaginationParams{Page: 1},
		userId,
	)
	if err != nil {
		slog.Error("importWithName: Search failed", "error", err)
		return domain.ImportResponse{}, errors.New("search failed")
	}
	slog.Debug("importWithName: Potential matches",
		"num_found", searchResp.TotalResults)

	// If no results at all, return IMPORT_NOTFOUND
	if searchResp.TotalResults <= 0 {
		// No results found...
		slog.Debug("importWithName: returning IMPORT_NOTFOUND")
		return domain.ImportResponse{Type: domain.IMPORT_NOTFOUND}, nil
	}

	// If we did a multi search, remove people
	results := []domain.Media{}
	if searchReq.Type == domain.SearchTypeMulti {
		for _, v := range searchResp.Results {
			if v.Type != domain.MediaTypeTMDBPerson {
				results = append(results, v)
			}
		}
	} else {
		results = searchResp.Results
	}

	// Process results
	if len(results) > 1 {
		slog.Debug("importWithName: Multiple results found")
		return s.importWithNameHandleMultipleResultsFound(userId, ar, results)
	} else {
		slog.Debug("importWithName: success.. only found one result")
		props, err := domain.NewSuccessfulImportPropsFromMedia(&results[0])
		if err != nil {
			slog.Error("importWithName: Couldn't create props!", "error", err)
			return domain.ImportResponse{}, err
		}
		return s.SuccessfulImport(userId, ar, props), nil
	}
}

// If the importWithName func ends up with multiple results found in the
// search response, this will handle the case by trying to find a perfectMatch,
// otherwise returning an IMPORT_MULTI response.
func (s *Service) importWithNameHandleMultipleResultsFound(
	userId uint,
	ar *domain.ImportRequest,
	results []domain.Media,
) (domain.ImportResponse, error) {
	perfectMatches := []domain.Media{}
	for _, r := range results {
		itemReleaseYear := 0
		// Only parse dates to find year if the import request has provided
		// a year (and to keep below matching logic working properly).
		if ar.Year != 0 {
			if !r.ReleaseDate.IsZero() {
				itemReleaseYear = r.ReleaseDate.Year()
			} else {
				slog.Error("importWithName: Item has no ReleaseDate to use in comparison.")
			}
		}
		if strings.EqualFold(r.Name, ar.Name) {
			slog.Debug("importWithName: Found a perfect name match",
				"itemReleaseYear", itemReleaseYear,
				"ar.Year", ar.Year,
				"match", r)
			// If we have a year for comparison, force a check to compare them for a
			// match to be deemed perfect.
			// `itemReleaseYear` can only ever have a value if `ar.Year` has one, so this
			// check is safe as is.
			if itemReleaseYear != 0 || ar.Year != 0 {
				if itemReleaseYear == ar.Year {
					perfectMatches = append(perfectMatches, r)
					slog.Debug("importWithName: Name match also matched year")
				} else {
					slog.Debug("importWithName: Name match didn't match year")
				}
				continue
			}
			// Otherwise, if we don't have valid dates to compare, append the perfect name match anyways.
			slog.Debug("importWithName: Name match didn't have release years to compare, adding to matches anyways")
			perfectMatches = append(perfectMatches, r)
		}
	}

	// If one perfect match found, import it
	if len(perfectMatches) == 1 && perfectMatches[0].IDs.TMDB != 0 {
		slog.Debug("importWithName: importing from perfect match")
		props, err := domain.NewSuccessfulImportPropsFromMedia(&perfectMatches[0])
		if err != nil {
			slog.Error("importWithName: Couldn't create props!", "error", err)
			return domain.ImportResponse{}, err
		}
		return s.SuccessfulImport(userId, ar, props), nil
	}

	slog.Debug("importWithName: returning all potential matches")
	return domain.ImportResponse{Type: domain.IMPORT_MULTI, Results: results}, nil
}

// Import with an IMDb ID.
func (s *Service) importWithIMDBID(
	userId uint,
	ar *domain.ImportRequest,
) (domain.ImportResponse, error) {
	if imdbResp, err := s.cp.SearchByExternalId(ar.ImdbID, "imdb"); err == nil {
		if len(imdbResp.Results) == 1 {
			onlyResult := imdbResp.Results[0]
			if onlyResult.MediaType == string(entity.MOVIE) || onlyResult.MediaType == string(entity.SHOW) {
				// Will only be one result
				slog.Debug("import: importing imdb match", "imdb_id", ar.ImdbID, "tmdb_id_thatwasfound", onlyResult.ID)
				return s.SuccessfulImport(
					userId,
					ar,
					domain.SuccessfulImportProps{
						TmdbID:      onlyResult.ID,
						ContentType: util.SupportedMedia(onlyResult.MediaType),
					}), nil
			} else if onlyResult.MediaType == string(entity.SHOW_EPISODE) {
				// Handle episodes differently.
				// Clients must import tv episodes last so that the actual show can be imported first
				// will fail if watched entry isn't imported first or already exists (we won't make it here).
				w, e := s.wp.GetWatchedItemByTmdbId(userId, uint(onlyResult.ShowId), "tv")
				if e != nil {
					slog.Error("import: imdb match: Failed to add watched episode (failed to find watched item, it must exist!).", "rq", ar, "error", err)
					return domain.ImportResponse{Type: domain.IMPORT_FAILED}, nil
				}
				ws, err := s.wep.AddWatchedEpisodes(userId, episode.WatchedEpisodeAddRequest{
					WatchedID:       w.ID,
					SeasonNumber:    onlyResult.SeasonNumber,
					EpisodeNumber:   onlyResult.EpisodeNumber,
					Status:          ar.Status,
					Rating:          int8(ar.Rating),
					AddActivityDate: *ar.RatingCustomDate,
				})
				if err != nil {
					slog.Error("import: imdb match: Failed to add watched episode.", "rq", ar, "error", err)
					return domain.ImportResponse{Type: domain.IMPORT_FAILED}, nil
				} else {
					w.WatchedEpisodes = ws.WatchedEpisodes
					return domain.ImportResponse{Type: domain.IMPORT_SUCCESS, WatchedEntry: w}, nil
				}
			} else {
				slog.Error("import: imdb match has unsupported media type.", "media_type", imdbResp.Results[0].MediaType, "rq", ar)
				return domain.ImportResponse{Type: domain.IMPORT_FAILED}, nil
			}
		} else {
			// Content in tmdb may just be missing a related imdb id, so allow search to continue by name below.
			slog.Warn("import: No results for search by imdb id.. search will contiue by content name.", "rq", ar)
		}
	} else {
		slog.Warn("import: Failed to get content by imdb id.. search will contiue by content name.", "rq", ar)
	}
	// ErrNoResult should be caught from caller and let fall through to search by name.
	return domain.ImportResponse{}, ErrNoResult
}
