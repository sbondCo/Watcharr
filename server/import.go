package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ImportResponseType string

var (
	// Successful import
	IMPORT_SUCCESS ImportResponseType = "IMPORT_SUCCESS"
	// Import failed for reasons user cant fix
	IMPORT_FAILED ImportResponseType = "IMPORT_FAILED"
	// Import query returned multiple results, user must decide
	IMPORT_MULTI ImportResponseType = "IMPORT_MULTI"
	// Import query returned zero results, user must provide more info
	IMPORT_NOTFOUND ImportResponseType = "IMPORT_NOTFOUND"
	// Item already exists so couldn't import (unique constraint hit when adding)
	IMPORT_EXISTS ImportResponseType = "IMPORT_EXISTS"
)

type ImportRequest struct {
	Name             string           `json:"name"`
	Year             int              `json:"year"`
	TmdbID           int              `json:"tmdbId"`
	Type             ContentType      `json:"type"`
	Rating           float64          `json:"rating" binding:"max=10"`
	RatingCustomDate *time.Time       `json:"ratingCustomDate"`
	Status           WatchedStatus    `json:"status"`
	Thoughts         string           `json:"thoughts"`
	DatesWatched     []time.Time      `json:"datesWatched"`
	Activity         []Activity       `json:"activity"`
	WatchedEpisodes  []WatchedEpisode `json:"watchedEpisodes"`
	WatchedSeason    []WatchedSeason  `json:"watchedSeasons"`
	Tags             []TagAddRequest  `json:"tags"`
	ImdbID           string           `json:"imdbId"`
}

type ImportResponse struct {
	Type    ImportResponseType       `json:"type"`
	Results []TMDBSearchMultiResults `json:"results"`
	Match   TMDBSearchMultiResults   `json:"match"`
	// On success this will be filled with the new watched entry
	WatchedEntry Watched `json:"watchedEntry"`
}

