package main

import "golang.org/x/time/rate"

// newRateLimiterFromConfig returns a limiter using GLOBAL_RATE_LIMIT_RPS from
// server config with the provided default fallback of 10 req/s
func newRateLimiterFromConfig(defaultR int) *rate.Limiter {
    r := defaultR
    if Config.GLOBAL_RATE_LIMIT_RPS > 0 {
        r = Config.GLOBAL_RATE_LIMIT_RPS
    }
    return rate.NewLimiter(rate.Limit(r), r)
}
