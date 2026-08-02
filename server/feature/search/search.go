// This package is our "master" search providing one API
// for access to all of our search endpoints, massively
// simplifying access for any client (aka our web ui).

package search

import (
	"errors"
	"log/slog"

	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/media/tmdb"
	"github.com/sbondCo/Watcharr/util"
	"gorm.io/gorm"
)

type ServiceWatchedProvider interface {
	GetWatchedPage(userId uint, pp util.PaginationParams, wr domain.WatchedGetPageRequest, extraProps *domain.WatchedGetPageExtraProps) (util.PaginationResponse[entity.Watched, util.None], error)
}

type Service struct {
	db              *gorm.DB
	cfg             *config.ServerConfig
	tmdb            *tmdb.TMDB
	watchedProvider ServiceWatchedProvider
}

func NewService(
	db *gorm.DB,
	cfg *config.ServerConfig,
	tmdb *tmdb.TMDB,
	watchedProvider ServiceWatchedProvider,
) *Service {
	return &Service{
		db,
		cfg,
		tmdb,
		watchedProvider,
	}
}

// `Limit` is not supported.
func (s *Service) Search(
	r domain.SearchRequest,
	pp util.PaginationParams,
	userId uint,
) (domain.SearchResponse, error) {
	slog.Debug("Search: Running.", "request", r, "user_id", userId)

	resp := domain.SearchResponse{}

	if r.Query == "" {
		return resp, errors.New("a query is required")
	}

	if s.searchExtProviderById(r.Query, &resp) {
		slog.Debug("Search: External provider id search worked.")
		return resp, nil
	}

	if r.PreferMyList {
		if err := s.searchMyList(r.Query, pp, userId, &resp); err != nil {
			// Silently fail allowing the below searches to try execution
			slog.Error("Search: searchMyList failed!", "error", err)
		} else if resp.TotalResults > 0 {
			resp.Meta.FromMyList = true
			return resp, nil
		}
	}

	// Parse query filters.
	query, qfilters := parseQueryFilters(r.Query)

	switch r.Type {
	case domain.SearchTypeMulti:
		sreq := tmdb.SearchUniversalOptions{
			Query: query,
			Page:  pp.Page,
			Adult: qfilters.Adult,
		}
		if err := s.searchMulti(sreq, &resp); err != nil {
			return resp, errors.New("multi search failed")
		}
	case domain.SearchTypeMovie:
		sreq := tmdb.SearchMoviesOptions{
			SearchUniversalOptions: tmdb.SearchUniversalOptions{
				Query: query,
				Page:  pp.Page,
				Adult: qfilters.Adult,
			},
			Year:        qfilters.Year,
			PrimaryYear: qfilters.FirstYear,
		}
		if err := s.searchMovie(sreq, &resp); err != nil {
			return resp, errors.New("movie search failed")
		}
	case domain.SearchTypeShow:
		sreq := tmdb.SearchShowsOptions{
			SearchUniversalOptions: tmdb.SearchUniversalOptions{
				Query: query,
				Page:  pp.Page,
				Adult: qfilters.Adult,
			},
			Year:        qfilters.Year,
			PrimaryYear: qfilters.FirstYear,
		}
		if err := s.searchShow(sreq, &resp); err != nil {
			return resp, errors.New("show search failed")
		}
	case domain.SearchTypePerson:
		sreq := tmdb.SearchUniversalOptions{
			Query: query,
			Page:  pp.Page,
			Adult: qfilters.Adult,
		}
		if err := s.searchPeople(sreq, &resp); err != nil {
			return resp, errors.New("person search failed")
		}
	case domain.SearchTypeGame:
		if err := s.searchGame(r.Query, pp.Page, &resp); err != nil {
			return resp, errors.New("game search failed")
		}
	}
	return resp, nil
}

// TODO if only one of the requests for data fails, we can still return the data?
// TODO but we'd need a way to tell the client that some data failed to get fetched,
// TODO either with a header OR a result added to array of type error
// SearchMulti is TMDB Multi search but with game data added to first page.
func (s *Service) searchMulti(
	req tmdb.SearchUniversalOptions,
	resp *domain.SearchResponse,
) error {
	slog.Debug("searchMulti: Running.", "req", req)
	// TMDB
	tmdbRes, err := s.tmdb.SearchMulti(req)
	if err != nil {
		slog.Error("SearchMulti: Failed to search tmdb!", "error", err)
		return errors.New("content request failed")
	}
	for _, v := range tmdbRes.Results {
		resp.Results = append(
			resp.Results,
			v.AsMedia(),
		)
	}
	// IGDB (we will only get results for the first page)
	if req.Page == 1 && s.cfg.TwitchEnabled() {
		igdbRes, err := s.cfg.TWITCH.Search(req.Query)
		if err != nil {
			slog.Error("SearchMulti: Failed to search igdb!", "error", err)
			return errors.New("content request failed")
		}
		for _, v := range igdbRes {
			resp.Results = append(
				resp.Results,
				v.AsMedia(),
			)
		}
	}
	resp.Page = tmdbRes.Page
	resp.TotalPages = tmdbRes.TotalPages
	resp.TotalResults = int64(tmdbRes.TotalResults)
	return nil
}

func (s *Service) searchMovie(
	req tmdb.SearchMoviesOptions,
	resp *domain.SearchResponse,
) error {
	slog.Debug("searchMovie: Running.", "req", req)
	tmdbRes, err := s.tmdb.SearchMovies(req)
	if err != nil {
		slog.Error("SearchMovie: Failed to search tmdb!", "error", err)
		return errors.New("content request failed")
	}
	for _, v := range tmdbRes.Results {
		resp.Results = append(
			resp.Results,
			v.AsMedia(),
		)
	}
	resp.Page = tmdbRes.Page
	resp.TotalPages = tmdbRes.TotalPages
	resp.TotalResults = int64(tmdbRes.TotalResults)
	return nil
}

