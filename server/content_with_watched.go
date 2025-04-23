// All the functions that help us turn a TMDB response struct
// into one that will also include Watched data.
// This process is very verbose. As far as I am aware, golangs
// generics are not mature (powerful) enough to support us doing
// this all with one function.
//
// TODO When possible look at turning all these funcs into one that
// is reuable for any tmdb search response type.
//
// Each function will basically perform these simple steps:
// 1. Repackage tmdb response so we can add Watched data to
// 2. Get all watched data for the tmdb results
// 4. Add any watched data to our new *WithWatched struct

package main

import (
	"log/slog"

	"gorm.io/gorm"
)

func searchContentAddWatched(
	db *gorm.DB,
	userId uint,
	content TMDBSearchMultiResponse,
) TMDBSearchMultiResponseWithWatched {
	withWatchedResp := TMDBSearchMultiResponseWithWatched{}
	withWatchedResp.TMDBSearchResponse.TMDBPageFields = content.TMDBPageFields
	contentIdAndTypePairs := [][]any{}
	for _, v := range content.Results {
		withWatchedResp.Results = append(withWatchedResp.Results, TMDBSearchMultiResultsWithWatched{
			TMDBSearchMultiResults: v,
		})
		contentIdAndTypePairs = append(contentIdAndTypePairs, []any{
			v.ID,
			ContentType(v.MediaType),
		})
	}
	if ws, err := getWatchedItemsByTmdbIds(db, userId, contentIdAndTypePairs); err == nil {
		for _, v := range ws {
			for i, vv := range withWatchedResp.Results {
				if vv.ID == v.Content.TmdbID && vv.MediaType == string(v.Content.Type) {
					withWatchedResp.Results[i].WatchedAddedToContent.Watched = &v
				}
			}
		}
	} else {
		// TODO Set 'FailedToGetWatched' to `true` for the whole response obj when supported in structs
		slog.Error("Getting watched items by tmdbIds failed!")
	}
	return withWatchedResp
}

func searchMoviesAddWatched(
	db *gorm.DB,
	userId uint,
	content TMDBSearchMoviesResponse,
) TMDBSearchMoviesResponseWithWatched {
	withWatchedResp := TMDBSearchMoviesResponseWithWatched{}
	withWatchedResp.TMDBSearchResponse.TMDBPageFields = content.TMDBPageFields
	contentIdAndTypePairs := [][]any{}
	for _, v := range content.Results {
		withWatchedResp.Results = append(withWatchedResp.Results, TMDBSearchMovieResultWithWatched{
			TMDBSearchMovieResult: v,
		})
		contentIdAndTypePairs = append(contentIdAndTypePairs, []any{
			v.ID,
			ContentType(v.MediaType),
		})
	}
	if ws, err := getWatchedItemsByTmdbIds(db, userId, contentIdAndTypePairs); err == nil {
		for _, v := range ws {
			for i, vv := range withWatchedResp.Results {
				if vv.ID == v.Content.TmdbID && vv.MediaType == string(v.Content.Type) {
					withWatchedResp.Results[i].WatchedAddedToContent.Watched = &v
				}
			}
		}
	} else {
		// TODO Set 'FailedToGetWatched' to `true` for the whole response obj when supported in structs
		slog.Error("Getting watched items by tmdbIds failed!")
	}
	return withWatchedResp
}

func searchTvAddWatched(
	db *gorm.DB,
	userId uint,
	content TMDBSearchShowsResponse,
) TMDBSearchShowsResponseWithWatched {
	withWatchedResp := TMDBSearchShowsResponseWithWatched{}
	withWatchedResp.TMDBSearchResponse.TMDBPageFields = content.TMDBPageFields
	contentIdAndTypePairs := [][]any{}
	for _, v := range content.Results {
		withWatchedResp.Results = append(withWatchedResp.Results, TMDBSearchShowsResultWithWatched{
			TMDBSearchShowsResult: v,
		})
		contentIdAndTypePairs = append(contentIdAndTypePairs, []any{
			v.ID,
			ContentType(v.MediaType),
		})
	}
	if ws, err := getWatchedItemsByTmdbIds(db, userId, contentIdAndTypePairs); err == nil {
		for _, v := range ws {
			for i, vv := range withWatchedResp.Results {
				if vv.ID == v.Content.TmdbID && vv.MediaType == string(v.Content.Type) {
					withWatchedResp.Results[i].WatchedAddedToContent.Watched = &v
				}
			}
		}
	} else {
		// TODO Set 'FailedToGetWatched' to `true` for the whole response obj when supported in structs
		slog.Error("Getting watched items by tmdbIds failed!")
	}
	return withWatchedResp
}

