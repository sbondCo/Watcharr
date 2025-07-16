package main

import (
    "context"
    "net/http"
    "strconv"
    "log/slog"

    "github.com/gin-gonic/gin"
)

// addWatchedShowRoute registers the /watched/emby/show helper that can be called
// either by logged-in users (JWT) or by external services supplying ?api_key=.
func (b *BaseRouter) addWatchedShowRoute() {
    b.rg.POST("/watched/emby/show", AuthOrAPIKey(b.db), func(c *gin.Context) {
        userId := c.MustGet("userId").(uint)
        type embyWebhook struct {
            Event string `json:"Event"`
            Item struct {
                ProviderIds struct {
                    Tvdb string `json:"Tvdb"`
                } `json:"ProviderIds"`
                IndexNumber       int `json:"IndexNumber"`
                ParentIndexNumber int `json:"ParentIndexNumber"`
            } `json:"Item"`
            PlaybackInfo struct {
                PlayedToCompletion bool `json:"PlayedToCompletion"`
            } `json:"PlaybackInfo"`
        }
        var payload embyWebhook
        if err := c.ShouldBindJSON(&payload); err != nil {
            c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
            return
        }
        // Accept only completed playback or manual markplayed
        if !payload.PlaybackInfo.PlayedToCompletion && payload.Event != "item.markplayed" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "not a completed playback or markplayed event"})
            return
        }
        // Rate-limit TMDB-intensive processing
        if err := embyLimiter.Wait(context.Background()); err != nil {
            c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{Error: "rate limiter error"})
            return
        }

        slog.Debug("emby webhook parsed", "tvdb", payload.Item.ProviderIds.Tvdb, "season", payload.Item.ParentIndexNumber, "ep", payload.Item.IndexNumber)
        tvdbIdStr := payload.Item.ProviderIds.Tvdb
        tvdbId, _ := strconv.Atoi(tvdbIdStr)

                        // Map TVDB episode id to TMDB S/E numbering via the new 3-call algorithm.
            _, seasonNum, episodeNum, err := MapTVDBEpisodeToTMDB(tvdbId)
            if err != nil {
                slog.Warn("episode mapping failed", "err", err)
                // Fallback: keep original season/episode from payload to avoid drop.
                seasonNum = payload.Item.ParentIndexNumber
                episodeNum = payload.Item.IndexNumber
                if episodeNum == 0 {
                    episodeNum = 1
                }
            }

            

            ar := WatchedShowAddRequest{
                TvdbID:        tvdbId,
                SeasonNumber:  seasonNum,
                EpisodeNumber: episodeNum,
                Status:        FINISHED,
                Rating:        func() int8 {
                    if Config.EMBY_DEFAULT_RATING != 0 {
                        return int8(Config.EMBY_DEFAULT_RATING)
                    }
                    return 5
                }(),
            }
            w, epResp, err := addWatchedShowAndEpisode(b.db, userId, ar)
            if err != nil {
                c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
                return
            }
            c.JSON(http.StatusOK, gin.H{"watched": w, "episodeResult": epResp})
            return
    })
}
