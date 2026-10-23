package arr

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Verifies the Arr client does not forward `apikey` param across a cross-host
// redirect but preserves it on a same-host redirect.
func TestArr_StripsTokenOnCrossHostRedirect(t *testing.T) {
	var token = "i_am_a_secret"

	t.Run("cross-host strips token", func(t *testing.T) {
		var finalToken string
		final := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			finalToken = r.URL.Query().Get("apikey")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}))
		defer final.Close()
		redir := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, final.URL+r.URL.Path+"?"+r.URL.Query().Encode(), http.StatusFound)
		}))
		defer redir.Close()

		sonarr := New(SONARR, &redir.URL, &token)
		sonarr.GetQualityProfiles()

		if finalToken != "" {
			t.Fatalf("apikey forwarded cross-host = %q, want empty", finalToken)
		}
	})

	t.Run("same-host keeps token", func(t *testing.T) {
		var finalToken string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("r") == "1" {
				finalToken = r.URL.Query().Get("apikey")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{}`))
				return
			}
			q := r.URL.Query()
			q.Add("r", "1")
			r.URL.RawQuery = q.Encode()
			http.Redirect(w, r, r.URL.String(), http.StatusFound)
		}))
		defer srv.Close()

		sonarr := New(SONARR, &srv.URL, &token)
		sonarr.GetQualityProfiles()
		if finalToken != token {
			t.Fatalf("apikey on same host = %q, want %q", finalToken, token)
		}
	})
}