func tvDetailsAddWatched(
	db *gorm.DB,
	userId uint,
	content TMDBShowDetails,
) TMDBShowDetailsWithWatched {
	withWatchedResp := TMDBShowDetailsWithWatched{}
	withWatchedResp.TMDBShowDetailsBase = content.TMDBShowDetailsBase
	// Append watched list entry if exists
	if watchedEntry, err := getWatchedItemByTmdbId(db, userId, uint(content.ID), SHOW); err != nil {
		if err != gorm.ErrRecordNotFound {
			withWatchedResp.FailedToGetWatched = true
		}
	} else {
		withWatchedResp.Watched = &watchedEntry
	}
	// Add similar content with any watched entries
	similarContentIdAndTypePairs := [][]any{}
	for _, v := range content.Similar.Results {
		withWatchedResp.Similar.Results = append(withWatchedResp.Similar.Results, TMDBShowSimilarResultWithWatched{
			TMDBShowSimilarResult: v,
		})
		similarContentIdAndTypePairs = append(similarContentIdAndTypePairs, []any{
			v.ID,
			SHOW,
		})
	}
	if ws, err := getWatchedItemsByTmdbIds(db, userId, similarContentIdAndTypePairs); err == nil {
		for _, v := range ws {
			for i, vv := range withWatchedResp.Similar.Results {
				if vv.ID == v.Content.TmdbID && string(SHOW) == string(v.Content.Type) {
					withWatchedResp.Similar.Results[i].WatchedAddedToContent.Watched = &v
				}
			}
		}
	} else {
		// TODO Set 'FailedToGetWatched' to `true` for the whole response obj when supported in structs
		slog.Error("Getting watched items by tmdbIds failed!")
	}
	return withWatchedResp
}

func allTrendingAddWatched(
	db *gorm.DB,
	userId uint,
	content TMDBTrendingAll,
) TMDBTrendingAllWithWatched {
	withWatchedResp := TMDBTrendingAllWithWatched{}
	contentIdAndTypePairs := [][]any{}
	for _, v := range content.Results {
		withWatchedResp.Results = append(withWatchedResp.Results, TMDBTrendingAllResultWithWatched{
			TMDBTrendingAllResult: v,
		})
		contentIdAndTypePairs = append(contentIdAndTypePairs, []any{
			v.ID,
			ContentType(v.MediaType),
		})
	}
	if ws, err := getWatchedItemsByTmdbIds(db, userId, contentIdAndTypePairs); err == nil {
		for _, v := range ws {
			for i, vv := range withWatchedResp.Results {
				if vv.ID == v.Content.TmdbID && vv.MediaType == string(v.Content.Type) {
					withWatchedResp.Results[i].WatchedAddedToContent.Watched = &v
				}
			}
		}
	} else {
		// TODO Set 'FailedToGetWatched' to `true` for the whole response obj when supported in structs
		slog.Error("Getting watched items by tmdbIds failed!")
	}
	return withWatchedResp
}

func discoverTvAddWatched(
	db *gorm.DB,
	userId uint,
	content TMDBDiscoverShows,
) TMDBDiscoverShowsWithWatched {
	withWatchedResp := TMDBDiscoverShowsWithWatched{}
	contentIdAndTypePairs := [][]any{}
	for _, v := range content.Results {
		withWatchedResp.Results = append(withWatchedResp.Results, TMDBDiscoverShowsResultWithWatched{
			TMDBDiscoverShowsResult: v,
		})
		contentIdAndTypePairs = append(contentIdAndTypePairs, []any{
			v.ID,
			SHOW,
		})
	}
	if ws, err := getWatchedItemsByTmdbIds(db, userId, contentIdAndTypePairs); err == nil {
		for _, v := range ws {
			for i, vv := range withWatchedResp.Results {
				if vv.ID == v.Content.TmdbID && SHOW == v.Content.Type {
					withWatchedResp.Results[i].WatchedAddedToContent.Watched = &v
				}
			}
		}
	} else {
		// TODO Set 'FailedToGetWatched' to `true` for the whole response obj when supported in structs
		slog.Error("Getting watched items by tmdbIds failed!")
	}
	return withWatchedResp
}

