package tmdb

import (
	"log/slog"
)

// Getting only region needed from api is not a feature yet
// https://trello.com/c/75tR4cpF/106-add-watch-provider-region-filtering
// When it is, this can be removed for that instead.
func transformProviders(c *any, country string) WatchProviders {
	slog.Debug("transformProviders called", "country", country)
	resp := WatchProviders{}

	cmap, ok := (*c).(map[string]any)
	if !ok {
		slog.Error("transformProviders: Assertion failed")
		return WatchProviders{}
	}

	rmap, ok := cmap["results"].(map[string]any)
	if !ok {
		slog.Warn("transformProviders: Couldn't find results property..")
		return WatchProviders{}
	}

	val, ok := rmap[country]
	if !ok {
		slog.Warn("transformProviders: Couldn't find country..",
			"country", country)
		return WatchProviders{}
	}
	slog.Debug("transformProviders: Found country..", "obj", val)

	rvmap, ok := val.(map[string]any)
	if !ok {
		slog.Warn("transformProviders: Couldn't assert country obj")
		return WatchProviders{}
	}

	// Turning any into a type safe object we can use later.
	// Here we are just getting the flatrate items and manually
	// mapping them to a WatchProvider struct.
	resp.Flatrate = transformProvidersType("flatrate", rvmap, resp.Flatrate)
	resp.Free = transformProvidersType("free", rvmap, resp.Free)

	tmdbLink, ok := rvmap["link"].(string)
	if ok {
		resp.Link = tmdbLink
	}

	return resp
}

// Transform the type of provider requested.
func transformProvidersType(
	ptype string,
	rvmap map[string]any,
	providers []WatchProvider,
) []WatchProvider {
	tm, ok := rvmap[ptype].([]any)
	if !ok {
		slog.Warn("transformProvidersType: Assertion failed")
		return providers
	}
	for i := range tm {
		v2, ok := tm[i].(map[string]any)
		if !ok {
			continue
		}
		providerName, ok := v2["provider_name"].(string)
		if !ok {
			continue
		}
		providers = append(providers,
			WatchProvider{
				ProviderName: providerName,
			})
	}
	return providers
}
