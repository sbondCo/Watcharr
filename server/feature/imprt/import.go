package imprt

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/watched"
	"github.com/sbondCo/Watcharr/feature/watched/episode"
	"github.com/sbondCo/Watcharr/feature/watched/season"
	"github.com/sbondCo/Watcharr/media/tmdb"
	"github.com/sbondCo/Watcharr/util"
	"gorm.io/gorm"
)

type WatchedProvider interface {
	AddWatched(userId uint, ar domain.WatchedAddRequest, extraProps domain.WatchedAddExtraProps) (entity.Watched, error)
	GetWatchedItemByTmdbId(userId uint, tmdbId uint, contentType entity.ContentType) (entity.Watched, error)
}

type WatchedSeasonProvider interface {
	AddWatchedSeason(userId uint, ar season.WatchedSeasonAddRequest) (season.WatchedSeasonAddResponse, error)
}

type WatchedEpisodeProvider interface {
	AddWatchedEpisodes(userId uint, ar episode.WatchedEpisodeAddRequest) (episode.WatchedEpisodeAddResponse, error)
}

type ContentProvider interface {
	SearchContent(query string, pageNum int) (tmdb.TMDBSearchMultiResponse, error)
	SearchByExternalId(id string, source string) (tmdb.TMDBSearchMultiResponse, error)
	MovieDetails(id string, country string, rParams map[string]string) (tmdb.TMDBMovieDetails, error)
	TvDetails(id string, country string, rParams map[string]string) (tmdb.TMDBShowDetails, error)
}

type TagProvider interface {
	AddTag(userId uint, tr domain.TagAddRequest) (entity.Tag, error)
	GetTagByNameAndColor(userId uint, tagName string, tagColor string, tagBgColor string) (entity.Tag, error)
}

type Service struct {
	db               *gorm.DB
	wp               WatchedProvider
	wsp              WatchedSeasonProvider
	wep              WatchedEpisodeProvider
	cp               ContentProvider
	activityProvider domain.ActivityAddProvider
	tagProvider      TagProvider
}

func NewService(
	db *gorm.DB,
	wp WatchedProvider,
	wsp WatchedSeasonProvider,
	wep WatchedEpisodeProvider,
	cp ContentProvider,
	activityProvider domain.ActivityAddProvider,
	tagProvider TagProvider,
) *Service {
	return &Service{
		db,
		wp,
		wsp,
		wep,
		cp,
		activityProvider,
		tagProvider,
	}
}

// TODO Support game importing