func upcomingTvAddWatched(
	db *gorm.DB,
	userId uint,
	content TMDBUpcomingShows,
) TMDBUpcomingShowsWithWatched {
	withWatchedResp := TMDBUpcomingShowsWithWatched{}
	contentIdAndTypePairs := [][]any{}
	for _, v := range content.Results {
		withWatchedResp.Results = append(withWatchedResp.Results, TMDBUpcomingShowsResultWithWatched{
			TMDBUpcomingShowsResult: v,
		})
		contentIdAndTypePairs = append(contentIdAndTypePairs, []any{
			v.ID,
			SHOW,
		})
	}
	if ws, err := getWatchedItemsByTmdbIds(db, userId, contentIdAndTypePairs); err == nil {
		for _, v := range ws {
			for i, vv := range withWatchedResp.Results {
				if vv.ID == v.Content.TmdbID && SHOW == v.Content.Type {
					withWatchedResp.Results[i].WatchedAddedToContent.Watched = &v
				}
			}
		}
	} else {
		// TODO Set 'FailedToGetWatched' to `true` for the whole response obj when supported in structs
		slog.Error("Getting watched items by tmdbIds failed!")
	}
	return withWatchedResp
}

func discoverMoviesAddWatched(
	db *gorm.DB,
	userId uint,
	content TMDBDiscoverMovies,
) TMDBDiscoverMoviesWithWatched {
	withWatchedResp := TMDBDiscoverMoviesWithWatched{}
	contentIdAndTypePairs := [][]any{}
	for _, v := range content.Results {
		withWatchedResp.Results = append(withWatchedResp.Results, TMDBDiscoverMoviesResultWithWatched{
			TMDBDiscoverMoviesResult: v,
		})
		contentIdAndTypePairs = append(contentIdAndTypePairs, []any{
			v.ID,
			MOVIE,
		})
	}
	if ws, err := getWatchedItemsByTmdbIds(db, userId, contentIdAndTypePairs); err == nil {
		for _, v := range ws {
			for i, vv := range withWatchedResp.Results {
				if vv.ID == v.Content.TmdbID && MOVIE == v.Content.Type {
					withWatchedResp.Results[i].WatchedAddedToContent.Watched = &v
				}
			}
		}
	} else {
		// TODO Set 'FailedToGetWatched' to `true` for the whole response obj when supported in structs
		slog.Error("Getting watched items by tmdbIds failed!")
	}
	return withWatchedResp
}

func upcomingMoviesAddWatched(
	db *gorm.DB,
	userId uint,
	content TMDBUpcomingMovies,
) TMDBUpcomingMoviesWithWatched {
	withWatchedResp := TMDBUpcomingMoviesWithWatched{}
	contentIdAndTypePairs := [][]any{}
	for _, v := range content.Results {
		withWatchedResp.Results = append(withWatchedResp.Results, TMDBUpcomingMoviesResultWithWatched{
			TMDBUpcomingMoviesResult: v,
		})
		contentIdAndTypePairs = append(contentIdAndTypePairs, []any{
			v.ID,
			MOVIE,
		})
	}
	if ws, err := getWatchedItemsByTmdbIds(db, userId, contentIdAndTypePairs); err == nil {
		for _, v := range ws {
			for i, vv := range withWatchedResp.Results {
				if vv.ID == v.Content.TmdbID && MOVIE == v.Content.Type {
					withWatchedResp.Results[i].WatchedAddedToContent.Watched = &v
				}
			}
		}
	} else {
		// TODO Set 'FailedToGetWatched' to `true` for the whole response obj when supported in structs
		slog.Error("Getting watched items by tmdbIds failed!")
	}
	return withWatchedResp
}