func (s *Service) searchMovieById(
	id string,
	resp *domain.SearchResponse,
) error {
	slog.Debug("searchMovieById: Running.", "id", id)
	details, err := s.tmdb.MovieDetails(id, "", map[string]string{})
	if err != nil {
		slog.Error("searchMovieById: Failed to search tmdb!", "error", err)
		return errors.New("content request failed")
	}
	resp.Results = append(
		resp.Results,
		details.AsMedia(),
	)
	resp.Page = 1
	resp.TotalPages = 1
	resp.TotalResults = int64(len(resp.Results))
	return nil
}

func (s *Service) searchShow(
	req tmdb.SearchShowsOptions,
	resp *domain.SearchResponse,
) error {
	slog.Debug("searchTv: Running.", "req", req)
	tmdbRes, err := s.tmdb.SearchShows(req)
	if err != nil {
		slog.Error("searchShow: Failed to search tmdb!", "error", err)
		return errors.New("content request failed")
	}
	for _, v := range tmdbRes.Results {
		resp.Results = append(
			resp.Results,
			v.AsMedia(),
		)
	}
	resp.Page = tmdbRes.Page
	resp.TotalPages = tmdbRes.TotalPages
	resp.TotalResults = int64(tmdbRes.TotalResults)
	return nil
}

func (s *Service) searchTvById(
	id string,
	resp *domain.SearchResponse,
) error {
	slog.Debug("searchTvById: Running.", "id", id)
	details, err := s.tmdb.ShowDetails(id, "", map[string]string{})
	if err != nil {
		slog.Error("searchTvById: Failed to search tmdb!", "error", err)
		return errors.New("content request failed")
	}
	resp.Results = append(
		resp.Results,
		details.AsMedia(),
	)
	resp.Page = 1
	resp.TotalPages = 1
	resp.TotalResults = int64(len(resp.Results))
	return nil
}

func (s *Service) searchPeople(
	req tmdb.SearchUniversalOptions,
	resp *domain.SearchResponse,
) error {
	slog.Debug("searchPeople: Running.", "req", req)
	tmdbRes, err := s.tmdb.SearchPeople(req)
	if err != nil {
		slog.Error("searchPeople: Failed to search tmdb!", "error", err)
		return errors.New("content request failed")
	}
	for _, v := range tmdbRes.Results {
		resp.Results = append(
			resp.Results,
			v.AsMedia(),
		)
	}
	resp.Page = tmdbRes.Page
	resp.TotalPages = tmdbRes.TotalPages
	resp.TotalResults = int64(tmdbRes.TotalResults)
	return nil
}

func (s *Service) searchGame(
	query string,
	page int,
	resp *domain.SearchResponse,
) error {
	slog.Debug("searchGame: Running.", "query", query, "page", page)
	igdbRes, err := s.cfg.TWITCH.Search(query)
	if err != nil {
		slog.Error("searchGame: Failed to search igdb!", "error", err)
		return errors.New("content request failed")
	}
	for _, v := range igdbRes {
		resp.Results = append(
			resp.Results,
			v.AsMedia(),
		)
	}
	resp.Page = 1
	resp.TotalPages = 1
	resp.TotalResults = int64(len(igdbRes))
	return nil
}

func (s *Service) searchGameById(
	id string,
	resp *domain.SearchResponse,
) error {
	slog.Debug("searchGameById: Running.", "id", id)
	igdbRes, err := s.cfg.TWITCH.SearchById(id)
	if err != nil {
		slog.Error("searchGameById: Failed to search igdb!", "error", err)
		return errors.New("content request failed")
	}
	for _, v := range igdbRes {
		resp.Results = append(
			resp.Results,
			v.AsMedia(),
		)
	}
	resp.Page = 1
	resp.TotalPages = 1
	resp.TotalResults = int64(len(igdbRes))
	return nil
}

func (s *Service) searchGameBySlug(
	slug string,
	resp *domain.SearchResponse,
) error {
	slog.Debug("searchGameBySlug: Running.", "slug", slug)
	igdbRes, err := s.cfg.TWITCH.SearchBySlug(slug)
	if err != nil {
		slog.Error("searchGameBySlug: Failed to search igdb!", "error", err)
		return errors.New("content request failed")
	}
	for _, v := range igdbRes {
		resp.Results = append(
			resp.Results,
			v.AsMedia(),
		)
	}
	resp.Page = 1
	resp.TotalPages = 1
	resp.TotalResults = int64(len(igdbRes))
	return nil
}

// This function differs to the other search funcs in this service,
// since it searched our watched table, it already has the watched data,
// so the controller doesn't need to add watched after this returns any results.
func (s *Service) searchMyList(
	query string,
	pp util.PaginationParams,
	userId uint,
	resp *domain.SearchResponse,
) error {
	internalRes, err := s.watchedProvider.GetWatchedPage(
		userId,
		pp,
		domain.WatchedGetPageRequest{},
		&domain.WatchedGetPageExtraProps{Query: query})
	if err != nil {
		slog.Error("searchMyList: Failed to search!", "error", err)
		return errors.New("request failed")
	}
	if internalRes.TotalResults <= 0 {
		slog.Debug("searchMyList: No results found.")
		return nil
	}
	resp.Results = domain.NewWatchedGetPageResponse(internalRes.Results)
	resp.Page = internalRes.Page
	resp.TotalPages = internalRes.TotalPages
	resp.TotalResults = internalRes.TotalResults
	return nil
}
