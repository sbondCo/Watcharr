package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/internal/testutil"
	"github.com/sbondCo/Watcharr/router"
)

func TestLoginJellyfin_SuccessPasswordless(t *testing.T) {
	testutil.SetupLogging()
	db := testutil.SetupDB(t)

	var receivedAuth JellyfinAuth
	jfServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Users/AuthenticateByName" {
			http.NotFound(w, r)
			return
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &receivedAuth)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"User": {
				"Id": "jf-user-123",
				"Name": "jean"
			},
			"AccessToken": "jf-access-token-456"
		}`))
	}))
	defer jfServer.Close()

	cfg := &config.ServerConfig{
		JELLYFIN_HOST:   jfServer.URL,
		JWT_SECRET:      "test-secret",
		DEFAULT_COUNTRY: "US",
	}
	svc := NewService(db, cfg, nil)

	// Attempt login with empty password string
	resp, err := svc.LoginJellyfin(&JellyfinLoginRequest{
		Username: "jean",
		Password: "",
	})
	if err != nil {
		t.Fatalf("expected successful login, got err: %v", err)
	}

	if resp.Token == "" {
		t.Fatalf("expected non-empty JWT token in response")
	}

	// Verify the payload sent to Jellyfin passed empty password
	if receivedAuth.Username != "jean" || receivedAuth.Pw != "" {
		t.Errorf("expected jellyfin payload {Username: 'jean', Pw: ''}, got %+v", receivedAuth)
	}

	// Verify user persisted in database as JELLYFIN_USER
	var dbUser entity.User
	if err := db.Where("third_party_id = ? AND type = ?", "jf-user-123", entity.JELLYFIN_USER).Take(&dbUser).Error; err != nil {
		t.Fatalf("failed to query created jellyfin user: %v", err)
	}
	if dbUser.Username != "jean" {
		t.Errorf("expected username 'jean', got %q", dbUser.Username)
	}
	if dbUser.ThirdPartyAuth != "jf-access-token-456" {
		t.Errorf("expected access token 'jf-access-token-456', got %q", dbUser.ThirdPartyAuth)
	}
}

func TestLoginJellyfin_UpstreamFailure(t *testing.T) {
	testutil.SetupLogging()
	db := testutil.SetupDB(t)

	jfServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid username or password"}`))
	}))
	defer jfServer.Close()

	cfg := &config.ServerConfig{
		JELLYFIN_HOST: jfServer.URL,
		JWT_SECRET:    "test-secret",
	}
	svc := NewService(db, cfg, nil)

	_, err := svc.LoginJellyfin(&JellyfinLoginRequest{
		Username: "jean",
		Password: "",
	})
	if err == nil {
		t.Fatalf("expected error on 401 from jellyfin, got nil")
	}
	if err.Error() != "incorrect details" {
		t.Errorf("expected error 'incorrect details', got %q", err.Error())
	}
}

func TestLoginJellyfin_RouterEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testutil.SetupLogging()
	db := testutil.SetupDB(t)

	jfServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"User": {"Id": "jf-id-99", "Name": "bob"},
			"AccessToken": "jf-token-99"
		}`))
	}))
	defer jfServer.Close()

	cfg := &config.ServerConfig{
		JELLYFIN_HOST:   jfServer.URL,
		JWT_SECRET:      "test-secret",
		DEFAULT_COUNTRY: "US",
	}
	svc := NewService(db, cfg, nil)
	trustedHeaderSvc := NewTrustedHeaderService(db, cfg, svc)

	engine := gin.New()
	br := &router.BaseRouter{
		DB:     db,
		Cfg:    cfg,
		Router: &engine.RouterGroup,
	}
	authRouter := NewRouter(br, svc, trustedHeaderSvc)
	authRouter.AddRoutes()

	// 1. Passwordless Jellyfin login succeeds (HTTP 200)
	{
		body := bytes.NewBufferString(`{"username": "bob", "password": ""}`)
		req, _ := http.NewRequest(http.MethodPost, "/auth/jellyfin", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200 for passwordless jellyfin login, got %d: %s", w.Code, w.Body.String())
		}
	}

	// 2. Jellyfin login with omitted password succeeds (HTTP 200)
	{
		body := bytes.NewBufferString(`{"username": "bob"}`)
		req, _ := http.NewRequest(http.MethodPost, "/auth/jellyfin", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200 for omitted password jellyfin login, got %d: %s", w.Code, w.Body.String())
		}
	}

	// 3. Jellyfin login missing username fails (HTTP 400)
	{
		body := bytes.NewBufferString(`{"password": ""}`)
		req, _ := http.NewRequest(http.MethodPost, "/auth/jellyfin", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400 for jellyfin login without username, got %d", w.Code)
		}
	}

	// 4. Regression test: native Watcharr login STILL requires password (HTTP 400)
	{
		body := bytes.NewBufferString(`{"username": "bob", "password": ""}`)
		req, _ := http.NewRequest(http.MethodPost, "/auth/", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400 for native login with empty password, got %d", w.Code)
		}
	}
}
