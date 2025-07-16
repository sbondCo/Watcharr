package main

// embyLimiter restricts Emby webhook processing using GLOBAL_RATE_LIMIT_RPS (default 10 req/s)
// added because if the user marks an entire show as played that can have like 1k+ episodes, it would be a lot of consecutive requests 
// and if you burst above 40ish requests per second you're likely to get banned/throttled from tvdb/tmdb api.
var embyLimiter = newRateLimiterFromConfig(10)