func importContent(db *gorm.DB, userId uint, ar ImportRequest) (ImportResponse, error) {
	slog.Debug("import: Processing request:", "request", ar)
	// If tmdbId and type passed in request body
	// we dont need to use a search tmdb request.
	// Retrieve the details directly.
	if ar.TmdbID != 0 && (ar.Type == MOVIE || ar.Type == SHOW) {
		tid := strconv.Itoa(ar.TmdbID)
		if ar.Type == MOVIE {
			cr, err := movieDetails(db, tid, "", map[string]string{})
			if err != nil {
				return ImportResponse{}, errors.New("movie details request failed")
			}
			slog.Debug("import: by tmdbid of movie", "cr", cr)
			return successfulImport(db, userId, cr.ID, MOVIE, ar)
		} else if ar.Type == SHOW {
			cr, err := tvDetails(db, tid, "", map[string]string{})
			if err != nil {
				return ImportResponse{}, errors.New("tv details request failed")
			}
			slog.Debug("import: by tmdbid of tv", "cr", cr)
			return successfulImport(db, userId, cr.ID, SHOW, ar)
		}
	}
	// If imdb id passed, attempt to get content with it
	if ar.ImdbID != "" && (ar.Type == MOVIE || ar.Type == SHOW || ar.Type == SHOW_EPISODE) {
		if imdbResp, err := searchByExternalId(ar.ImdbID, "imdb"); err == nil {
			if len(imdbResp.Results) == 1 {
				onlyResult := imdbResp.Results[0]
				if onlyResult.MediaType == string(MOVIE) || onlyResult.MediaType == string(SHOW) {
					// Will only be one result
					slog.Debug("import: importing imdb match", "imdb_id", ar.ImdbID, "tmdb_id_thatwasfound", onlyResult.ID)
					return successfulImport(db, userId, onlyResult.ID, ContentType(onlyResult.MediaType), ar)
				} else if onlyResult.MediaType == string(SHOW_EPISODE) {
					// Handle episodes differently.
					// Clients must import tv episodes last so that the actual show can be imported first
					// will fail if watched entry isn't imported first or already exists (we won't make it here).
					w, e := getWatchedItemByTmdbId(db, userId, uint(onlyResult.ShowId), "tv")
					if e != nil {
						slog.Error("import: imdb match: Failed to add watched episode (failed to find watched item, it must exist!).", "rq", ar, "error", err)
						return ImportResponse{Type: IMPORT_FAILED}, nil
					}
					ws, err := addWatchedEpisodes(db, userId, WatchedEpisodeAddRequest{
						WatchedID:       w.ID,
						SeasonNumber:    onlyResult.SeasonNumber,
						EpisodeNumber:   onlyResult.EpisodeNumber,
						Status:          ar.Status,
						Rating:          int8(ar.Rating),
						addActivityDate: *ar.RatingCustomDate,
					})
					if err != nil {
						slog.Error("import: imdb match: Failed to add watched episode.", "rq", ar, "error", err)
						return ImportResponse{Type: IMPORT_FAILED}, nil
					} else {
						w.WatchedEpisodes = ws.WatchedEpisodes
						return ImportResponse{Type: IMPORT_SUCCESS, WatchedEntry: w}, nil
					}
				} else {
					slog.Error("import: imdb match has unsupported media type.", "media_type", imdbResp.Results[0].MediaType, "rq", ar)
					return ImportResponse{Type: IMPORT_FAILED}, nil
				}
			} else {
				// Content in tmdb may just be missing a related imdb id, so allow search to continue by name below.
				slog.Warn("import: No results for search by imdb id.. search will contiue by content name.", "rq", ar)
			}
		} else {
			slog.Warn("import: Failed to get content by imdb id.. search will contiue by content name.", "rq", ar)
		}
	}
	// tmdbId not passed.. search for the content by name.
	sr, err := searchContent(ar.Name, 1)
	if err != nil {
		slog.Error("import: content search failed", "error", err)
		return ImportResponse{}, errors.New("Content search failed")
	}
	pMatches := []TMDBSearchMultiResults{}
	for _, r := range sr.Results {
		if r.MediaType != "person" {
			pMatches = append(pMatches, r)
		}
	}
	resLen := len(pMatches)
	slog.Debug("import: potential matches", "num_found", resLen)
	if resLen <= 0 {
		slog.Debug("import: returning IMPORT_NOTFOUND")
		return ImportResponse{Type: IMPORT_NOTFOUND}, nil
	} else if resLen > 1 {
		slog.Debug("import: multiple results found")
		// If there are multiple responses, but only one item
		// from the results is a 100% match for the imported
		// items name, then consider successful match with that.
		perfectMatches := []TMDBSearchMultiResults{}
		for _, r := range pMatches {
			itemName := r.Name
			if itemName == "" {
				itemName = r.Title
			}
			itemReleaseYear := 0
			// Only parse dates to find year if the import request has provided
			// a year to comparisons.. otherwise don't do it to save some performance juice.
			if ar.Year != 0 {
				itemReleaseDateStr := r.ReleaseDate
				if itemReleaseDateStr == "" {
					itemReleaseDateStr = r.FirstAirDate
				}
				if itemReleaseDate, err := time.Parse("2006-01-02", itemReleaseDateStr); err == nil {
					itemReleaseYear = itemReleaseDate.Year()
				} else {
					slog.Error("import: failed to check item release year, it can't be used for matching", "error", err, "item", r)
				}
			}
			if strings.EqualFold(itemName, ar.Name) {
				slog.Debug("import: multiple results processing: found a perfect name match", "itemReleaseYear", itemReleaseYear, "ar.Year", ar.Year, "match", r)
				// If we have a year for comparison, force a check to compare them for a
				// match to be deemed perfect.
				// `itemReleaseYear` can only ever have a value if `ar.Year` has one, so this
				// check is safe as is.
				if itemReleaseYear != 0 || ar.Year != 0 {
					if itemReleaseYear == ar.Year {
						perfectMatches = append(perfectMatches, r)
						slog.Debug("import: multiple results processing: name match also matched year")
					} else {
						slog.Debug("import: multiple results processing: name match didnt match year")
					}
					continue
				}
				// Otherwise, if we don't have valid dates to compare, append the perfect name match anyways.
				slog.Debug("import: multiple results processing: name match didn't have valid release year, adding to matches anyways")
				perfectMatches = append(perfectMatches, r)
			}
		}
		// If one perfect match found, import it
		pmLen := len(perfectMatches)
		if pmLen == 1 && perfectMatches[0].ID != 0 {
			slog.Debug("import: importing from perfect match")
			return successfulImport(db, userId, perfectMatches[0].ID, ContentType(perfectMatches[0].MediaType), ar)
		} else if pmLen > 1 {
			slog.Debug("import: returning multiple perfect matches")
			return ImportResponse{Type: IMPORT_MULTI, Results: perfectMatches}, nil
		}
		slog.Debug("import: returning all potential matches")
		return ImportResponse{Type: IMPORT_MULTI, Results: pMatches}, nil
	} else {
		slog.Debug("import: success.. only found one result")
		return successfulImport(db, userId, pMatches[0].ID, ContentType(pMatches[0].MediaType), ar)
	}
}