func (s *Service) ImportContent(
	userId uint,
	ar domain.ImportRequest,
) (domain.ImportResponse, error) {
	slog.Debug("import: Processing request:", "request", ar)
	// If tmdbId and type passed in request body
	// we dont need to use a search tmdb request.
	// Retrieve the details directly.
	if ar.TmdbID != 0 && (ar.Type == entity.MOVIE || ar.Type == entity.SHOW) {
		tid := strconv.Itoa(ar.TmdbID)
		if ar.Type == entity.MOVIE {
			cr, err := s.cp.MovieDetails(tid, "", map[string]string{})
			if err != nil {
				return domain.ImportResponse{}, errors.New("movie details request failed")
			}
			slog.Debug("import: by tmdbid of movie", "cr", cr)
			return s.SuccessfulImport(userId, cr.ID, util.SupportedMediaMovie, ar)
		} else if ar.Type == entity.SHOW {
			cr, err := s.cp.TvDetails(tid, "", map[string]string{})
			if err != nil {
				return domain.ImportResponse{}, errors.New("tv details request failed")
			}
			slog.Debug("import: by tmdbid of tv", "cr", cr)
			return s.SuccessfulImport(userId, cr.ID, util.SupportedMediaShow, ar)
		}
	}
	// If imdb id passed, attempt to get content with it
	if ar.ImdbID != "" && (ar.Type == entity.MOVIE || ar.Type == entity.SHOW || ar.Type == entity.SHOW_EPISODE) {
		if imdbResp, err := s.cp.SearchByExternalId(ar.ImdbID, "imdb"); err == nil {
			if len(imdbResp.Results) == 1 {
				onlyResult := imdbResp.Results[0]
				if onlyResult.MediaType == string(entity.MOVIE) || onlyResult.MediaType == string(entity.SHOW) {
					// Will only be one result
					slog.Debug("import: importing imdb match", "imdb_id", ar.ImdbID, "tmdb_id_thatwasfound", onlyResult.ID)
					return s.SuccessfulImport(userId, onlyResult.ID, util.SupportedMedia(onlyResult.MediaType), ar)
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
	}
	// tmdbId not passed.. search for the content by name.
	sr, err := s.cp.SearchContent(ar.Name, 1)
	if err != nil {
		slog.Error("import: content search failed", "error", err)
		return domain.ImportResponse{}, errors.New("content search failed")
	}
	// potential matches
	pMatches := []domain.Media{}
	for _, r := range sr.Results {
		if r.MediaType != "person" {
			pMatches = append(pMatches, r.AsMedia())
		}
	}
	resLen := len(pMatches)
	slog.Debug("import: potential matches", "num_found", resLen)
	if resLen <= 0 {
		slog.Debug("import: returning IMPORT_NOTFOUND")
		return domain.ImportResponse{Type: domain.IMPORT_NOTFOUND}, nil
	} else if resLen > 1 {
		slog.Debug("import: multiple results found")
		// If there are multiple responses, but only one item
		// from the results is a 100% match for the imported
		// items name, then consider successful match with that.
		perfectMatches := []domain.Media{}
		for _, r := range pMatches {
			itemReleaseYear := 0
			// Only parse dates to find year if the import request has provided
			// a year to comparisons.. otherwise don't do it to save some performance juice.
			if ar.Year != 0 {
				if !r.ReleaseDate.IsZero() {
					itemReleaseYear = r.ReleaseDate.Year()
				} else {
					slog.Error("import: failed to check item release year, it can't be used for matching",
						"error", err, "item", r)
				}
			}
			if strings.EqualFold(r.Name, ar.Name) {
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
		if pmLen == 1 && perfectMatches[0].IDs.TMDB != 0 {
			slog.Debug("import: importing from perfect match")
			return s.SuccessfulImport(
				userId,
				perfectMatches[0].IDs.TMDB,
				perfectMatches[0].GetMediaType(),
				ar)
		}
		slog.Debug("import: returning all potential matches")
		return domain.ImportResponse{Type: domain.IMPORT_MULTI, Results: pMatches}, nil
	} else {
		slog.Debug("import: success.. only found one result")
		return s.SuccessfulImport(
			userId,
			pMatches[0].IDs.TMDB,
			pMatches[0].GetMediaType(),
			ar)
	}
}

func (s *Service) SuccessfulImport(
	userId uint,
	contentId int,
	contentType util.SupportedMedia,
	ar domain.ImportRequest,
) (domain.ImportResponse, error) {
	status := entity.FINISHED
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
	w, err := s.wp.AddWatched(
		userId,
		domain.WatchedAddRequest{
			Status:      status,
			TMDBID:      contentId,
			ContentType: contentType,
			Rating:      ar.Rating,
			Thoughts:    ar.Thoughts,
			WatchedDate: wDate,
		},
		domain.WatchedAddExtraProps{
			ActivityType: entity.IMPORTED_WATCHED,
		})
	if err != nil {
		if errors.Is(err, domain.ErrWatchedExists) {
			slog.Error("successfulImport: Must already be on watch list", "error", err)
			return domain.ImportResponse{Type: domain.IMPORT_EXISTS}, nil
		}
		slog.Error("successfulImport: Failed to add content as watched", "error", err)
		return domain.ImportResponse{Type: domain.IMPORT_FAILED}, nil
	}
	// Add activity of the original time the show was added to the users
	// watchlist on whichever platform they are coming from.
	if ar.RatingCustomDate != nil {
		var addedActivity entity.Activity
		if len(w.Activity) > 0 {
			activityJson, _ := json.Marshal(map[string]interface{}{
				"rating":         ar.Rating,
				"linkedActivity": w.Activity[0].ID,
			})
			addedActivity, _ = s.activityProvider.AddActivity(userId, domain.ActivityAddRequest{WatchedID: w.ID, Type: entity.IMPORTED_RATING, Data: string(activityJson), CustomDate: ar.RatingCustomDate})
		} else {
			addedActivity, _ = s.activityProvider.AddActivity(userId, domain.ActivityAddRequest{WatchedID: w.ID, Type: entity.IMPORTED_RATING, Data: strconv.Itoa(int(ar.Rating)), CustomDate: ar.RatingCustomDate})
		}
		w.Activity = append(w.Activity, addedActivity)
	}
	// Add all dates watched as activity, if any
	if len(ar.DatesWatched) > 0 {
		for _, v := range ar.DatesWatched {
			customDate := v
			addedActivity, err := s.activityProvider.AddActivity(userId, domain.ActivityAddRequest{WatchedID: w.ID, Type: entity.IMPORTED_ADDED_WATCHED, CustomDate: &customDate})
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
			addedActivity, err := s.activityProvider.AddActivity(userId, domain.ActivityAddRequest{WatchedID: w.ID, Type: v.Type, Data: v.Data, CustomDate: activityDate})
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
			ws, err := s.wsp.AddWatchedSeason(userId, season.WatchedSeasonAddRequest{
				WatchedID:       w.ID,
				SeasonNumber:    v.SeasonNumber,
				Status:          v.Status,
				Rating:          v.Rating,
				AddActivityDate: v.CreatedAt,
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
			ws, err := s.wep.AddWatchedEpisodes(userId, episode.WatchedEpisodeAddRequest{
				WatchedID:       w.ID,
				SeasonNumber:    v.SeasonNumber,
				EpisodeNumber:   v.EpisodeNumber,
				Status:          v.Status,
				Rating:          v.Rating,
				AddActivityDate: v.CreatedAt,
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
			var t entity.Tag
			t, err := s.tagProvider.GetTagByNameAndColor(userId, v.Name, v.Color, v.BgColor)
			if err != nil && err.Error() != "tag does not exist" {
				slog.Error("successfulImport: Failed to check for an existing tag", "name", v.Name, "error", err)
				continue
			}
			if t.ID == 0 {
				tag, err := s.tagProvider.AddTag(userId, domain.TagAddRequest{
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
			err = watched.AddWatchedToTag(s.db, userId, t.ID, w.ID)
			if err != nil {
				slog.Error("successfulImport: Failed to associate watched entry with tag.", "error", err)
				continue
			}
			w.Tags = append(w.Tags, t)
		}
	}
	return domain.ImportResponse{Type: domain.IMPORT_SUCCESS, WatchedEntry: w}, nil
}
