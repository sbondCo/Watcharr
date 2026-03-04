package domain

import (
	"github.com/go-playground/validator/v10"
	"github.com/sbondCo/Watcharr/util"
)

type DiscoverFilter string

const (
	// Generic "What's Trending?" (all)
	DiscoverFilterTrending DiscoverFilter = "trending"
	// Popular stuff (basically trending but over months instead of just a day).
	DiscoverFilterPopular DiscoverFilter = "popular"
	// Upcoming content (all).
	DiscoverFilterUpcoming DiscoverFilter = "upcoming"
	// What's streaming (movies/tv).
	DiscoverFilterStreaming DiscoverFilter = "streaming"
	// What's in theatres (movies).
	DiscoverFilterInTheatres DiscoverFilter = "intheatres"
)

type DiscoverRequest struct {
	// The type of content we want to discover.
	// Reusing the SearchType enum here, but if this needs to diverge,
	// then make our own enum in this file.
	Type SearchType `form:"type" binding:"validsearchtype"`
	// A main filter.
	// Not every `Type` of discover will support all Filters (service funcs
	// will error individually based on what they support).
	Filter DiscoverFilter `form:"filter" binding:"validdiscoverfilter"`
}

// Extra data that we provide to the Discover service func.
type DiscoverRequestMeta struct {
	PageParams util.PaginationParams
	Region     string
}

type DiscoverResponse struct {
	util.PaginationResponse[Media, util.None]
}

var ValidDiscoverFilter validator.Func = func(fl validator.FieldLevel) bool {
	st, ok := fl.Field().Interface().(DiscoverFilter)
	if ok {
		switch st {
		case DiscoverFilterTrending,
			DiscoverFilterPopular,
			DiscoverFilterUpcoming,
			DiscoverFilterStreaming,
			DiscoverFilterInTheatres:
			return true
		}
	}
	return false
}
