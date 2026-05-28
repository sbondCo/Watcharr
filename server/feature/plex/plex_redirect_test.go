package plex

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sbondCo/Watcharr/config"
)

// TestGetPlexLibraries_StripsTokenOnCrossHostRedirect verifies the Plex client
// does not forward X-Plex-Token across a cross-host redirect but preserves it
// on a same-host redirect.
func TestGetPlexLibraries_StripsTokenOnCrossHostRedirect(t *testing.T) {
	const token = "secret-plex-token"

	t.Run("cross-host strips token", func(t *testing.T) {
		var finalToken string
		final := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			finalToken = r.Header.Get("X-Plex-Token")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"MediaContainer":{}}`))
		}))
		defer final.Close()
		redir := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, final.URL+r.URL.Path, http.StatusFound)
		}))
		defer redir.Close()

		svc := NewService(&config.ServerConfig{PLEX_HOST: redir.URL})
		_, _ = svc.GetPlexLibraries(token)
		if finalToken != "" {
			t.Fatalf("X-Plex-Token forwarded cross-host = %q, want empty", finalToken)
		}
	})

	t.Run("same-host keeps token", func(t *testing.T) {
		var finalToken string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("r") == "1" {
				finalToken = r.Header.Get("X-Plex-Token")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"MediaContainer":{}}`))
				return
			}
			http.Redirect(w, r, r.URL.Path+"?r=1", http.StatusFound)
		}))
		defer srv.Close()

		svc := NewService(&config.ServerConfig{PLEX_HOST: srv.URL})
		_, _ = svc.GetPlexLibraries(token)
		if finalToken != token {
			t.Fatalf("X-Plex-Token on same host = %q, want %q", finalToken, token)
		}
	})
}
