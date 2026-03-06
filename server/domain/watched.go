package domain

import (
	"errors"
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/feature/watched/watchedutil"
	"github.com/sbondCo/Watcharr/util"
)

var (
	// Watched entry already exists.
	ErrWatchedExists = errors.New("watched entry exists")
	// Watched entry exists as a soft deleted record (..and restore wasn't attempted by choice).
	ErrWatchedExistsSoftDeleted = errors.New("watched entry exists soft deleted")
)

type WatchedSort string

const (
	WatchedSortDateAdded    WatchedSort = "DATEADDED"
	WatchedSortLastChanged  WatchedSort = "LASTCHANGED"
	WatchedSortLastFinished WatchedSort = "LASTFIN"
	WatchedSortRating       WatchedSort = "RATING"
	WatchedSortAlphabetical WatchedSort = "ALPHA"
	WatchedSortDateReleased WatchedSort = "DATERELEASED"
)

type SortDirection string

const (
	WatchedSortDirAsc  SortDirection = "asc"
	WatchedSortDirDesc SortDirection = "desc"
)

// Get watched page request extra (GET) options.
// Since this is user input, validity of string types cannot be guaranteed.
type WatchedGetPageRequest struct {
	// Sorting type.
	Sort WatchedSort `form:"sort"`
	// Sorting direction (asc or desc).
	SortDir SortDirection `form:"sortDir,default=desc"`
	// Filtering options.
	FilterType   []util.SupportedMedia  `form:"type" collection_format:"csv"`
	FilterStatus []entity.WatchedStatus `form:"status" collection_format:"csv"`
}

type WatchedGetPageExtraProps struct {
	// Only get these watched ids.
	WatchedIds []int
	// Only get watched items where content matches this query.
	Query string
}

type WatchedAddExtraProps struct {
	// The type of activity this is (will be added to db as this type).
	ActivityType entity.ActivityType
	// When watched entry to db, if a unique constraint is hit, should
	// we skip the restoration logic (bring back soft deleted record)?
	// - Syncing logic may use this field as it might not make sense to restore
	// a record when doing a task like a regular sync, otherwise the user could
	// soft delte a record and then have it come back when they next sync.
	// - Importers however probably shouldn't use this flag as true because
	// importing will always be a manual decision.
	DontRestore bool
}

type WatchedDto struct {
	// Properties that always exist in every watched dto below.

	ID        uint                 `json:"id"`
	CreatedAt time.Time            `json:"createdAt"`
	UpdatedAt time.Time            `json:"updatedAt"`
	Status    entity.WatchedStatus `json:"status"`
	Rating    float64              `json:"rating"`
	Pinned    bool                 `json:"pinned"`

	// Properties that may not be included in all watched dtos
	// (depending on where we are making the dto for)

	// Thoughts aren't always included (not needed on watched list pages
	// & on public pages because they could be private).
	Thoughts string `json:"thoughts,omitempty"`
	// Watching Season extra detail for list.
	WatchingSeason   string                  `json:"watchingSeason,omitempty"`
	Activity         []entity.Activity       `json:"activity,omitempty"`
	WatchedSeasons   []entity.WatchedSeason  `json:"watchedSeasons,omitempty"`
	WatchedEpisodes  []entity.WatchedEpisode `json:"watchedEpisodes,omitempty"`
	Tags             []entity.Tag            `json:"tags,omitempty"`
	LastViewedSeason *int                    `json:"lastViewedSeason,omitempty"`
}

// New dto with base properties that we have for all WatchedDtos.
// Note: If this is updated, ensure whatever uses this still makes sense.
func NewWatchedDtoWithBaseProps(w *entity.Watched) WatchedDto {
	return WatchedDto{
		ID:        w.ID,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
		Status:    w.Status,
		Rating:    w.Rating,
		Pinned:    w.Pinned,
	}
}

