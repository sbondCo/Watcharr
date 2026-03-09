package imprt

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
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
	SearchByExternalId(id string, source string) (tmdb.TMDBSearchMultiResponse, error)
}

type TagProvider interface {
	AddTag(userId uint, tr domain.TagAddRequest) (entity.Tag, error)
	GetTagByNameAndColor(userId uint, tagName string, tagColor string, tagBgColor string) (entity.Tag, error)
}

type SearchProvider interface {
	Search(r domain.SearchRequest, pp util.PaginationParams, userId uint) (domain.SearchResponse, error)
}

type Service struct {
	db               *gorm.DB
	wp               WatchedProvider
	wsp              WatchedSeasonProvider
	wep              WatchedEpisodeProvider
	cp               ContentProvider
	activityProvider domain.ActivityAddProvider
	tagProvider      TagProvider
	searchProvider   SearchProvider
}

func NewService(
	db *gorm.DB,
	wp WatchedProvider,
	wsp WatchedSeasonProvider,
	wep WatchedEpisodeProvider,
	cp ContentProvider,
	activityProvider domain.ActivityAddProvider,
	tagProvider TagProvider,
	searchProvider SearchProvider,
) *Service {
	return &Service{
		db,
		wp,
		wsp,
		wep,
		cp,
		activityProvider,
		tagProvider,
		searchProvider,
	}
}

func (s *Service) ImportContent(
	userId uint,
	ar domain.ImportRequest,
) (domain.ImportResponse, error) {
	slog.Debug("import: Processing request:", "request", ar)

	// If we have a TMDB ID given to us, we can go directly to
	// `SuccessfulImport` and let AddWatched fail if it doesn't exist.
	if ar.TmdbID != 0 {
		switch ar.Type {
		case domain.ImportContentTypeMovie:
			return s.SuccessfulImport(userId, &ar, domain.SuccessfulImportProps{
				TmdbID:      ar.TmdbID,
				ContentType: util.SupportedMediaMovie,
			}), nil
		case domain.ImportContentTypeShow:
			return s.SuccessfulImport(userId, &ar, domain.SuccessfulImportProps{
				TmdbID:      ar.TmdbID,
				ContentType: util.SupportedMediaShow,
			}), nil
		}
		// Unsupported types will continue to below...
	}

	// If imdb id passed, attempt to get content with it
	if ar.ImdbID != "" && (ar.Type == domain.ImportContentTypeMovie ||
		ar.Type == domain.ImportContentTypeShow ||
		ar.Type == domain.ImportContentTypeShowEpisode) {
		// Try import with imdb id.
		resp, err := s.importWithIMDBID(userId, &ar)
		if err == nil || !errors.Is(err, ErrNoResult) {
			return resp, err
		}
		// If ErrNoResult then we allow going below to search by name.
	}

	// If igdb provided, go straight to SuccessfulImport with it.
	if ar.IgdbID != 0 && (ar.Type == domain.ImportContentTypeGame) {
		return s.SuccessfulImport(userId, &ar, domain.SuccessfulImportProps{
			IgdbID:      ar.IgdbID,
			ContentType: util.SupportedMediaGame,
		}), nil
	}

	// If we have no IDs, run importWithName, which searches for content
	// by name.
	return s.importWithName(userId, &ar)
}

func (s *Service) SuccessfulImport(
	userId uint,
	ar *domain.ImportRequest,
	props domain.SuccessfulImportProps,
) domain.ImportResponse {
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
	// Build WatchedAddRequest
	wAddReq := domain.WatchedAddRequest{
		ContentType: props.ContentType,
		Status:      status,
		Rating:      ar.Rating,
		Thoughts:    ar.Thoughts,
		WatchedDate: wDate,
	}
	switch props.ContentType {
	case util.SupportedMediaMovie, util.SupportedMediaShow:
		wAddReq.TMDBID = props.TmdbID
	case util.SupportedMediaGame:
		wAddReq.IGDBID = props.IgdbID
	default:
		slog.Error("successfulImport: Invalid contentType provided!",
			"content_type", props.ContentType)
		return domain.ImportResponse{Type: domain.IMPORT_FAILED}
	}
	// Running add watched
	w, err := s.wp.AddWatched(
		userId,
		wAddReq,
		domain.WatchedAddExtraProps{
			ActivityType: entity.IMPORTED_WATCHED,
		})
	if err != nil {
		if errors.Is(err, domain.ErrWatchedExists) {
			slog.Error("successfulImport: Must already be on watch list", "error", err)
			return domain.ImportResponse{Type: domain.IMPORT_EXISTS}
		}
		slog.Error("successfulImport: Failed to add content as watched", "error", err)
		return domain.ImportResponse{Type: domain.IMPORT_FAILED}
	}
	// Add activity of the original time the show was added to the users
	// watchlist on whichever platform they are coming from.
	if ar.RatingCustomDate != nil {
		var addedActivity entity.Activity
		if len(w.Activity) > 0 {
			activityJson, _ := json.Marshal(map[string]any{
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
	return domain.ImportResponse{Type: domain.IMPORT_SUCCESS, WatchedEntry: w}
}
