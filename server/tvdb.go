package main

import (
    "bytes"
    "encoding/json"
    "errors"
    "io"
    "log/slog"
    "net/http"
    "strconv"
    "sync"
    "time"
)


const tvdbBaseURL = "https://api4.thetvdb.com/v4"

var (
    tvdbToken     string
    tvdbTokenExp  time.Time
    tvdbTokenMu   sync.Mutex
)

type tvdbLoginResp struct {
    Data struct {
        Token string `json:"token"`
        Expires string `json:"expires"`
    } `json:"data"`
}

// getTVDBToken ensures we have a valid bearer token (token lasts 30days).
func getTVDBToken() (string, error) {
    tvdbTokenMu.Lock()
    defer tvdbTokenMu.Unlock()

    if tvdbToken != "" && time.Now().Before(tvdbTokenExp.Add(-time.Minute)) {
        return tvdbToken, nil
    }

    	key := Config.TVDB_KEY
	if key == "" {
		slog.Error("TVDB_KEY not set in config; requests will fail")
		return "", errors.New("TVDB_KEY not configured")
	}
	body, _ := json.Marshal(map[string]string{"apikey": key})
    req, _ := http.NewRequest("POST", tvdbBaseURL+"/login", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        return "", errors.New("tvdb login failed: " + resp.Status)
    }
    var lr tvdbLoginResp
    if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
        return "", err
    }
    tvdbToken = lr.Data.Token
    tvdbTokenExp = time.Now().Add(23 * time.Hour) // safe margin
    return tvdbToken, nil
}

func tvdbRequest(endpoint string, v interface{}) error {
    for attempt := 0; attempt < 2; attempt++ {
        token, err := getTVDBToken()
        if err != nil {
            slog.Error("tvdbRequest: token err", "err", err)
            return err
        }

        req, _ := http.NewRequest("GET", tvdbBaseURL+endpoint, nil)
        req.Header.Set("Authorization", "Bearer "+token)
        resp, err := http.DefaultClient.Do(req)
        if err != nil {
            return err
        }

        // Handle auth failures – invalidate cached token once and retry, i believe tvdb uses the status code 401/403 to indicate auth failure, i haven't waited a month to confirm this though
        // if users report issues with tvdb auth expiries, we can make simple changes to tweak this.
        if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
            resp.Body.Close()
            tvdbTokenMu.Lock()
            tvdbToken = ""
            tvdbTokenExp = time.Time{}
            tvdbTokenMu.Unlock()
            if attempt == 0 {
                // retry once with fresh token
                continue
            }
        }

        if resp.StatusCode != http.StatusOK {
            body, _ := io.ReadAll(resp.Body)
            resp.Body.Close()
            return errors.New("tvdb api: " + resp.Status + " " + string(body))
        }
        err = json.NewDecoder(resp.Body).Decode(&v)
        resp.Body.Close()
        return err
    }
    return errors.New("tvdb api: exhausted retries")
}

// getTVDBEpisodeInfo fetches /episodes/{id}/extended and returns the TVDB
// series id and absolute episode number.
func getTVDBEpisodeInfo(epId int) (seriesId int, absoluteNumber int, err error) {
    var resp struct {
        Data struct {
            SeriesID       int `json:"seriesId"`
            AbsoluteNumber int `json:"absoluteNumber"`
        } `json:"data"`
    }
    if err = tvdbRequest("/episodes/"+strconv.Itoa(epId)+"/extended", &resp); err != nil {
        return 0, 0, err
    }
    return resp.Data.SeriesID, resp.Data.AbsoluteNumber, nil
} 
// The v4 episode payload now includes the seriesId directly on the root data
// object, so we extract it from there instead of the nested series struct.

func getTVDBSeriesIDFromEpisode(epId int) (int, error) {
    var resp struct {
        Data struct {
            SeriesID int `json:"seriesId"`
        } `json:"data"`
    }
    if err := tvdbRequest("/episodes/"+strconv.Itoa(epId), &resp); err != nil {
        return 0, err
    }
    return resp.Data.SeriesID, nil
}