func NewWatchedDtoForLists(w *entity.Watched) WatchedDto {
	dto := NewWatchedDtoWithBaseProps(w)

	if w.Content != nil && w.Content.Type == entity.SHOW {
		dto.WatchingSeason = watchedutil.GetLatestWatchedInTv(
			w.WatchedSeasons, w.WatchedEpisodes)
	}

	return dto
}

// For public lists showing other users watched data.
func NewWatchedDtoForPublicLists(w *entity.Watched) WatchedDto {
	dto := NewWatchedDtoWithBaseProps(w)

	if w.Content != nil && w.Content.Type == entity.SHOW {
		dto.WatchingSeason = watchedutil.GetLatestWatchedInTv(
			w.WatchedSeasons, w.WatchedEpisodes)
	}

	return dto
}

// A fuller dto with all details needed for a content details page.
func NewWatchedDtoForContentPage(w *entity.Watched) WatchedDto {
	dto := NewWatchedDtoWithBaseProps(w)

	dto.Thoughts = w.Thoughts
	dto.Activity = w.Activity
	dto.WatchedSeasons = w.WatchedSeasons
	dto.WatchedEpisodes = w.WatchedEpisodes
	dto.Tags = w.Tags
	dto.LastViewedSeason = w.LastViewedSeason

	return dto
}

// Get our watched page response.
type WatchedGetPageResponse []Media

// Used for GetWatchedPage (and GetWatchedPage for our search).
func NewWatchedGetPageResponse(w []entity.Watched) WatchedGetPageResponse {
	r := WatchedGetPageResponse{}
	for i := range w {
		v := &w[i]
		d := NewWatchedDtoForLists(v)
		r = append(r, NewMediaFromWatched(v, &d))
	}
	return r
}

// Get a public users list response.
type WatchedPublicGetPageResponse []Media

func NewWatchedPublicGetPageResponse(w []entity.Watched) WatchedPublicGetPageResponse {
	r := WatchedPublicGetPageResponse{}
	for i := range w {
		v := &w[i]
		d := NewWatchedDtoForPublicLists(v)
		r = append(r, NewMediaFromWatched(v, &d))
	}
	return r
}

// Add a watched entry request
type WatchedAddRequest struct {
	// Type of content we are adding to watched.
	ContentType util.SupportedMedia `json:"contentType" binding:"required,oneof=movie tv game"`
	// ID of content from tmdb (if ContentType is movie or tv).
	TMDBID int `json:"tmdbId"`
	// DEPRECATED!! This will be removed soon, I've left it in only so any third
	// party scripts can migrate pain-free. This will work as if you have passed
	// the id for `tmdbId`, please replace 'contentId' properties in requests
	// with `tmdbId`.
	Deprecated_ContentID int `json:"contentId"`
	// ID of content from igdb (if ContentType is game).
	IGDBID int `json:"igdbId"`

	Status   entity.WatchedStatus `json:"status"`
	Rating   float64              `json:"rating" binding:"max=10"`
	Thoughts string               `json:"thoughts"`
	// Pass a watched date and we will set the CreatedAt (and initial UpdatedAt)
	// properties for this watched entry to this specific date.
	WatchedDate time.Time `json:"watchedDate,omitempty"`
}

// Update watched entry request
type WatchedUpdateRequest struct {
	Status         entity.WatchedStatus `json:"status" binding:"required_without_all=Rating Thoughts RemoveThoughts Pinned"`
	Rating         float64              `json:"rating" binding:"max=10,required_without_all=Status Thoughts RemoveThoughts Pinned"`
	Thoughts       string               `json:"thoughts" binding:"required_without_all=Status Rating RemoveThoughts Pinned"`
	RemoveThoughts bool                 `json:"removeThoughts"`
	Pinned         *bool                `json:"pinned" binding:"required_without_all=Status Rating Thoughts RemoveThoughts"`
}

// Update response.
type WatchedUpdateResponse struct {
	NewActivity entity.Activity `json:"newActivity"`
}

// Removal response.
type WatchedRemoveResponse struct {
	NewActivity entity.Activity `json:"newActivity"`
}
