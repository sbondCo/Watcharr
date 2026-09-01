// Types that we can use for all content types (movie, tv, game, book, everything).
// Data responses to the client can use these "uniform" types to make access
// easier.

package domain

import (
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/util"
)

type MediaType string

const (
	MediaTypeTMDBMovie  MediaType = "tmdb_movie"
	MediaTypeTMDBShow   MediaType = "tmdb_tv"
	MediaTypeTMDBPerson MediaType = "tmdb_person"

	MediaTypeIGDBGame MediaType = "igdb_game"

	MediaTypeOLBook       MediaType = "ol_book"
	MediaTypeOLBookAuthor MediaType = "ol_book_author"
)

type Media struct {
	// The type of media.
	Type MediaType `json:"type,omitempty"`
	// The ids associated with this media.
	IDs MediaIDs `json:"ids"`
	// The name of the media.
	Name string `json:"name,omitempty"`
	// A description.
	Summary string `json:"summary,omitempty"`
	// The poster.
	Poster *entity.Image `json:"poster,omitempty"`
	// The external poster path.
	ExtPosterPath string `json:"extPosterPath,omitempty"`
	// The rating.
	Rating uint `json:"rating,omitempty"`
	// The amount of votes that made up the rating.
	RatingCount uint `json:"ratingCount,omitempty"`
	// Watched data.
	Watched WatchedDto `json:"watched,omitzero"`
	// Similar media.
	Similar []Media `json:"similar,omitempty"`
	// Release date / first air date.
	ReleaseDate time.Time `json:"releaseDate,omitzero"`
	// Videos (trailers, etc)
	Videos []MediaVideo `json:"videos,omitempty"`

	//
	// Properties that are less important (not used for all responses).
	//

	// Backdrop path.
	ExtBackdropPath string `json:"extBackdropPath,omitempty"`
	// Genres.
	Genres []MediaGenre `json:"genres,omitempty"`
	// Media website.
	Homepage string `json:"homepage,omitempty"`
	// Media providers (eg Streaming sites, game markets)
	Providers []MediaProvider `json:"providers,omitempty"`
	// A link to the database we are using that lists all providers with max details.
	// (especially for TMDB since it's data from JustWatch isn't available to us).
	ProvidersFullListLink string `json:"providersFullListLink,omitempty"`
	// Status of the media (released, ended, etc).
	// Depending on media type, this will contain different values.
	Status string `json:"status,omitempty"`

	//
	// Properties only for movies/tv.
	//

	// Runtime.
	Runtime uint `json:"runtime,omitempty"`
	// Seasons.
	Seasons []MediaSeason `json:"seasons,omitempty"`
	// Simple bool for our RequestShow component since Sonarr can be given a
	// series type (ideally the frontend doesn't need to do that, but for now it
	// does.. if (son)arr code is refactored, can the client just pass very basic
	// details for the server to fetch fully/verify, i.e fetched full details from
	// tmdb again to verify if show is anime itself, etc).
	IsShowAnime bool `json:"isShowAnime,omitempty"`
	// Last release date.
	// Currently used for tv shows so frontend can display its end date.
	ReleaseDateLast time.Time `json:"releaseDateLast,omitzero"`

	//
	// Properties only for Games
	//

	// Game modes.
	GameModes []MediaGenre `json:"gameModes,omitempty"`

	//
	// Properties only for Books
	//

	// Authors.
	Authors []MediaPerson `json:"authors,omitempty"`
}

func (t Media) GetId() int {
	switch t.Type {
	case MediaTypeTMDBMovie,
		MediaTypeTMDBShow:
		return t.IDs.TMDB
	case MediaTypeIGDBGame:
		return t.IDs.IGDB
	case MediaTypeOLBook:
		// OL12345W -> 12345 (for compatibility with existing APIs)
		// in the future, if other book providers will be added, GetId likely
		// has to be refactored to return a string instead
		id, err := strconv.Atoi(t.IDs.OLID[2 : len(t.IDs.OLID)-1])
		if err != nil {
			return -99
		}
		return id
	}
	return -99
}

// If this changes, verify all use cases still make sense!
func (t Media) GetMediaType() util.SupportedMedia {
	switch t.Type {
	case MediaTypeTMDBMovie:
		return util.SupportedMediaMovie
	case MediaTypeTMDBShow:
		return util.SupportedMediaShow
	case MediaTypeIGDBGame:
		return util.SupportedMediaGame
	case MediaTypeOLBook:
		return util.SupportedMediaBook
	}
	// Unsupported...
	slog.Warn("GetMediaType: Requested, but unsupported type encountered.",
		"type", t.Type)
	return ""
}

