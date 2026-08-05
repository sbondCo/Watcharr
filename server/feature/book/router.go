package book

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/auth/authmiddleware"
	"github.com/sbondCo/Watcharr/feature/watched/addedtocontent"
	"github.com/sbondCo/Watcharr/router"
	"github.com/sbondCo/Watcharr/util"
)

type WatchedProvider interface {
	UpdateWatchedLastViewedSeason(userId uint, id uint, seasonNum int) error
	GetWatchedItemBySupportedMediaId(userId uint, id uint, t util.SupportedMedia) (entity.Watched, error)
	GetWatchedItemsBySupportedMediaIds(userId uint, c []addedtocontent.IdToTypePair) ([]entity.Watched, error)
}

type Router struct {
	br              *router.BaseRouter
	service         *Service
	watchedProvider WatchedProvider
}

func NewRouter(br *router.BaseRouter, service *Service, watchedProvider WatchedProvider) *Router {
	return &Router{
		br,
		service,
		watchedProvider,
	}
}

func (r *Router) AddRoutes() {
	bookRouter := r.br.Router.Group("/book").Use(authmiddleware.AuthRequired(nil, r.br.Cfg))

	// book details for book page
	bookRouter.GET("/:id", r.GetBookDetails)
	bookRouter.GET("/author/:id", r.GetAuthorDetails)
	bookRouter.GET("/author/:id/credits", r.GetAuthorCredits)
}

func (r *Router) GetBookDetails(c *gin.Context) {
	userId := c.MustGet("userId").(uint)

	if c.Param("id") == "" {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "an id was not provided"})
		return
	}

	content, err := r.service.openLibrary.GetBookDetails(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	contentAsMedia := domain.NewMediaFromBook(&content)
	if err := addedtocontent.AddSingularAndList(
		r.watchedProvider,
		userId,
		contentAsMedia,
		func(w *entity.Watched) {
			contentAsMedia.Watched = domain.NewWatchedDtoForContentPage(w)
		},
		[]*addedtocontent.AddListCall[domain.Media]{
			addedtocontent.NewAddListCall(
				contentAsMedia.Similar,
				func(i int, w *entity.Watched) {
					contentAsMedia.Similar[i].Watched = domain.NewWatchedDtoForLists(w)
				},
			),
		},
	); err != nil {
		slog.Error("GetBookDetails: Failed to add watched to content!", "error", err)
		c.JSON(
			http.StatusInternalServerError,
			router.ErrorResponse{Error: "failed to add watched data to response"},
		)
		return
	}
	c.JSON(http.StatusOK, contentAsMedia)
}

func (r *Router) GetAuthorDetails(c *gin.Context) {
	if c.Param("id") == "" {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "an id was not provided"})
		return
	}

	content, err := r.service.openLibrary.GetAuthorDetails(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	contentAsMedia := content.AsPersonDetailsResponse()
	c.JSON(http.StatusOK, contentAsMedia)
}

func (r *Router) GetAuthorCredits(c *gin.Context) {
	if c.Param("id") == "" {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "an id was not provided"})
		return
	}

	content, err := r.service.openLibrary.GetAuthorCredits(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}

	var booksAsMedia []domain.Media
	for _, book := range content {
		booksAsMedia = append(booksAsMedia, domain.NewMediaFromBook(&book))
	}

	c.JSON(http.StatusOK, domain.PersonCreditsResponse{
		Credits: booksAsMedia,
	})
}
