// This package is our "master" search providing one API
// for access to all of our search endpoints, massively
// simplifying access for any client (aka our web ui).

package search

import (
	"errors"
	"log/slog"
	"net/url"
	"strings"

	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/media/tmdb"
	"github.com/sbondCo/Watcharr/util"
	"gorm.io/gorm"
)

type ContentProvider interface {
	SearchContent(query string, pageNum int) (tmdb.TMDBSearchMultiResponse, error)
	SearchMovies(query string, pageNum int) (tmdb.TMDBSearchMoviesResponse, error)
	SearchTv(query string, pageNum int) (tmdb.TMDBSearchShowsResponse, error)
	SearchPeople(query string, pageNum int) (tmdb.TMDBSearchPeopleResponse, error)
	SearchByExternalId(id string, source string) (tmdb.TMDBSearchMultiResponse, error)
	MovieDetails(id string, country string, rParams map[string]string) (tmdb.TMDBMovieDetails, error)
	TvDetails(id string, country string, rParams map[string]string) (tmdb.TMDBShowDetails, error)
}

type Service struct {
	db              *gorm.DB
	cfg             *config.ServerConfig
	contentProvider ContentProvider
}

func NewService(
	db *gorm.DB,
	cfg *config.ServerConfig,
	contentProvider ContentProvider,
) *Service {
	return &Service{
		db,
		cfg,
		contentProvider,
	}
}

// `Limit` is not supported.
func (s *Service) Search(
	r domain.SearchRequest,
	pp util.PaginationParams,
) (domain.SearchResponse, error) {
	resp := domain.SearchResponse{}

	if r.Query == "" {
		return resp, errors.New("a query is required")
	}

	if s.searchExtProviderById(r.Query, &resp) {
		slog.Debug("Search: External provider id search worked.")
		return resp, nil
	}

	switch r.Type {
	case domain.SearchTypeMulti:
		if err := s.searchMulti(r.Query, pp.Page, &resp); err != nil {
			return resp, errors.New("multi search failed")
		}
	case domain.SearchTypeMovie:
		if err := s.searchMovie(r.Query, pp.Page, &resp); err != nil {
			return resp, errors.New("movie search failed")
		}
	case domain.SearchTypeShow:
		if err := s.searchTv(r.Query, pp.Page, &resp); err != nil {
			return resp, errors.New("tv search failed")
		}
	case domain.SearchTypePerson:
		if err := s.searchPeople(r.Query, pp.Page, &resp); err != nil {
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
	query string,
	page int,
	resp *domain.SearchResponse,
) error {
	// TMDB
	tmdbRes, err := s.contentProvider.SearchContent(query, page)
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
	// IGDB
	if s.cfg.TwitchEnabled() {
		igdbRes, err := s.cfg.TWITCH.Search(query)
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
	query string,
	page int,
	resp *domain.SearchResponse,
) error {
	tmdbRes, err := s.contentProvider.SearchMovies(query, page)
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
	details, err := s.contentProvider.MovieDetails(id, "", map[string]string{})
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

func (s *Service) searchTv(
	query string,
	page int,
	resp *domain.SearchResponse,
) error {
	tmdbRes, err := s.contentProvider.SearchTv(query, page)
	if err != nil {
		slog.Error("searchTv: Failed to search tmdb!", "error", err)
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
	details, err := s.contentProvider.TvDetails(id, "", map[string]string{})
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
	query string,
	page int,
	resp *domain.SearchResponse,
) error {
	tmdbRes, err := s.contentProvider.SearchPeople(query, page)
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
		tmdbRes, err := s.contentProvider.SearchByExternalId(
			providerID,
			provider,
		)
		if err != nil {
			slog.Error("searchExtProviderById: Failed to search tmdb!", "error", err)
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