type MediaIDs struct {
	// The internal ID
	// Watcharr uint

	// For tmdb data
	TMDB     int    `json:"tmdb,omitempty"`
	IMDB     string `json:"imdb,omitempty"`
	Wikidata string `json:"wikidata,omitempty"`
	TVDB     int    `json:"tvdb,omitempty"`

	// For igdb data
	IGDB int `json:"igdb,omitempty"`

	// For book data
	OLID     string `json:"olid,omitempty"`
	OLAuthor string `json:"olAuthor,omitempty"`
}

type MediaGenre struct {
	// ID of the genre on the external database.
	ID uint `json:"id,omitempty"`
	// Name of genre.
	Name string `json:"name,omitempty"`
}

type MediaSeason struct {
	// Season number (doesn't omit empty to keep support for season 0).
	Number int `json:"number"`
	// Season name.
	Name string `json:"name,omitempty"`
	// Season air date.
	ReleaseDate time.Time `json:"releaseDate,omitzero"`
	// Number of episodes in season.
	EpisodeCount int `json:"episodeCount"`
}

type MediaPerson struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// Create Media dto from Watched entity.
func NewMediaFromWatched(w *entity.Watched, watchedDto *WatchedDto) Media {
	var media Media

	if w.Content != nil {
		media = NewMediaFromContent(w.Content)
	} else if w.Game != nil {
		media = NewMediaFromGame(w.Game)
	} else if w.Book != nil {
		media = NewMediaFromBook(w.Book)
	}

	media.Watched = *watchedDto

	return media
}

// Converter for Content (tv/movie) entity to Media
func NewMediaFromContent(c *entity.Content) Media {
	m := Media{
		IDs: MediaIDs{
			TMDB: c.TmdbID,
		},
		Name:          c.Title,
		Summary:       c.Overview,
		ExtPosterPath: c.PosterPath,
		Rating:        uint(c.VoteAverage),
		RatingCount:   uint(c.VoteCount),
		Runtime:       uint(c.Runtime),
	}
	switch c.Type {
	case entity.MOVIE:
		m.Type = MediaTypeTMDBMovie
	case entity.SHOW:
		m.Type = MediaTypeTMDBShow
	}
	if c.ReleaseDate != nil {
		m.ReleaseDate = *c.ReleaseDate
	}
	return m
}

// Converter for Game entity to Media
func NewMediaFromGame(c *entity.Game) Media {
	var genres []MediaGenre
	for genre := range strings.SplitSeq(c.Genres, "|") {
		genres = append(genres, MediaGenre{Name: genre})
	}
	m := Media{
		IDs: MediaIDs{
			IGDB: c.IgdbID,
		},
		Type:          MediaTypeIGDBGame,
		Name:          c.Name,
		Summary:       c.Summary,
		Poster:        c.Poster,
		ExtPosterPath: c.CoverID,
		Rating:        uint(c.Rating),
		RatingCount:   uint(c.RatingCount),
		Genres:        genres,
	}
	if c.ReleaseDate != nil {
		m.ReleaseDate = *c.ReleaseDate
	}
	return m
}

// Converter for Book entity to Media
func NewMediaFromBook(c *entity.Book) Media {
	m := Media{
		IDs: MediaIDs{
			OLID: c.OLID,
		},
		Type:          MediaTypeOLBook,
		Name:          c.Title,
		Summary:       c.Storyline,
		Poster:        c.Cover,
		ExtPosterPath: c.OLID,
		Rating:        uint(c.RatingAverage),
		RatingCount:   uint(c.RatingCount),
	}
	if c.ReleaseDate != nil {
		m.ReleaseDate = *c.ReleaseDate
	}

	if c.AuthorIDs != "" && c.AuthorNames != "" {
		// append list of book authors
		idsSplit := strings.Split(c.AuthorIDs, "|")
		namesSplit := strings.Split(c.AuthorNames, "|")
		for i := range idsSplit {
			m.Authors = append(m.Authors, MediaPerson{
				ID:   idsSplit[i],
				Name: namesSplit[i],
			})
		}
	}

	return m
}