func successfulImport(db *gorm.DB, userId uint, contentId int, contentType ContentType, ar ImportRequest) (ImportResponse, error) {
	status := FINISHED
	if ar.Status != "" {
		status = ar.Status
	}
	// Get the latest date from DatesWatched if we have any.
	var wDate time.Time
	if len(ar.DatesWatched) > 0 {
		for _, dw := range ar.DatesWatched {
			if dw.After(wDate) {
				wDate = dw
			}
		}
	}
	w, err := addWatched(db, userId, WatchedAddRequest{
		Status:      status,
		ContentID:   contentId,
		ContentType: contentType,
		Rating:      ar.Rating,
		Thoughts:    ar.Thoughts,
		WatchedDate: wDate,
	}, IMPORTED_WATCHED)
	if err != nil {
		if err.Error() == "content already on watched list" {
			slog.Error("successfulImport: unique constraint hit.. show must already be on watch list", "error", err)
			return ImportResponse{Type: IMPORT_EXISTS}, nil
		}
		slog.Error("successfulImport: Failed to add content as watched", "error", err)
		return ImportResponse{Type: IMPORT_FAILED}, nil
	}
	// Add activity of the original time the show was added to the users watchlist on whichever platform they are coming from.
	if ar.RatingCustomDate != nil {
		var addedActivity Activity
		if len(w.Activity) > 0 {
			activityJson, _ := json.Marshal(map[string]interface{}{"rating": ar.Rating, "linkedActivity": w.Activity[0].ID})
			addedActivity, _ = addActivity(db, userId, ActivityAddRequest{WatchedID: w.ID, Type: IMPORTED_RATING, Data: string(activityJson), CustomDate: ar.RatingCustomDate})
		} else {
			addedActivity, _ = addActivity(db, userId, ActivityAddRequest{WatchedID: w.ID, Type: IMPORTED_RATING, Data: strconv.Itoa(int(ar.Rating)), CustomDate: ar.RatingCustomDate})
		}
		w.Activity = append(w.Activity, addedActivity)
	}
	// Add all dates watched as activity, if any
	if len(ar.DatesWatched) > 0 {
		for _, v := range ar.DatesWatched {
			customDate := v
			addedActivity, err := addActivity(db, userId, ActivityAddRequest{WatchedID: w.ID, Type: IMPORTED_ADDED_WATCHED, CustomDate: &customDate})
			if err == nil {
				w.Activity = append(w.Activity, addedActivity)
			} else {
				slog.Error("successfulImport: Failed to add dateswatched activity.", "date", v, "error", err)
			}
		}
	}
	// Add all activity passed in.
	// Probably was is a Watcharr export being imported, so it'll have all it's activity too.
	if len(ar.Activity) > 0 {
		slog.Debug("successfulImport: Importing activity")
		for i, v := range ar.Activity {
			activityDate := ar.Activity[i].CustomDate
			if activityDate == nil || activityDate.IsZero() {
				activityDate = &ar.Activity[i].CreatedAt
			}
			addedActivity, err := addActivity(db, userId, ActivityAddRequest{WatchedID: w.ID, Type: v.Type, Data: v.Data, CustomDate: activityDate})
			if err == nil {
				w.Activity = append(w.Activity, addedActivity)
			} else {
				slog.Error("successfulImport: Failed to add imported activity.", "full_object", v, "error", err)
			}
		}
	}
	// Import watched seasons, if any
	if len(ar.WatchedSeason) > 0 {
		slog.Debug("successfulImport: Importing watched seasons")
		for _, v := range ar.WatchedSeason {
			ws, err := addWatchedSeason(db, userId, WatchedSeasonAddRequest{
				WatchedID:       w.ID,
				SeasonNumber:    v.SeasonNumber,
				Status:          v.Status,
				Rating:          v.Rating,
				addActivityDate: v.CreatedAt,
			})
			if err != nil {
				slog.Error("successfulImport: Failed to add watched season.", "error", err)
				continue
			}
			w.WatchedSeasons = ws.WatchedSeasons
		}
	}
	// Import watched episodes, if any
	if len(ar.WatchedEpisodes) > 0 {
		slog.Debug("successfulImport: Importing watched episodes")
		for _, v := range ar.WatchedEpisodes {
			ws, err := addWatchedEpisodes(db, userId, WatchedEpisodeAddRequest{
				WatchedID:       w.ID,
				SeasonNumber:    v.SeasonNumber,
				EpisodeNumber:   v.EpisodeNumber,
				Status:          v.Status,
				Rating:          v.Rating,
				addActivityDate: v.CreatedAt,
			})
			if err != nil {
				slog.Error("successfulImport: Failed to add watched episodes.", "error", err)
				continue
			}
			w.WatchedEpisodes = ws.WatchedEpisodes
		}
	}
	// Import tags, if any
	if len(ar.Tags) > 0 {
		// Create tags if they dont exist
		slog.Debug("successfulImport: Importing tags")
		for _, v := range ar.Tags {
			// Check if tag exists
			var t Tag
			t, err := getTagByNameAndColor(db, userId, v.Name, v.Color, v.BgColor)
			if err != nil && err.Error() != "tag does not exist" {
				slog.Error("successfulImport: Failed to check for an existing tag", "name", v.Name, "error", err)
				continue
			}
			if t.ID == 0 {
				tag, err := addTag(db, userId, TagAddRequest{
					Name:    v.Name,
					Color:   v.Color,
					BgColor: v.BgColor,
				})
				if err != nil {
					slog.Error("successfulImport: Failed to add a tag.", "name", v.Name, "error", err)
					continue
				}
				t = tag
			}

			// Associate the watched entry with the tag
			err = addWatchedToTag(db, userId, t.ID, w.ID)
			if err != nil {
				slog.Error("successfulImport: Failed to associate watched entry with tag.", "error", err)
				continue
			}
			w.Tags = append(w.Tags, t)
		}
	}
	return ImportResponse{Type: IMPORT_SUCCESS, WatchedEntry: w}, nil
}
