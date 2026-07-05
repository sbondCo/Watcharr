// This router simply serves the images stored in the server data folder
// under the `img` folder.
// Note: The `img` folder contains user uploaded content (eg profile pictures).

package img

import (
	"path"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/router"
)

type Router struct {
	br *router.BaseRouter
}

func NewRouter(br *router.BaseRouter) *Router {
	return &Router{
		br,
	}
}

func (r *Router) AddRoutes() {
	img := r.br.Router.Group("/img").
		Use(func(c *gin.Context) {
			// The two following headers are preventative since this group
			// (the static route below) hosts user uploaded content, which can
			// potentially include malicious data. We are trying to protect
			// against XSS attacks here by telling the browser to:
			//  - Not sniff content; and
			//  - Not execute JS; and
			//  - treat the content as if it was a separate domain (so if eg
			//    somehow js runs, it won't be in same context as our tokens).
			// RESOURCE: web.dev/articles/securely-hosting-user-data
			c.Header("X-Content-Type-Options", "nosniff")
			c.Header("Content-Security-Policy", "default-src 'none'; sandbox")
			c.Next()
		})

	// Serve up img folder.
	img.Static("/", path.Join(config.DataPath, "img"))
}
